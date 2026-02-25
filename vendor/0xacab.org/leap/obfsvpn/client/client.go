// Package client exposes a proxy that uses obfs4 to communicate with the server,
// with an optional KCP wire transport.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"0xacab.org/leap/obfsvpn/internal/runtimex"
	"0xacab.org/leap/obfsvpn/obfsvpn"
)

type clientState string

const (
	starting clientState = "STARTING"
	running  clientState = "RUNNING"
	stopping clientState = "STOPPING"
	stopped  clientState = "STOPPED"
)

var (
	ErrAlreadyRunning = errors.New("already initialized")
	ErrNotRunning     = errors.New("server not running")
	ErrBadConfig      = errors.New("configuration error")
)

type EventLogger interface {
	Log(state string, message string)
	Error(message string)
}

const (
	dialGiveUpTime = 15 * time.Second
	reconnectTime  = 3 * time.Second
)

type Obfs4Config struct {
	Remote string
	Cert   string
}

type Obfs4Conn struct {
	net.Conn
	config Obfs4Config
}

func NewObfs4Conn(conn net.Conn, config Obfs4Config) *Obfs4Conn {
	return &Obfs4Conn{
		Conn:   conn,
		config: config,
	}
}

func (oc *Obfs4Config) String() string {
	return oc.Remote
}

// Config configures the pluggable transport.
type Config struct {
	ProxyAddr     string             `json:"proxy_addr"`
	HoppingConfig HoppingConfig      `json:"hopping_config"`
	KCPConfig     obfsvpn.KCPConfig  `json:"kcp_config"`
	QUICConfig    obfsvpn.QUICConfig `json:"quic_config"`
	RemoteIP      string             `json:"remote_ip"`
	RemotePort    string             `json:"remote_port"`
	Obfs4Cert     string             `json:"obfs4_cert"`
}

// HoppingConfig configures the hopping transport.
type HoppingConfig struct {
	Enabled       bool     `json:"enabled"`
	Remotes       []string `json:"remotes"`
	Obfs4Certs    []string `json:"obfs4_certs"`
	PortSeed      int64    `json:"port_seed"`
	PortCount     uint     `json:"port_count"`
	MinHopPort    uint     `json:"min_hop_port"`
	MaxHopPort    uint     `json:"max_hop_port"`
	MinHopSeconds uint     `json:"min_hop_seconds"`
	HopJitter     uint     `json:"hop_jitter"`
}

// portHopRange returns the port hop range or an error if the config is invalid.
//
// Note: this method assumes that hopping is enabled and the HoppingConfig is being used.
func (hc *HoppingConfig) portHopRange() (int, error) {
	if hc.MaxHopPort <= hc.MinHopPort {
		return 0, fmt.Errorf("hoppingconfig: MaxHopPort (%d) must be greater than MinHopPort (%d)", hc.MaxHopPort, hc.MinHopPort)
	}
	if hc.MaxHopPort > math.MaxUint16 {
		return 0, fmt.Errorf("hoppingconfig: MaxHopPort (%d) must not be greater than MaxUint16", hc.MaxHopPort)
	}
	if hc.MinHopPort > math.MaxUint16 {
		return 0, fmt.Errorf("hoppingconfig: MinHopPort (%d) must not be greater than MaxUint16", hc.MinHopPort)
	}
	// #nosec G115 - the following operation is safe because of the above check
	v := int(hc.MaxHopPort - hc.MinHopPort)
	return v, nil
}

type Client struct {
	kcpConfig       obfsvpn.KCPConfig
	quicConfig      obfsvpn.QUICConfig
	ProxyAddr       string
	newObfs4Conn    chan Obfs4Conn
	obfs4Conns      []Obfs4Conn
	obfs4Endpoints  []*Obfs4Config
	obfs4Failures   map[string]int32
	EventLogger     EventLogger
	state           clientState
	ctx             context.Context
	mux             sync.Mutex
	stop            context.CancelFunc
	openvpnConn     *net.UDPConn
	openvpnAddr     *net.UDPAddr
	openvpnAddrLock sync.RWMutex
	outLock         sync.Mutex
	hopEnabled      bool
	minHopSeconds   uint
	hopJitter       uint
}

// TODO(bassosimone): refactor sleepSeconds to return a [time.Duration].

// sleepSeconds returns the number of seconds to sleep or an error.
//
// Note: this method assumes that c.hopEnabled is true.
func (c *Client) sleepSeconds() (int, error) {
	if c.hopJitter <= 0 {
		return 0, fmt.Errorf("c.hopJitter (%d) must be positive", c.hopJitter)
	}
	if c.hopJitter > math.MaxInt {
		return 0, fmt.Errorf("c.hopJitter (%d) must not be greater than math.MaxInt", c.hopJitter)
	}
	if c.minHopSeconds > math.MaxInt {
		return 0, fmt.Errorf("c.minHopSeconds (%d) must not be greater than math.MaxInt", c.minHopSeconds)
	}
	if c.hopJitter > math.MaxInt-c.minHopSeconds {
		return 0, fmt.Errorf("c.hopJitter + c.minHopSeconds (%d + %d) must not be greater than math.MaxInt", c.hopJitter, c.minHopSeconds)
	}
	// #nosec G404 G115 -- OK to use weak RNG and to convert c.hopJitter to int
	secs := rand.Intn(int(c.hopJitter)) + int(c.minHopSeconds)
	return secs, nil
}

// TODO(bassosimone): storing a context inside a structure is not considered
// idiomatic Go: https://go.dev/wiki/CodeReviewComments#contexts

// NewClient constructs a [*Client] instance.
func NewClient(ctx context.Context, stop context.CancelFunc, config Config) (*Client, error) {
	obfs4Endpoints, err := generateObfs4Config(config)
	if err != nil {
		return nil, err
	}

	c := &Client{
		ProxyAddr:      config.ProxyAddr,
		hopEnabled:     config.HoppingConfig.Enabled,
		ctx:            ctx,
		hopJitter:      config.HoppingConfig.HopJitter,
		kcpConfig:      config.KCPConfig,
		quicConfig:     config.QUICConfig,
		obfs4Failures:  map[string]int32{},
		minHopSeconds:  config.HoppingConfig.MinHopSeconds,
		newObfs4Conn:   make(chan Obfs4Conn),
		obfs4Endpoints: obfs4Endpoints,
		stop:           stop,
		state:          stopped,
	}

	// Run the check now to avoid panicking in [*Client.hop]
	if c.hopEnabled {
		if _, err := c.sleepSeconds(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// NewFFIClient creates a new client
// This function is exposed to the JNI and since it's not allowed to pass objects that contain slices (other than byte slices) over the JNI
// we have to pass a json formatted string and convert it to a Config struct for further processing
func NewFFIClient(jsonConfig string) (*Client, error) {
	config := Config{}
	err := json.Unmarshal([]byte(jsonConfig), &config)
	if err != nil {
		return nil, err
	}
	ctx, stop := context.WithCancel(context.Background())
	return NewClient(ctx, stop, config)
}

// generateObfs4Config generates the OBFS4 config or returns an error.
func generateObfs4Config(config Config) ([]*Obfs4Config, error) {
	obfsEndpoints := []*Obfs4Config{}

	if config.HoppingConfig.Enabled {
		// Safely compute the portHopRange and bail if the config is wrong
		portHopRange, err := config.HoppingConfig.portHopRange()
		if err != nil {
			return nil, err
		}

		for i, obfs4Remote := range config.HoppingConfig.Remotes {
			// We want a non-crypto RNG so that we can share a seed
			// #nosec G404
			r := rand.New(rand.NewSource(config.HoppingConfig.PortSeed))
			for pi := uint(0); pi < config.HoppingConfig.PortCount; pi++ {
				// #nosec G115 - Intn panics if given a negative value
				portOffset := uint(r.Intn(portHopRange))

				addr := net.JoinHostPort(obfs4Remote, fmt.Sprint(portOffset+config.HoppingConfig.MinHopPort))
				obfsEndpoints = append(obfsEndpoints, &Obfs4Config{
					Cert:   config.HoppingConfig.Obfs4Certs[i],
					Remote: addr,
				})
			}
		}
	} else {
		addr := net.JoinHostPort(config.RemoteIP, config.RemotePort)
		obfsEndpoints = append(obfsEndpoints, &Obfs4Config{
			Cert:   config.Obfs4Cert,
			Remote: addr,
		})
	}

	log.Printf("obfs4 endpoints: %+v", obfsEndpoints)
	return obfsEndpoints, nil
}

func (c *Client) Start() (_ bool, err error) {
	c.mux.Lock()
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1<<16) // 64KB buffer
			stackSize := runtime.Stack(buf, true)
			err = fmt.Errorf("recovered from panic: %v\nStack trace:\n%s", r, buf[:stackSize])
		}

		c.updateState(stopped)
		if err != nil {
			c.mux.Unlock()
		}
	}()

	if c.IsStarted() {
		c.error("Cannot start proxy server, already running")
		err = ErrAlreadyRunning
		return false, err
	}

	if len(c.obfs4Endpoints) == 0 {
		c.error("Cannot start proxy server, no valid endpoints")
		err = ErrBadConfig
		return false, err
	}

	c.updateState(starting)

	obfs4Endpoint := c.obfs4Endpoints[0]
	obfs4Conn, err := c.createObfs4Connection(obfs4Endpoint)
	if err != nil {
		c.error("Could not createObfs4Connection: %v", err)
		return false, fmt.Errorf("could not dial remote: %w", err)
	}
	c.obfs4Conns = []Obfs4Conn{*obfs4Conn}

	c.openvpnConn, err = c.createOpenvpnConnection()
	if err != nil {
		return false, err
	}

	c.updateState(running)

	if c.hopEnabled {
		go c.hop()
	}

	go c.readUDPWriteTCP()

	go c.readTCPWriteUDP()

	c.mux.Unlock()

	<-c.ctx.Done()

	return true, nil
}

func (c *Client) createObfs4Connection(obfs4Endpoint *Obfs4Config) (*Obfs4Conn, error) {
	var err error

	obfs4Dialer, err := obfsvpn.NewDialerFromCert(obfs4Endpoint.Cert)
	if err != nil {
		c.error("Could not get obfs4Dialer from cert: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(c.ctx, dialGiveUpTime)
	defer cancel()

	if c.kcpConfig.Enabled {
		obfs4Dialer.DialFunc = obfsvpn.GetKCPDialer(c.kcpConfig, c.log)
	} else if c.quicConfig.Enabled {
		obfs4Dialer.DialFunc = obfsvpn.GetQUICDialer(ctx, c.log)
	}

	c.log("Dialing remote: %v", obfs4Endpoint.Remote)
	conn, err := obfs4Dialer.DialContext(ctx, "tcp", obfs4Endpoint.Remote)
	if err != nil {
		return nil, fmt.Errorf("error in obfs4Dialer.DialContext: %w", err)
	}
	return NewObfs4Conn(conn, *obfs4Endpoint), nil
}

// updateState sets a new client state, logs it and sends an event to the clients
// EventLogger in case it is available. Always set the state with this function in
// order to ensure integrating clients receive an update state event via FFI.
func (c *Client) updateState(state clientState) {
	c.state = state
	c.log("Update state: %v", state)
}

// pickRandomRemote returns a random remote from the internal array.
// An obvious improvement to this function is to check the number of failures in c.obfs4Failures and avoid
// a given remote if it failed more than a threshold. A consecuence is that
// we'll have to return an unrecoverable error from hop() if there are no
// more usable remotes. If we ever want to get fancy, an even better heuristic
// can be to avoid IPs that have more failures than the average.
func (c *Client) pickRandomEndpoint() *Obfs4Config {
	// #nosec G404
	i := rand.Intn(len(c.obfs4Endpoints))
	endpoint := c.obfs4Endpoints[i]
	// here we could check if the number of failures is ok-ish. we can also do moving averages etc.
	return endpoint
}

func (c *Client) connectObfs4(obfs4Endpoint *Obfs4Config, sleepSeconds int) {
	newObfs4Conn, err := c.createObfs4Connection(obfs4Endpoint)

	if err != nil {
		newRemote := obfs4Endpoint.Remote
		_, ok := c.obfs4Failures[newRemote]
		if ok {
			c.obfs4Failures[newRemote] += 1
		} else {
			c.obfs4Failures[newRemote] = 1
		}
		c.error("Could not dial obfs4 remote: %v (failures: %d)", err, c.obfs4Failures[newRemote])
	}

	if newObfs4Conn == nil {
		c.error("Did not get obfs4: %v ", err)
	} else {
		c.outLock.Lock()
		c.obfs4Conns = append([]Obfs4Conn{*newObfs4Conn}, c.obfs4Conns...)
		c.outLock.Unlock()

		c.newObfs4Conn <- *newObfs4Conn
		c.log("Dialed new remote")

		// If we wait sleepSeconds here to clean up the previous connection, we can guarantee that the
		// connection list will not grow unbounded
		go func() {
			cancelled := cancellableSleep(c.ctx, time.Duration(sleepSeconds)*time.Second)
			if cancelled {
				c.log("Sleep for connection cleanup cancelled. Immediately cleaning up old connections")
			}
			c.cleanupOldConn()
		}()
	}
}

// cancellableSleep simulates a sleep that can be canceled by a context.
func cancellableSleep(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return false // Sleep completed
	case <-ctx.Done():
		return true // Sleep canceled
	}
}

// hop implements the port hopping logic for the client.
//
// Note: this method assumes that c.hopEnabled is true and should only be called
// when hopping is enabled. The caller is responsible for checking c.hopEnabled.
func (c *Client) hop() {
	for {
		// Note that we ensure that c.sleepSeconds does not fail in [NewClient]
		sleepSeconds := runtimex.AssertNotErrorWithReturnOrPanic(c.sleepSeconds())

		c.log("Sleeping %d seconds...", sleepSeconds)
		cancelled := cancellableSleep(c.ctx, time.Duration(sleepSeconds)*time.Second)
		if cancelled {
			c.log("Hop sleep cancelled")
			return
		}

		obfs4Endpoint := c.pickRandomEndpoint()

		host, port, err := net.SplitHostPort(obfs4Endpoint.Remote)
		if err != nil {
			c.error("Could not split obfs4 remote: %v", err)
			continue
		}
		remoteAddrs, err := net.DefaultResolver.LookupHost(c.ctx, host)
		if err != nil {
			c.error("Could not lookup obfs4 remote: %v", err)
			continue
		}

		if len(remoteAddrs) <= 0 {
			c.error("Could not lookup obfs4 remote: %v", err)
			continue
		}

		newRemote := net.JoinHostPort(remoteAddrs[0], port)

		for _, obfs4Conn := range c.obfs4Conns {
			if obfs4Conn.RemoteAddr().String() == newRemote {
				c.log("Not hopping to address already in obfs4Conns list: %v", newRemote)
				continue
			}
		}

		c.log("HOPPING to %+v", newRemote)
		c.connectObfs4(obfs4Endpoint, sleepSeconds)
	}
}

func (c *Client) cleanupOldConn() {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if len(c.obfs4Conns) > 1 {
		c.log("Connections: %v", len(c.obfs4Conns))
		connToClose := c.obfs4Conns[len(c.obfs4Conns)-1]
		c.log("Cleaning up old connection to %v", connToClose.RemoteAddr())

		err := connToClose.Close()
		if err != nil {
			c.log("Error closing obfs4 connection to %v: %v", connToClose.RemoteAddr(), err)
		}

		// Remove the connection from our tracking list
		c.obfs4Conns = c.obfs4Conns[:len(c.obfs4Conns)-1]
	}
}

func (c *Client) readUDPWriteTCP() {
	datagramBuffer := make([]byte, obfsvpn.MaxUDPLen)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		tcpBuffer, newOpenvpnAddr, err := obfsvpn.ReadUDPFrameTCP(c.openvpnConn, datagramBuffer)
		if err != nil {
			c.error("Read err from %v: %v", c.openvpnConn.LocalAddr(), err)
			continue
		}

		if newOpenvpnAddr != c.openvpnAddr {
			c.openvpnAddrLock.Lock()
			c.openvpnAddr = newOpenvpnAddr
			c.openvpnAddrLock.Unlock()
		}

		// Always write to the first connection in our list because it will be most up to date
		conn, err := c.getUsableConnection()
		if err != nil {
			c.error("Cannot get connection: %s", err)
			continue
		}
		_, err = conn.Write(tcpBuffer)
		if err != nil {
			c.error("readUDPWriteTCP: Write err from %v to %v: %v", conn.LocalAddr(), conn.RemoteAddr(), err)
			cancelled := cancellableSleep(c.ctx, reconnectTime)
			if !cancelled {
				conn, err = c.getUsableConnection()
				if err != nil {
					return
				}
				config := conn.config
				c.connectObfs4(&config, 20)
			}
		}
	}
}

func (c *Client) getUsableConnection() (*Obfs4Conn, error) {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if len(c.obfs4Conns) == 0 {
		return nil, errors.New("no usable connection")
	} else {
		return &c.obfs4Conns[0], nil
	}
}

func (c *Client) readTCPWriteUDP() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		fromTCP := make(chan []byte, 2048)

		handleObfs4Conn := func(conn net.Conn) {
			datagramBuffer := make([]byte, obfsvpn.MaxUDPLen)
			lengthBuffer := make([]byte, 2)
			for {
				udpBuffer, err := obfsvpn.ReadTCPFrameUDP(conn, datagramBuffer, lengthBuffer)
				if err != nil {
					c.error("Reading/framing error: %v", err)
					return
				}

				fromTCP <- udpBuffer
			}
		}

		go func() {
			for {
				newObfs4Conn := <-c.newObfs4Conn

				go handleObfs4Conn(newObfs4Conn)
			}
		}()

		go handleObfs4Conn(c.obfs4Conns[0])

		for {
			tcpBytes := <-fromTCP
			c.openvpnAddrLock.RLock()
			_, err := c.openvpnConn.WriteToUDP(tcpBytes, c.openvpnAddr)
			c.openvpnAddrLock.RUnlock()
			if err != nil && !errors.Is(err, net.ErrClosed) {
				c.error("readTCPWriteUDP: Write err from %v to %v: %v", c.openvpnConn.LocalAddr(), c.openvpnConn.RemoteAddr(), err)
				c.openvpnAddrLock.Lock()
				c.openvpnConn.Close()
				c.openvpnConn, err = c.createOpenvpnConnection()
				c.openvpnAddrLock.Unlock()
				if err == nil {
					c.openvpnAddrLock.RLock()
					_, err := c.openvpnConn.WriteToUDP(tcpBytes, c.openvpnAddr)
					c.openvpnAddrLock.RUnlock()
					if err != nil {
						c.error("Failed to resend. %v", err)
					}
				}
			}
		}
	}
}

func (c *Client) createOpenvpnConnection() (*net.UDPConn, error) {
	proxyAddr, err := net.ResolveUDPAddr("udp", c.ProxyAddr)
	if err != nil {
		c.error("cannot resolve UDP addr: %v", err)
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", proxyAddr)
	if err != nil {
		c.error("error accepting udp connection: %v", err)
		return nil, fmt.Errorf("error accepting udp connection: %w", err)
	}
	return udpConn, nil
}

func (c *Client) Stop() (_ bool, err error) {
	c.mux.Lock()
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1<<16) // 64KB buffer
			stackSize := runtime.Stack(buf, true)
			err = fmt.Errorf("recovered from panic: %v\nStack trace:\n%s", r, buf[:stackSize])
		}
		c.mux.Unlock()
	}()

	if !c.IsStarted() {
		return false, ErrNotRunning
	}

	c.stop()
	c.openvpnConn.Close()

	c.updateState(stopped)

	return true, nil
}

func (c *Client) log(format string, a ...interface{}) {
	if c.EventLogger != nil {
		c.EventLogger.Log(string(c.state), fmt.Sprintf(format, a...))
		return
	}
	if format == "" {
		log.Println(a...)
		return
	}
	log.Printf(format+"\n", a...)
}

func (c *Client) error(format string, a ...interface{}) {
	if c.EventLogger != nil {
		c.EventLogger.Error(fmt.Sprintf(format, a...))
		return
	}
	if format == "" {
		log.Println(a...)
		return
	}
	log.Printf(format+"\n", a...)
}

func (c *Client) IsStarted() bool {
	return c.state != stopped
}

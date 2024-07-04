// Package client exposes a proxy that uses obfs4 to communicate with the server,
// with an optional KCP wire transport.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

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
)

type Obfs4Config struct {
	Remote string
	Cert   string
}

func (oc *Obfs4Config) String() string {
	return oc.Remote
}

type Config struct {
	ProxyAddr     string            `json:"proxy_addr"`
	HoppingConfig HoppingConfig     `json:"hopping_config"`
	KCPConfig     obfsvpn.KCPConfig `json:"kcp_config"`
	RemoteIP      string            `json:"remote_ip"`
	RemotePort    string            `json:"remote_port"`
	Obfs4Cert     string            `json:"obfs4_cert"`
}

type HoppingConfig struct {
	Enabled       bool     `json:"enabled"`
	Remotes       []string `json:"remotes"`
	Obfs4Certs    []string `json:"obfs4_certs"`
	PortSeed      int64    `json:"port_seed"`
	PortCount     uint     `json:"port_count"`
	MinHopSeconds uint     `json:"min_hop_seconds"`
	HopJitter     uint     `json:"hop_jitter"`
}

type Client struct {
	kcpConfig       obfsvpn.KCPConfig
	ProxyAddr       string
	newObfs4Conn    chan net.Conn
	obfs4Conns      []net.Conn
	obfs4Endpoints  []*Obfs4Config
	obfs4Dialer     *obfsvpn.Dialer
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

func NewClient(ctx context.Context, stop context.CancelFunc, config Config) *Client {
	obfs4Endpoints := generateObfs4Config(config)
	return &Client{
		ProxyAddr:      config.ProxyAddr,
		hopEnabled:     config.HoppingConfig.Enabled,
		ctx:            ctx,
		hopJitter:      config.HoppingConfig.HopJitter,
		kcpConfig:      config.KCPConfig,
		obfs4Failures:  map[string]int32{},
		minHopSeconds:  config.HoppingConfig.MinHopSeconds,
		newObfs4Conn:   make(chan net.Conn),
		obfs4Endpoints: obfs4Endpoints,
		stop:           stop,
		state:          stopped,
	}
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
	return NewClient(ctx, stop, config), nil
}

func generateObfs4Config(config Config) []*Obfs4Config {
	obfsEndpoints := []*Obfs4Config{}

	if config.HoppingConfig.Enabled {
		for i, obfs4Remote := range config.HoppingConfig.Remotes {
			// We want a non-crypto RNG so that we can share a seed
			// #nosec G404
			r := rand.New(rand.NewSource(config.HoppingConfig.PortSeed))
			for pi := 0; pi < int(config.HoppingConfig.PortCount); pi++ {
				portOffset := r.Intn(obfsvpn.PortHopRange)
				addr := net.JoinHostPort(obfs4Remote, fmt.Sprint(portOffset+obfsvpn.MinHopPort))
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
	return obfsEndpoints
}

func (c *Client) Start() (bool, error) {
	var err error

	c.mux.Lock()

	defer func() {
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

	c.obfs4Dialer, err = obfsvpn.NewDialerFromCert(obfs4Endpoint.Cert)
	if err != nil {
		return false, fmt.Errorf("could not dial obfs4 remote: %w", err)
	}

	if c.kcpConfig.Enabled {
		c.obfs4Dialer.DialFunc = obfsvpn.GetKCPDialer(c.kcpConfig, c.log)
	}

	obfs4Conn, err := c.obfs4Dialer.Dial("tcp", obfs4Endpoint.Remote)
	if err != nil {
		c.error("Could not dial obfs4 remote: %v", err)
		return false, fmt.Errorf("could not dial remote: %w", err)
	}

	c.obfs4Conns = []net.Conn{obfs4Conn}

	c.updateState(running)

	proxyAddr, err := net.ResolveUDPAddr("udp", c.ProxyAddr)
	if err != nil {
		return false, fmt.Errorf("cannot resolve UDP addr: %w", err)
	}

	c.openvpnConn, err = net.ListenUDP("udp", proxyAddr)
	if err != nil {
		return false, fmt.Errorf("error accepting udp connection: %w", err)
	}

	if c.hopEnabled {
		go c.hop()
	}

	go c.readUDPWriteTCP()

	go c.readTCPWriteUDP()

	c.mux.Unlock()

	<-c.ctx.Done()

	return true, nil
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

func (c *Client) hop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// #nosec G404
		sleepSeconds := rand.Intn(int(c.hopJitter)) + int(c.minHopSeconds)
		c.log("Sleeping %d seconds...", sleepSeconds)
		time.Sleep(time.Duration(sleepSeconds) * time.Second)

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

		obfs4Dialer, err := obfsvpn.NewDialerFromCert(obfs4Endpoint.Cert)
		if err != nil {
			c.error("Could not dial obfs4 remote: %v", err)
			return
		}

		if c.kcpConfig.Enabled {
			c.obfs4Dialer.DialFunc = obfsvpn.GetKCPDialer(c.kcpConfig, c.log)
		}

		ctx, cancel := context.WithTimeout(context.Background(), dialGiveUpTime)
		defer cancel()

		c.log("Dialing new remote: %v", newRemote)
		newObfs4Conn, err := obfs4Dialer.DialContext(ctx, "tcp", newRemote)

		if err != nil {
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
			c.obfs4Conns = append([]net.Conn{newObfs4Conn}, c.obfs4Conns...)
			c.outLock.Unlock()

			c.newObfs4Conn <- newObfs4Conn
			c.log("Dialed new remote")

			// If we wait sleepSeconds here to clean up the previous connection, we can guarantee that the
			// connection list will not grow unbounded
			go func() {
				time.Sleep(time.Duration(sleepSeconds) * time.Second)

				c.cleanupOldConn()
			}()
		}
	}
}

func (c *Client) cleanupOldConn() {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if len(c.obfs4Conns) > 1 {
		c.log("Connections: %v", len(c.obfs4Conns))
		connToClose := c.obfs4Conns[len(c.obfs4Conns)-1]
		if connToClose != nil {
			c.log("Cleaning up old connection to %v", connToClose.RemoteAddr())

			err := connToClose.Close()
			if err != nil {
				c.log("Error closing obfs4 connection to %v: %v", connToClose.RemoteAddr(), err)
			}
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

		func() {
			// Always write to the first connection in our list because it will be most up to date
			func() {
				conn, err := c.getUsableConnection()
				if err != nil {
					c.log("Cannot get connection: %s", err)
					return
				}
				_, err = conn.Write(tcpBuffer)
				if err != nil {
					c.log("Write err from %v to %v: %v", conn.LocalAddr(), conn.RemoteAddr(), err)
					return
				}
			}()
		}()
	}
}

func (c *Client) getUsableConnection() (net.Conn, error) {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if len(c.obfs4Conns) == 0 {
		return nil, errors.New("no usable connection")
	} else {
		return c.obfs4Conns[0], nil
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
			if err != nil {
				c.error("Write err from %v to %v: %v", c.openvpnConn.LocalAddr(), c.openvpnConn.RemoteAddr(), err)
				return
			}
		}
	}
}

func (c *Client) Stop() (bool, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

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

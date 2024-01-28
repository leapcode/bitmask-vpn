// Package client exposes a socks5 proxy that uses obfs4 to communicate with the server,
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
	"github.com/xtaci/kcp-go"
)

type clientState string

const (
	starting clientState = "STARTING"
	running  clientState = "RUNNING"
	stopping clientState = "STOPPING"
	stopped  clientState = "STOPPED"
)

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

type HoppingConfig struct {
	KCP           bool     `json:"kcp"`
	ProxyAddr     string   `json:"proxy_addr"`
	Remotes       []string `json:"remotes"`
	Certs         []string `json:"certs"`
	PortSeed      int64    `json:"port_seed"`
	PortCount     uint     `json:"port_count"`
	MinHopSeconds uint     `json:"min_hop_seconds"`
	HopJitter     uint     `json:"hop_jitter"`
}

type HopClient struct {
	kcp             bool
	ProxyAddr       string
	newObfs4Conn    chan net.Conn
	obfs4Conns      []net.Conn
	obfs4Endpoints  []*Obfs4Config
	obfs4Dialer     *obfsvpn.Dialer
	obfs4Failures   map[string]int32
	EventLogger     EventLogger
	state           clientState
	ctx             context.Context
	stop            context.CancelFunc
	openvpnConn     *net.UDPConn
	openvpnAddr     *net.UDPAddr
	openvpnAddrLock sync.RWMutex
	outLock         sync.Mutex
	minHopSeconds   uint
	hopJitter       uint
}

func NewHopClient(ctx context.Context, stop context.CancelFunc, config HoppingConfig) ObfsClient {
	obfs4Endpoints := generateObfs4Config(config.Remotes, config.PortSeed, config.PortCount, config.Certs)
	return &HopClient{
		ProxyAddr: config.ProxyAddr,

		ctx:            ctx,
		hopJitter:      config.HopJitter,
		kcp:            config.KCP,
		obfs4Failures:  map[string]int32{},
		minHopSeconds:  config.MinHopSeconds,
		newObfs4Conn:   make(chan net.Conn),
		obfs4Endpoints: obfs4Endpoints,
		stop:           stop,
		state:          stopped,
	}
}

// NewFFIHopClient creates a new Hopping PT client
// This function is exposed to the JNI and since it's not allowed to pass objects that contain slices (other than byte slices) over the JNI
// we have to pass a json formatted string and convert it to a HoppingConfig struct for further processing
func NewFFIHopClient(hoppingConfig string) (*HopClient, error) {
	config := HoppingConfig{}
	err := json.Unmarshal([]byte(hoppingConfig), &config)
	if err != nil {
		return nil, err
	}
	ctx, stop := context.WithCancel(context.Background())
	return NewHopClient(ctx, stop, config).(*HopClient), nil
}

func generateObfs4Config(remoteIPs []string, portSeed int64, portCount uint, certs []string) []*Obfs4Config {
	obfsEndpoints := []*Obfs4Config{}

	for i, obfs4Remote := range remoteIPs {
		// We want a non-crypto RNG so that we can share a seed
		// #nosec G404
		r := rand.New(rand.NewSource(portSeed))
		for pi := 0; pi < int(portCount); pi++ {
			portOffset := r.Intn(obfsvpn.PortHopRange)
			addr := net.JoinHostPort(obfs4Remote, fmt.Sprint(portOffset+obfsvpn.MinHopPort))
			obfsEndpoints = append(obfsEndpoints, &Obfs4Config{
				Cert:   certs[i],
				Remote: addr,
			})
		}
	}

	log.Printf("obfs4 endpoints: %+v", obfsEndpoints)
	return obfsEndpoints
}

func (c *HopClient) Start() (bool, error) {
	defer func() {
		c.state = stopped
		c.log("Start function ended")
	}()

	if c.IsStarted() {
		c.error("Cannot start proxy server, already running")
		return false, ErrAlreadyRunning
	}

	if len(c.obfs4Endpoints) == 0 {
		c.error("Cannot start proxy server, no valid endpoints")
		return false, ErrBadConfig
	}

	c.state = starting

	var err error

	obfs4Endpoint := c.obfs4Endpoints[0]

	c.obfs4Dialer, err = obfsvpn.NewDialerFromCert(obfs4Endpoint.Cert)
	if err != nil {
		return false, fmt.Errorf("could not dial obfs4 remote: %w", err)
	}

	if c.kcp {
		c.obfs4Dialer.DialFunc = func(network, address string) (net.Conn, error) {
			c.log("Dialing kcp://%s", address)
			return kcp.Dial(address)
		}
	}

	obfs4Conn, err := c.obfs4Dialer.Dial("tcp", obfs4Endpoint.Remote)
	if err != nil {
		c.error("Could not dial obfs4 remote: %v", err)
	}

	c.obfs4Conns = []net.Conn{obfs4Conn}

	// We want a non-crypto RNG so that we can share a seed
	// #nosec G404
	rand.Seed(time.Now().UnixNano())

	c.state = running

	proxyAddr, err := net.ResolveUDPAddr("udp", c.ProxyAddr)
	if err != nil {
		return false, fmt.Errorf("cannot resolve UDP addr: %w", err)
	}

	c.openvpnConn, err = net.ListenUDP("udp", proxyAddr)
	if err != nil {
		return false, fmt.Errorf("error accepting udp connection: %w", err)
	}

	go c.hop()

	go c.readUDPWriteTCP()

	go c.readTCPWriteUDP()

	<-c.ctx.Done()

	return true, nil
}

// pickRandomRemote returns a random remote from the internal array.
// An obvious improvement to this function is to check the number of failures in c.obfs4Failures and avoid
// a given remote if it failed more than a threshold. A consecuence is that
// we'll have to return an unrecoverable error from hop() if there are no
// more usable remotes. If we ever want to get fancy, an even better heuristic
// can be to avoid IPs that have more failures than the average.
func (c *HopClient) pickRandomEndpoint() *Obfs4Config {
	// #nosec G404
	i := rand.Intn(len(c.obfs4Endpoints))
	endpoint := c.obfs4Endpoints[i]
	// here we could check if the number of failures is ok-ish. we can also do moving averages etc.
	return endpoint
}

func (c *HopClient) hop() {
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

		c.log("HOPPING to %+v", newRemote)

		obfs4Dialer, err := obfsvpn.NewDialerFromCert(obfs4Endpoint.Cert)
		if err != nil {
			c.error("Could not dial obfs4 remote: %v", err)
			return
		}

		if c.kcp {
			obfs4Dialer.DialFunc = func(network, address string) (net.Conn, error) {
				c.log("Dialing kcp://%s", address)
				return kcp.Dial(address)
			}
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

func (c *HopClient) cleanupOldConn() {
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

func (c *HopClient) readUDPWriteTCP() {
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

func (c *HopClient) getUsableConnection() (net.Conn, error) {
	c.outLock.Lock()
	defer c.outLock.Unlock()

	if len(c.obfs4Conns) == 0 {
		return nil, errors.New("no usable connection")
	} else {
		return c.obfs4Conns[0], nil
	}
}

func (c *HopClient) readTCPWriteUDP() {
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

func (c *HopClient) Stop() (bool, error) {
	if !c.IsStarted() {
		return false, ErrNotRunning
	}

	c.stop()

	c.state = stopped

	return true, nil
}

func (c *HopClient) log(format string, a ...interface{}) {
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

func (c *HopClient) error(format string, a ...interface{}) {
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

func (c *HopClient) IsStarted() bool {
	return c.state != stopped
}

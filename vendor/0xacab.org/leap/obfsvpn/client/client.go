// Package client exposes a socks5 proxy that uses obfs4 to communicate with the server,
// with an optional kcp wire transport.
package client

import (
	"errors"
	"fmt"
	"log"
	"net"

	"0xacab.org/leap/obfsvpn"

	"github.com/kalikaneko/socks5"
	"github.com/xtaci/kcp-go"
)

var (
	ErrAlreadyRunning = errors.New("already initialized")
	ErrNotRunning     = errors.New("server not running")
)

type Client struct {
	kcp         bool
	SocksAddr   string
	obfs4Cert   string
	server      *socks5.Server
	EventLogger EventLogger
}

type EventLogger interface {
	Log(state string, message string)
	Error(message string)
}

func NewClient(kcp bool, socksAddr, obfs4Cert string) *Client {
	return &Client{
		kcp:       kcp,
		SocksAddr: socksAddr,
		obfs4Cert: obfs4Cert,
	}
}

func (c *Client) Start() (bool, error) {
	defer func() {
		c.log("STOPPED", "")
	}()

	if c.IsStarted() {
		c.error("Cannot start proxy server, already running")
		return false, ErrAlreadyRunning
	}

	c.server = &socks5.Server{
		Addr:   c.SocksAddr,
		BindIP: "127.0.0.1",
	}

	dialer, err := obfsvpn.NewDialerFromCert(c.obfs4Cert)
	if err != nil {
		c.error("Error getting dialer: %v\n", err)
		return false, err
	}

	if c.kcp {
		dialer.DialFunc = func(network, address string) (net.Conn, error) {
			c.log("RUNNING", "Dialing kcp://%s\n", address)
			return kcp.Dial(address)
		}
	}

	c.server.Dial = dialer.Dial

	c.log("RUNNING", "[+] Starting socks5 proxy at %s\n", c.SocksAddr)
	if err := c.server.ListenAndServe(); err != nil {
		c.error("error while listening: %v\n", err)
		c.server = nil
		return false, err
	}
	return true, nil
}

func (c *Client) Stop() (bool, error) {
	if !c.IsStarted() {
		return false, ErrNotRunning
	}

	if err := c.server.Close(); err != nil {
		c.error("error while stopping: %v\n", err)
		return false, err
	}

	c.server = nil
	return true, nil
}

func (c *Client) log(state string, format string, a ...interface{}) {
	if c.EventLogger != nil {
		c.EventLogger.Log(state, fmt.Sprintf(format, a...))
		return
	}
	if format == "" {
		log.Print(a...)
		return
	}
	log.Printf(format, a...)
}

func (c *Client) error(format string, a ...interface{}) {
	if c.EventLogger != nil {
		c.EventLogger.Error(fmt.Sprintf(format, a...))
		return
	}
	if format == "" {
		log.Print(a...)
		return
	}
	log.Printf(format, a...)
}

func (c *Client) IsStarted() bool {
	return c.server != nil
}

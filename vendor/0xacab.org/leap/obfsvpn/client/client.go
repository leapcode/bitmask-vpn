// Package client exposes a socks5 proxy that uses obfs4 to communicate with the server,
// with an optional kcp wire transport.
package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"0xacab.org/leap/obfsvpn/obfsvpn"
	"github.com/kalikaneko/socks5"
	"github.com/xtaci/kcp-go"
)

var (
	ErrAlreadyRunning = errors.New("already initialized")
	ErrNotRunning     = errors.New("server not running")
	ErrBadConfig      = errors.New("configuration error")
)

type Client struct {
	ctx         context.Context
	kcp         bool
	SocksAddr   string
	obfs4Cert   string
	server      *socks5.Server
	EventLogger EventLogger
	mux         sync.Mutex
}

type EventLogger interface {
	Log(state string, message string)
	Error(message string)
}

func NewClient(ctx context.Context, kcp bool, socksAddr, obfs4Cert string) ObfsClient {
	return &Client{
		ctx:       ctx,
		kcp:       kcp,
		obfs4Cert: obfs4Cert,
		SocksAddr: socksAddr,
	}
}

func (c *Client) Start() (bool, error) {
	c.mux.Lock()

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

	switch {
	case c.kcp:
		dialer.DialFunc = func(network, address string) (net.Conn, error) {
			c.log("RUNNING", "client.Start(): dialing kcp://%s\n", address)
			return kcp.Dial(address)
		}
	}

	c.server.Dial = dialer.Dial

	c.log("RUNNING", "[+] Starting socks5 proxy at %s\n", c.SocksAddr)

	errCh := make(chan error)
	go c.startSocksServer(errCh)

	c.mux.Unlock()

	select {
	case <-c.ctx.Done():
		return true, nil
	case err := <-errCh:
		c.server = nil
		return false, err
	}
}

func (c *Client) startSocksServer(ch chan error) {
	if err := c.server.ListenAndServe(); err != nil {
		c.error("error while listening: %v\n", err)
		ch <- err
	}
}

func (c *Client) Stop() (bool, error) {
	if !c.IsStarted() {
		return false, ErrNotRunning
	}

	c.mux.Lock()
	defer c.mux.Unlock()

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

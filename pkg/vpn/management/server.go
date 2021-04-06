// Copyright (c) 2016 Martin Atkins
// Copyright (c) 2021 LEAP Encryption Access Project

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package management

import (
	"net"
	"time"
)

// MgmtListener accepts incoming connections from OpenVPN.
//
// The primary way to instantiate this type is via the function Listen.
// See its documentation for more information.
type MgmtListener struct {
	l net.Listener
}

// NewMgmtListener constructs a MgmtListener from an already-established
// net.Listener. In most cases it will be more convenient to use
// the function Listen.
func NewMgmtListener(l net.Listener) *MgmtListener {
	return &MgmtListener{l}
}

// Listen opens a listen port and awaits incoming connections from OpenVPN
// processes.
//
// OpenVPN will behave in this manner when launched with the following options:
//
//    --management ipaddr port --management-client
//
// Note that in this case the terminology is slightly confusing, since from
// the standpoint of TCP/IP it is OpenVPN that is the client and our program
// that is the server, but once the connection is established the channel
// is indistinguishable from the situation where OpenVPN exposed a management
// *server* and we connected to it. Thus we still refer to our program as
// the "client" and OpenVPN as the "server" once the connection is established.
//
// When running on Unix systems it's possible to instead listen on a Unix
// domain socket. To do this, pass an absolute path to the socket as
// the listen address, and then run OpenVPN with the following options:
//
//    --management /path/to/socket unix --management-client
//
func Listen(laddr string) (*MgmtListener, error) {
	proto := "tcp"
	if len(laddr) > 0 && laddr[0] == '/' {
		proto = "unix"
	}
	listener, err := net.Listen(proto, laddr)
	if err != nil {
		return nil, err
	}

	return NewMgmtListener(listener), nil
}

// Accept waits for and returns the next connection.
func (l *MgmtListener) Accept() (*IncomingConn, error) {
	conn, err := l.l.Accept()
	if err != nil {
		return nil, err
	}

	return &IncomingConn{conn}, nil
}

// Close closes the listener. Any blocked Accept operations
// will be blocked and each will return an error.
func (l *MgmtListener) Close() error {
	return l.l.Close()
}

// Addr returns the listener's network address.
func (l *MgmtListener) Addr() net.Addr {
	return l.l.Addr()
}

// Serve will await new connections and call the given handler
// for each.
//
// Serve does not return unless the listen port is closed; a non-nil
// error is always returned.
func (l *MgmtListener) Serve(handler IncomingConnHandler) error {
	defer l.Close()

	var tempDelay time.Duration

	for {
		incoming, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				// Wait a while before we try again.
				time.Sleep(tempDelay)
				continue
			} else {
				// Listen socket is permanently closed or errored,
				// so it's time for us to exit.
				return err
			}
		}

		// always reset our retry delay once we successfully read
		tempDelay = 0

		go handler.ServeOpenVPNMgmt(*incoming)
	}
}

type IncomingConn struct {
	conn net.Conn
}

// Open initiates communication with the connected OpenVPN process,
// and establishes the channel on which events will be delivered.
//
// See the documentation for NewClient for discussion about the requirements
// for eventCh.
func (ic IncomingConn) Open(eventCh chan<- Event) *MgmtClient {
	return NewClient(ic.conn, eventCh)
}

// Close abruptly closes the socket connected to the OpenVPN process.
//
// This is a rather abrasive way to close the channel, intended for rejecting
// unwanted incoming clients that may or may not speak the OpenVPN protocol.
//
// Once communication is accepted and established, it is generally better
// to close the connection gracefully using commands on the client returned
// from Open.
func (ic IncomingConn) Close() error {
	return ic.conn.Close()
}

type IncomingConnHandler interface {
	ServeOpenVPNMgmt(IncomingConn)
}

// IncomingConnHandlerFunc is an adapter to allow the use of ordinary
// functions as connection handlers.
//
// Given a function with the appropriate signature, IncomingConnHandlerFunc(f)
// is an IncomingConnHandler that calls f.
type IncomingConnHandlerFunc func(IncomingConn)

func (f IncomingConnHandlerFunc) ServeOpenVPNMgmt(i IncomingConn) {
	f(i)
}

// ListenAndServe creates a MgmtListener for the given listen address
// and then calls AcceptAndServe on it.
//
// This is just a convenience wrapper. See the AcceptAndServe method for
// more details. Just as with AcceptAndServe, this function does not return
// except on error; in addition to the error cases handled by AcceptAndServe,
// this function may also fail if the listen socket cannot be established
// in the first place.
func ListenAndServe(laddr string, handler IncomingConnHandler) error {
	listener, err := Listen(laddr)
	if err != nil {
		return err
	}

	return listener.Serve(handler)
}

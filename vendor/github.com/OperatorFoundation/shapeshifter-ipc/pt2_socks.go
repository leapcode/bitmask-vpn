package pt

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	socksVersion = 0x05

	socksAuthNoneRequired        = 0x00
	socksAuthPrivateMethodPT2    = 0x80
	socksAuthNoAcceptableMethods = 0xff

	socksCmdConnect = 0x01
	socksRsv        = 0x00

	socksAtypeV4         = 0x01
	socksAtypeDomainName = 0x03
	socksAtypeV6         = 0x04

	socksRepSucceeded = 0x00
	// "general SOCKS server failure"
	SocksRepGeneralFailure = 0x01
	// "connection not allowed by ruleset"
	SocksRepConnectionNotAllowed = 0x02
	// "Network unreachable"
	SocksRepNetworkUnreachable = 0x03
	// "Host unreachable"
	SocksRepHostUnreachable = 0x04
	// "Connection refused"
	SocksRepConnectionRefused = 0x05
	// "TTL expired"
	SocksRepTTLExpired = 0x06
	// "Command not supported"
	SocksRepCommandNotSupported = 0x07
	// "Address type not supported"
	SocksRepAddressNotSupported = 0x08
)

// Put a sanity timeout on how long we wait for a SOCKS request.
const socksRequestTimeout = 5 * time.Second

// SocksRequest describes a SOCKS request.
type SocksRequest struct {
	// The endpoint requested by the client as a "host:port" string.
	Target string
	// The parsed contents of private authentication method as a key–value mapping.
	Args Args
}

// SocksConn encapsulates a net.Conn and information associated with a SOCKS request.
type SocksConn struct {
	net.Conn
	Req SocksRequest
}

// Send a message to the proxy client that access to the given address is
// granted. Addr is ignored, and "0.0.0.0:0" is always sent back for
// BND.ADDR/BND.PORT in the SOCKS response.
func (conn *SocksConn) Grant(addr *net.TCPAddr) error {
	return sendSocks5ResponseGranted(conn)
}

// Send a message to the proxy client that access was rejected or failed.  This
// sends back a "General Failure" error code.  RejectReason should be used if
// more specific error reporting is desired.
func (conn *SocksConn) Reject() error {
	return conn.RejectReason(SocksRepGeneralFailure)
}

// Send a message to the proxy client that access was rejected, with the
// specific error code indicating the reason behind the rejection.
func (conn *SocksConn) RejectReason(reason byte) error {
	return sendSocks5ResponseRejected(conn, reason)
}

// SocksListener wraps a net.Listener in order to read a SOCKS request on Accept.
//
// 	func handleConn(conn *pt.SocksConn) error {
// 		defer conn.Close()
// 		remote, err := net.Dial("tcp", conn.Req.Target)
// 		if err != nil {
// 			conn.Reject()
// 			return err
// 		}
// 		defer remote.Close()
// 		err = conn.Grant(remote.RemoteAddr().(*net.TCPAddr))
// 		if err != nil {
// 			return err
// 		}
// 		// do something with conn and remote
// 		return nil
// 	}
// 	...
// 	ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	for {
// 		conn, err := ln.AcceptSocks()
// 		if err != nil {
// 			log.Printf("accept error: %s", err)
// 			if e, ok := err.(net.Error); ok && e.Temporary() {
// 				continue
// 			}
// 			break
// 		}
// 		go handleConn(conn)
// 	}
type SocksListener struct {
	net.Listener
}

// Open a net.Listener according to network and laddr, and return it as a
// SocksListener.
func ListenSocks(network, laddr string) (*SocksListener, error) {
	ln, err := net.Listen(network, laddr)
	if err != nil {
		return nil, err
	}
	return NewSocksListener(ln), nil
}

// Create a new SocksListener wrapping the given net.Listener.
func NewSocksListener(ln net.Listener) *SocksListener {
	return &SocksListener{ln}
}

// Accept is the same as AcceptSocks, except that it returns a generic net.Conn.
// It is present for the sake of satisfying the net.Listener interface.
func (ln *SocksListener) Accept() (net.Conn, error) {
	return ln.AcceptSocks()
}

// Call Accept on the wrapped net.Listener, do SOCKS negotiation, and return a
// SocksConn. After accepting, you must call either conn.Grant or conn.Reject
// (presumably after trying to connect to conn.Req.Target).
//
// Errors returned by AcceptSocks may be temporary (for example, EOF while
// reading the request, or a badly formatted userid string), or permanent (e.g.,
// the underlying socket is closed). You can determine whether an error is
// temporary and take appropriate action with a type conversion to net.Error.
// For example:
//
// 	for {
// 		conn, err := ln.AcceptSocks()
// 		if err != nil {
// 			if e, ok := err.(net.Error); ok && e.Temporary() {
// 				log.Printf("temporary accept error; trying again: %s", err)
// 				continue
// 			}
// 			log.Printf("permanent accept error; giving up: %s", err)
// 			break
// 		}
// 		go handleConn(conn)
// 	}
func (ln *SocksListener) AcceptSocks() (*SocksConn, error) {
retry:
	c, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}
	conn := new(SocksConn)
	conn.Conn = c
	err = conn.SetDeadline(time.Now().Add(socksRequestTimeout))
	if err != nil {
		conn.Close()
		goto retry
	}
	conn.Req, err = socks5Handshake(conn, false)
	if err != nil {
		conn.Close()
		goto retry
	}
	err = conn.SetDeadline(time.Time{})
	if err != nil {
		conn.Close()
		goto retry
	}
	return conn, nil
}

// Returns "socks5", suitable to be included in a call to Cmethod.
func (ln *SocksListener) Version() string {
	return "socks5"
}

// socks5handshake conducts the SOCKS5 handshake up to the point where the
// client command is read and the proxy must open the outgoing connection.
// Returns a SocksRequest.
func socks5Handshake(s io.ReadWriter, needOptions bool) (req SocksRequest, err error) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// Negotiate the authentication method.
	var method byte
	if method, err = socksNegotiateAuth(rw, needOptions); err != nil {
		return
	}

	// Authenticate the client.
	if err = socksAuthenticate(rw, method, &req); err != nil {
		return
	}

	// Read the command.
	err = socksReadCommand(rw, &req)
	return
}

// socksNegotiateAuth negotiates the authentication method and returns the
// selected method as a byte.  On negotiation failures an error is returned.
func socksNegotiateAuth(rw *bufio.ReadWriter, needOptions bool) (method byte, err error) {
	// Validate the version.
	if err = socksReadByteVerify(rw, "version", socksVersion); err != nil {
		return
	}

	// Read the number of methods.
	var nmethods byte
	if nmethods, err = socksReadByte(rw); err != nil {
		return
	}

	// Read the methods.
	var methods []byte
	if methods, err = socksReadBytes(rw, int(nmethods)); err != nil {
		return
	}

	// Pick the most "suitable" method.
	method = socksAuthNoAcceptableMethods
	for _, m := range methods {
		if needOptions {
			switch m {
			case socksAuthNoneRequired:
				// Pick PT 2.0 private authentication method over None if the client happens to
				// send both and SOCKS header options are needed.
				if method == socksAuthNoAcceptableMethods {
					method = m
				}

			case socksAuthPrivateMethodPT2:
				method = m
			}
		} else {
			switch m {
			case socksAuthPrivateMethodPT2:
				// Pick None over PT 2.0 private authentication method if the client happens to
				// send both and SOCKS header options are not needed.
				if method == socksAuthNoAcceptableMethods {
					method = m
				}

			case socksAuthNoneRequired:
				method = m
			}
		}
	}

	// Send the negotiated method.
	var msg [2]byte
	msg[0] = socksVersion
	msg[1] = method
	if _, err = rw.Writer.Write(msg[:]); err != nil {
		return
	}

	if err = socksFlushBuffers(rw); err != nil {
		return
	}
	return
}

// socksAuthenticate authenticates the client via the chosen authentication
// mechanism.
func socksAuthenticate(rw *bufio.ReadWriter, method byte, req *SocksRequest) (err error) {
	switch method {
	case socksAuthNoneRequired:
		// Straight into reading the connect.

	case socksAuthPrivateMethodPT2:
		if err = socksAuthPT2(rw, req); err != nil {
			return
		}

	case socksAuthNoAcceptableMethods:
		err = fmt.Errorf("SOCKS method select had no compatible methods")
		return

	default:
		err = fmt.Errorf("SOCKS method select picked a unsupported method 0x%02x", method)
		return
	}

	if err = socksFlushBuffers(rw); err != nil {
		return
	}
	return
}

// socksAuthRFC1929 authenticates the client via Pluggable Transports 2.0, draft 1
// private auth method.
func socksAuthPT2(rw *bufio.ReadWriter, req *SocksRequest) (err error) {
	// Read the authentication data.
	var len uint32
	if len, err = socksReadUint32(rw); err != nil {
		return
	}
	if len == 0 {
		err = fmt.Errorf("PT 2.0 authentication data with 0 length")
		return
	}
	var data []byte
	if data, err = socksReadBytes(rw, int(len)); err != nil {
		return
	}

	var result string = string(data)

	// Parse the authentication data according to the PT 2.0 specification
	if req.Args, err = ParsePT2ClientParameters(result); err != nil {
		fmt.Println("Error parsing PT2 client parameters", err)
		return
	}

	return
}

// socksReadCommand reads a SOCKS5 client command and parses out the relevant
// fields into a SocksRequest.  Only CMD_CONNECT is supported.
func socksReadCommand(rw *bufio.ReadWriter, req *SocksRequest) (err error) {
	sendErrResp := func(reason byte) {
		// Swallow errors that occur when writing/flushing the response,
		// connection will be closed anyway.
		sendSocks5ResponseRejected(rw, reason)
		socksFlushBuffers(rw)
	}

	// Validate the fixed parts of the command message.
	if err = socksReadByteVerify(rw, "version", socksVersion); err != nil {
		sendErrResp(SocksRepGeneralFailure)
		return
	}
	if err = socksReadByteVerify(rw, "command", socksCmdConnect); err != nil {
		sendErrResp(SocksRepCommandNotSupported)
		return
	}
	if err = socksReadByteVerify(rw, "reserved", socksRsv); err != nil {
		sendErrResp(SocksRepGeneralFailure)
		return
	}

	// Read the destination address/port.
	// XXX: This should probably eventually send socks 5 error messages instead
	// of rudely closing connections on invalid addresses.
	var atype byte
	if atype, err = socksReadByte(rw); err != nil {
		return
	}
	var host string
	switch atype {
	case socksAtypeV4:
		var addr []byte
		if addr, err = socksReadBytes(rw, net.IPv4len); err != nil {
			return
		}
		host = net.IPv4(addr[0], addr[1], addr[2], addr[3]).String()

	case socksAtypeDomainName:
		var alen byte
		if alen, err = socksReadByte(rw); err != nil {
			return
		}
		if alen == 0 {
			err = fmt.Errorf("SOCKS request had domain name with 0 length")
			return
		}
		var addr []byte
		if addr, err = socksReadBytes(rw, int(alen)); err != nil {
			return
		}
		host = string(addr)

	case socksAtypeV6:
		var rawAddr []byte
		if rawAddr, err = socksReadBytes(rw, net.IPv6len); err != nil {
			return
		}
		addr := make(net.IP, net.IPv6len)
		copy(addr[:], rawAddr[:])
		host = fmt.Sprintf("[%s]", addr.String())

	default:
		sendErrResp(SocksRepAddressNotSupported)
		err = fmt.Errorf("SOCKS request had unsupported address type 0x%02x", atype)
		return
	}
	var rawPort []byte
	if rawPort, err = socksReadBytes(rw, 2); err != nil {
		return
	}
	port := int(rawPort[0])<<8 | int(rawPort[1])<<0

	if err = socksFlushBuffers(rw); err != nil {
		return
	}

	req.Target = fmt.Sprintf("%s:%d", host, port)
	return
}

// Send a SOCKS5 response with the given code. BND.ADDR/BND.PORT is always the
// IPv4 address/port "0.0.0.0:0".
func sendSocks5Response(w io.Writer, code byte) error {
	resp := make([]byte, 4+4+2)
	resp[0] = socksVersion
	resp[1] = code
	resp[2] = socksRsv
	resp[3] = socksAtypeV4

	// BND.ADDR/BND.PORT should be the address and port that the outgoing
	// connection is bound to on the proxy, but Tor does not use this
	// information, so all zeroes are sent.

	_, err := w.Write(resp[:])
	return err
}

// Send a SOCKS5 response code 0x00.
func sendSocks5ResponseGranted(w io.Writer) error {
	return sendSocks5Response(w, socksRepSucceeded)
}

// Send a SOCKS5 response with the provided failure reason.
func sendSocks5ResponseRejected(w io.Writer, reason byte) error {
	return sendSocks5Response(w, reason)
}

func socksFlushBuffers(rw *bufio.ReadWriter) error {
	if err := rw.Writer.Flush(); err != nil {
		return err
	}
	if rw.Reader.Buffered() > 0 {
		return fmt.Errorf("%d bytes left after SOCKS message", rw.Reader.Buffered())
	}
	return nil
}

func socksReadByte(rw *bufio.ReadWriter) (byte, error) {
	return rw.Reader.ReadByte()
}

func socksReadUint32(rw *bufio.ReadWriter) (uint32, error) {
	buffer, err := socksReadBytes(rw, 4)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(buffer), nil
}

func socksReadBytes(rw *bufio.ReadWriter, n int) ([]byte, error) {
	ret := make([]byte, n)
	if _, err := io.ReadFull(rw.Reader, ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func socksReadByteVerify(rw *bufio.ReadWriter, descr string, expected byte) error {
	val, err := socksReadByte(rw)
	if err != nil {
		return err
	}
	if val != expected {
		return fmt.Errorf("SOCKS message field %s was 0x%02x, not 0x%02x", descr, val, expected)
	}
	return nil
}

var _ net.Listener = (*SocksListener)(nil)

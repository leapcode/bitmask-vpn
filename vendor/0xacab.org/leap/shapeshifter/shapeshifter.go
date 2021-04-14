package shapeshifter

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/OperatorFoundation/obfs4/common/ntor"
	"github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4"
	"golang.org/x/net/proxy"
)

type Logger interface {
	Log(msg string)
}

type ShapeShifter struct {
	Cert      string
	IatMode   int
	Target    string // remote ip:port obfs4 server
	SocksAddr string // -proxylistenaddr in shapeshifter-dispatcher
	Logger    Logger
	ln        net.Listener
	errChan   chan error
}

func (ss *ShapeShifter) Open() error {
	err := ss.checkOptions()
	if err != nil {
		return err
	}

	ss.ln, err = net.Listen("tcp", ss.SocksAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err.Error())
	}

	go ss.clientAcceptLoop()
	return nil
}

func (ss *ShapeShifter) Close() error {
	if ss.ln != nil {
		return ss.ln.Close()
	}
	if ss.errChan != nil {
		close(ss.errChan)
	}
	return nil
}

func (ss *ShapeShifter) GetErrorChannel() chan error {
	if ss.errChan == nil {
		ss.errChan = make(chan error, 2)
	}
	return ss.errChan
}

func (ss ShapeShifter) clientAcceptLoop() error {
	for {
		conn, err := ss.ln.Accept()
		if err != nil {
			if e, ok := err.(net.Error); ok && !e.Temporary() {
				return err
			}
			ss.sendError("Error accepting connection: %v", err)
			continue
		}
		go ss.clientHandler(conn)
	}
}

func (ss ShapeShifter) clientHandler(conn net.Conn) {
	defer conn.Close()

	dialer := proxy.Direct
	transport, err := obfs4.NewObfs4Client(ss.Cert, ss.IatMode, dialer)
	if err != nil {
		ss.sendError("Can not create an obfs4 client (cert: %s, iat-mode: %d): %v", ss.Cert, ss.IatMode, err)
		return
	}
	remote, err := transport.Dial(ss.Target)
	if err != nil {
		ss.sendError("outgoing connection failed %s: %v", ss.Target, err)
		return
	}
	if remote == nil {
		ss.sendError("outgoing connection failed %s", ss.Target)
		return
	}
	defer remote.Close()

	err = copyLoop(conn, remote)
	if err != nil {
		ss.sendError("%s - closed connection: %v", ss.Target, err)
	} else {
		log.Printf("%s - closed connection", ss.Target)
	}

	return
}

func copyLoop(a net.Conn, b net.Conn) error {
	// Note: b is always the pt connection.  a is the SOCKS/ORPort connection.
	errChan := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer b.Close()
		defer a.Close()
		_, err := io.Copy(b, a)
		errChan <- err
	}()
	go func() {
		defer wg.Done()
		defer a.Close()
		defer b.Close()
		_, err := io.Copy(a, b)
		errChan <- err
	}()

	// Wait for both upstream and downstream to close.  Since one side
	// terminating closes the other, the second error in the channel will be
	// something like EINVAL (though io.Copy() will swallow EOF), so only the
	// first error is returned.
	wg.Wait()
	if len(errChan) > 0 {
		return <-errChan
	}

	return nil
}

func (ss *ShapeShifter) checkOptions() error {
	if ss.SocksAddr == "" {
		ss.SocksAddr = "127.0.0.1:0"
	}
	return isCertValid(ss.Cert)
}

func (ss *ShapeShifter) sendError(format string, a ...interface{}) {
	if ss.Logger != nil {
		ss.Logger.Log(fmt.Sprintf(format, a...))
		return
	}

	if ss.errChan == nil {
		ss.errChan = make(chan error, 2)
	}
	select {
	case ss.errChan <- fmt.Errorf(format, a...):
	default:
		log.Printf(format, a...)
	}
}

func isCertValid(cert string) error {
	// copied from github.com/OperatorFoundation/shapeshifter-transports/transports/obfs4/statefile.go
	const certSuffix = "=="
	const certLength = ntor.NodeIDLength + ntor.PublicKeyLength

	if cert == "" {
		return fmt.Errorf("obfs4 transport missing cert argument")
	}

	decoded, err := base64.StdEncoding.DecodeString(cert + certSuffix)
	if err != nil {
		return fmt.Errorf("failed to decode cert: %s", err)
	}

	if len(decoded) != certLength {
		return fmt.Errorf("cert length %d is invalid", len(decoded))
	}

	return nil
}

package obfsvpn

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/quic-go/quic-go"
)

type QUICConfig struct {
	Enabled bool
	TLSCert *tls.Certificate
}

type QUICConn struct {
	quic.Stream
	conn quic.Connection
}

func (q *QUICConn) Close() error {
	return q.conn.CloseWithError(0, "closing obfs4 connection")
}

func (q *QUICConn) LocalAddr() net.Addr {
	return q.conn.LocalAddr()
}

func (q *QUICConn) RemoteAddr() net.Addr {
	return q.conn.RemoteAddr()
}

type QUICListener struct {
	ql  *quic.Listener
	ctx context.Context
}

func (q *QUICListener) Accept() (net.Conn, error) {
	conn, err := q.ql.Accept(q.ctx)
	if err != nil {
		return nil, err
	}

	str, err := conn.AcceptStream(q.ctx)
	if err != nil {
		return nil, err
	}

	return &QUICConn{str, conn}, nil
}

func (q *QUICListener) Close() error {
	return q.ql.Close()
}

// Addr returns the listener's network address.
func (q *QUICListener) Addr() net.Addr {
	return q.ql.Addr()
}

func DefaultQUICConfig() *QUICConfig {
	return &QUICConfig{
		Enabled: true,
	}
}

func GetQUICDialer(ctx context.Context, quicConfig QUICConfig, logger func(format string, a ...interface{})) func(network, address string) (net.Conn, error) {
	quicConf := &quic.Config{}
	return func(network, address string) (net.Conn, error) {
		tlsConf := &tls.Config{
			//nolint:gosec // We can skip actual TLS verification because we're only using this layer for obfuscation
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}

		if logger != nil {
			logger("Dialing quic://%s", address)
		}

		conn, err := quic.DialAddr(ctx, address, tlsConf, quicConf)
		if err != nil {
			return nil, err
		}

		str, err := conn.OpenStreamSync(ctx)
		if err != nil {
			return nil, err
		}

		return &QUICConn{str, conn}, nil
	}
}

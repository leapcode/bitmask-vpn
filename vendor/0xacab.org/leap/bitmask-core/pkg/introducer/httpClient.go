package introducer

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/obfsvpn/obfsvpn"
	"github.com/xtaci/kcp-go"
)

// NewHTTPClientFromIntroducer returns an http.Client that will use the passed introducer.
func NewHTTPClientFromIntroducer(introducer *Introducer) (*http.Client, error) {

	// Validate the introducer
	if err := introducer.Validate(); err != nil {
		return nil, err
	}

	// Get an OBFS4 dialer
	dialer, err := obfsvpn.NewDialerFromCert(introducer.Cert)
	if err != nil {
		return nil, err
	}

	switch {
	case introducer.KCP:
		dialer.DialFunc = func(network, address string) (net.Conn, error) {
			log.Debug().Msg(fmt.Sprintf("dialing kcp://%s", address))
			return kcp.Dial(address)
		}
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			ch := make(chan struct {
				conn net.Conn
				err  error
			}, 1)
			go func() {
				conn, err := dialer.Dial(network, introducer.Addr)
				ch <- struct {
					conn net.Conn
					err  error
				}{conn, err}
			}()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case result := <-ch:
				return result.conn, result.err
			}
		},
	}

	client := &http.Client{
		Transport: transport,
	}
	return client, nil
}

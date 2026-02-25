package introducer

import (
	"context"
	"net"
	"net/http"

	"0xacab.org/leap/obfsvpn/obfsvpn"
	"github.com/rs/zerolog/log"
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
		dialer.DialFunc = obfsvpn.GetKCPDialer(*obfsvpn.DefaultKCPConfig(), log.Debug().Msgf)
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

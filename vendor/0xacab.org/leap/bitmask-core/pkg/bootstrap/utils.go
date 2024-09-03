package bootstrap

import (
	gotls "crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	utls "github.com/refraction-networking/utls"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
)

// Parses API URL of menshen. Returns hostname/ip, port, useTLS
func parseApiURL(menshenURL string) (string, int, bool, error) {
	url, err := url.Parse(menshenURL)
	if err != nil {
		return "", -1, false, fmt.Errorf("Could not parse API url %s: %s", url, err)
	}

	hostname := url.Hostname()
	useTLS := url.Scheme != "http"

	var port int
	if url.Port() == "" {
		port = 443
	} else {
		port, err = strconv.Atoi(url.Port())
		if err != nil {
			return "", -1, false, fmt.Errorf("Could not parse port to int %s: %s", url.Port(), err)
		}
	}

	log.Trace().
		Bool("useTLS", useTLS).
		Str("hostname", hostname).
		Int("port", port).
		Msg("Parsed API URL")

	return hostname, port, useTLS, nil
}

func (c *Config) getAPIClient() *http.Client {
	if c.UseTLS {
		client := &http.Client{
			Transport: &http2.Transport{
				// Hook into TLS connection buildup to resolve IP with DNS over HTTP (DoH)
				DialTLS: func(network, addr string, tlsCfg *gotls.Config) (net.Conn, error) {
					if c.ResolveWithDoH {
						log.Debug().
							Str("domain", addr).
							Msg("Resolving host with DNS over HTTPs")

						ip4, err := dohQuery(c.Host)
						if err != nil {
							return nil, err
						}

						log.Debug().
							Str("domain", addr).
							Str("ip4", ip4).
							Msg("Sucessfully resolved host via DNS over HTTPs")
						addr = fmt.Sprintf("%s:%d", ip4, c.Port)
					}

					roller, err := utls.NewRoller()
					if err != nil {
						return nil, err
					}
					uconn, err := roller.Dial(network, addr, c.Host)
					if err != nil {
						return nil, err
					}

					uconn.SetSNI(c.Host)
					return uconn, err
				},
			},
			Timeout: time.Duration(30) * time.Second,
		}
		return client
	} else {
		return &http.Client{Timeout: time.Duration(30) * time.Second}
	}
}

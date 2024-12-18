package bootstrap

import (
	gotls "crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	bitmask_storage "0xacab.org/leap/bitmask-core/pkg/storage"
	"github.com/go-openapi/runtime"
	openapi "github.com/go-openapi/runtime/client"
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

// Returns authentication header (invite token) from database
// Returns nil if no introducer is saved or an error occurs
func (api *API) getInviteTokenAuth() runtime.ClientAuthInfoWriter {
	if len(api.config.Introducer) == 0 {
		return nil
	}

	log.Trace().Msg("Getting invite token from db")
	storage, err := bitmask_storage.GetStorage()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not get storage to load invite token")
		return nil
	}

	introducer, err := storage.GetIntroducerByFQDN(api.config.Host)
	if err != nil {
		log.Debug().
			Str("err", err.Error()).
			Str("fqdn", api.config.Host).
			Msg("Could not get introducer by fqdn")
		return nil
	}

	if len(introducer.Auth) == 0 {
		log.Warn().Msg("An introducer was found for this fqdn, but the invite token is empty")
		return nil
	}

	log.Debug().Msg("Sending invite token")
	return openapi.APIKeyAuth("x-menshen-auth-token", "header", introducer.Auth)
}

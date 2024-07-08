package bootstrap

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"

	"0xacab.org/leap/bitmask-core/models"
	"0xacab.org/leap/bitmask-core/pkg/client"
	"0xacab.org/leap/bitmask-core/pkg/client/provisioning"
	"0xacab.org/leap/bitmask-core/pkg/introducer"

	gotls "crypto/tls"

	utls "github.com/refraction-networking/utls"
)

type Config struct {
	// BaseCountry is an ISO-2 country code. If present, we will skip
	// geolocation lookup, and we will send our "base" country to menshen
	// when asking for gateways/bridges.
	BaseCountry string
	// Host we will connect to for API operations.
	Host string
	// Port we will connect to for API operations (default 443)
	Port int
	// Use TLS to connect to menshen (default: true)
	UseTLS bool
	// Introducer is an obfsucated introducer to use for all bootstrap operations.
	Introducer string
	// Proxy is a local SOCKS5 proxy for all bootstrap operations.
	Proxy string
	// ResolveWithDoH indicates whether we should use a DoH resolver.
	ResolveWithDoH bool
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

func NewConfig() *Config {
	return &Config{
		Port:           443,
		UseTLS:         true,
		ResolveWithDoH: true,
	}
}

type API struct {
	client     *client.MenshenAPI
	gateways   []*models.ModelsGateway
	httpClient *http.Client

	countryBase     string
	countryOverride bool
}

func NewAPI(cfg *Config) (*API, error) {
	transportConfig := client.DefaultTransportConfig()

	var intro *introducer.Introducer

	if cfg.Introducer != "" {
		intro, err := introducer.NewIntroducerFromURL(cfg.Introducer)
		if err != nil {
			return nil, err
		}

		// If we have received an introducer, we override the Host field
		// with the FQDN specified in the introducer, but lets remind the user of the override:
		if cfg.Host != "" && cfg.Host != intro.FQDN {
			return nil, fmt.Errorf("Invalid configuration. --host=%s will be overriden with --fqdn=%s "+
				"because introducer has precedence", cfg.Host, intro.FQDN)
		}
		cfg.Host = intro.FQDN
	}

	host := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	transportConfig = transportConfig.WithHost(host).WithSchemes([]string{"https"})

	if !cfg.UseTLS {
		transportConfig = transportConfig.WithSchemes([]string{"http"})
		log.Debug().Msg("Disabling DNS over HTTP (not using SSL)")
		cfg.ResolveWithDoH = false
	}

	client := client.NewHTTPClientWithConfig(nil, transportConfig)
	api := &API{
		client:   client,
		gateways: make([]*models.ModelsGateway, 0),
	}

	if cfg.BaseCountry != "" {
		api.countryOverride = true
		api.countryBase = cfg.BaseCountry
	}

	// Introducer has precedence over the Proxy parameter, unless it fails.
	// Above we've parsed the introducer URL, here we try to get an http client
	// configured to use it.
	// In the future, we might want to add a timeout and mark it as unusable if it fails.
	if intro != nil {
		client, err := introducer.NewHTTPClientFromIntroducer(intro)
		if client != nil {
			return nil, err
		}
		log.Info().Msg("Using obfuscated http client")
		api.httpClient = client
		// We got an http client configured to use the obfuscated introducer,
		// so we'll stop here.
		return api, nil
	}

	if cfg.Proxy != "" {
		client, err := getSocksProxyClient(cfg.Proxy)
		if err != nil {
			return nil, err
		}
		log.Debug().
			Str("proxy", cfg.Proxy).
			Msg("Enabled proxy")

		api.httpClient = client
		return api, nil
	}

	api.httpClient = cfg.getAPIClient()
	return api, nil
}

// proxyURI should be in the format like socks5://localhost:9050
func getSocksProxyClient(proxyString string) (*http.Client, error) {
	proxyURL, err := url.Parse(proxyString)
	if err != nil {
		return nil, err
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
		Timeout:   time.Duration(30) * time.Second,
	}
	return client, nil
}

// DoGeolocationLookup will try to fetch a valid country code from an online
// geolocation service (not controlled by us). This country code will be stored and
// sent in any subsequent resource queries to menshen (gateways/bridges), so that proximity
// is used in addition to the load information.
// This method should be called only once, right after initializing the API object.
func (api *API) DoGeolocationLookup() error {
	if api.countryOverride {
		log.Debug().
			Str("countryCode", api.countryBase).
			Msg("Skipping geolocation lookup as country is already set")
		return nil
	}

	// we're not using uTLS in this lookup, because the geoip service used (ubuntu) does not seem to support h2.
	// if this fails, we could explore other additional lookups (stun, etc) as the OONI probe does.
	// in any case, we should not sweat it too much: we can also play with different heuristics, as setting
	// up the region (continent) or time zone. that should be enough, if we consider load and probably some
	// other estimation of proximity.
	client := &http.Client{Timeout: time.Duration(30) * time.Second}
	cc, err := ubuntuGeoLookup(client)
	if err != nil {
		return err
	}
	// FIXME scrub if we're going to submit logs.
	log.Info().
		Str("countryCode", cc).
		Msg("Successfully got country code from geo API")
	api.countryBase = cc
	return nil
}

// call menshen endpoint /service and return response
// TODO: rename endpoint and and this function
// TODO: split /service into multiple endpoints:
// locations, openvpn arguments, serial+version, auth
func (api *API) GetService() (*models.ModelsEIPService, error) {
	params := provisioning.NewGet5ServiceParams()
	if api.httpClient != nil {
		params = params.WithHTTPClient(api.httpClient)
	}

	// TODO: menshen needs to accept cc as param too.
	/*
		if api.countryBase != "" {
			params.Cc = api.countryBase
		}
	*/

	service, err := api.client.Provisioning.Get5Service(params)
	if err != nil {
		return nil, err
	}
	return service.Payload, nil
}

// TODO: split /service endpoint into multiple endpoints
// then call this endpoint and return locations
// do not use an internal variable to store it
/*
func (api *API) Locations() interface{} {
	panic("TODO")
}
*/

// GatewayParams contains the fields that can be used to filter the listing of available gateways.
type GatewayParams struct {
	Location  string
	Port      string
	Transport string
	CC        string
}

// GetGateways returns a list of gateways (it it's enabled by the menshen
// API). It optionally accepts a GatewayParams object where you can set
// different filters.
func (api *API) GetGateways(p *GatewayParams) ([]*models.ModelsGateway, error) {
	params := provisioning.NewGet5GatewaysParams()
	if p != nil {
		params.Loc = &p.Location
		params.Port = &p.Port
		params.Tr = &p.Transport
		params.Cc = &p.CC
	}
	if api.httpClient != nil {
		params = params.WithHTTPClient(api.httpClient)
	}

	gateways, err := api.client.Provisioning.Get5Gateways(params)
	if err != nil {
		return nil, err
	}
	return gateways.Payload, err
}

// GetOpenVPNCert returns valid OpenVPN client credentials (certificate and
// private key)
func (api *API) GetOpenVPNCert() (string, error) {
	params := provisioning.NewGet5OpenvpnCertParams()
	if api.httpClient != nil {
		params = params.WithHTTPClient(api.httpClient)
	}

	cert, err := api.client.Provisioning.Get5OpenvpnCert(params)
	if err != nil {
		return "", err
	}
	return cert.Payload, nil
}

// SerializeConfig returns a single string containing a valid OpenVPN
// configuration file.
func (api *API) SerializeConfig(params *GatewayParams) (string, error) {
	rawCert, err := api.GetOpenVPNCert()
	if err != nil {
		return "", err
	}

	var key string
	if strings.Contains(rawCert, rsaBegin) {
		key = matchDelimitedString(rawCert, rsaBegin, rsaEnd)
	} else {
		key = matchDelimitedString(rawCert, keyBegin, keyEnd)
	}

	crt := matchDelimitedString(rawCert, certBegin, certEnd)
	gateways, err := api.GetGateways(params)
	if err != nil {
		return "", err
	}

	// TODO we can loop for a maximum of gateways
	gw := gateways[0]

	vars := configVars{
		CA:        riseupCA,
		Cert:      crt,
		Key:       key,
		IPAddr:    gw.IPAddr,
		Port:      fmt.Sprintf("%d", gw.Port),
		Transport: gw.Transport + "4",
	}
	tmpl, err := template.New("openvpncert").Parse(openvpnConfigTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, vars)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

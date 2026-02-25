package bootstrap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/proxy"

	"0xacab.org/leap/bitmask-core/pkg/introducer"
	"0xacab.org/leap/bitmask-core/pkg/storage"
	bitmask_storage "0xacab.org/leap/bitmask-core/pkg/storage"
	"0xacab.org/leap/menshen/client"
	"0xacab.org/leap/menshen/client/provisioning"
	"0xacab.org/leap/menshen/models"

	"0xacab.org/leap/tunnel-telemetry/pkg/geolocate"
)

type Config struct {
	// country code used to fetch gateways/bridges.
	CountryCode string
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
	// STUNServers is a list of STUN users to be used to get the current ip adress
	// The order is kept. A provider can use a list of public STUN servers, use
	// its self-hosted STUN servers or use public STUN servers as a fallback here.
	// A STUN server should be in the format ip/host:port
	STUNServers []string
	// The CountryCodeLookupURL returns a country code for a given ip address.
	CountryCodeLookupURL string
}

type API struct {
	client     *client.MenshenAPI
	httpClient *http.Client
	config     *Config
}

type ProviderSetup struct {
	Provider           *models.ModelsProvider
	Gateways           []*models.ModelsGateway
	Bridges            []*models.ModelsBridge
	Service            *models.ModelsEIPService
	OpenvpnCredentials string
}

func NewConfig() *Config {
	return &Config{
		Port:           443,
		UseTLS:         true,
		ResolveWithDoH: true,
	}
}

func NewConfigFromURL(url string) (*Config, error) {
	host, port, useTLS, err := parseApiURL(url)
	if err != nil {
		return nil, err

	}
	return &Config{
		Host:           host,
		Port:           port,
		UseTLS:         useTLS,
		ResolveWithDoH: useTLS,
	}, nil
}

func NewAPI(cfg *Config) (*API, error) {
	transportConfig := client.DefaultTransportConfig()

	var intro *introducer.Introducer
	var err error

	if cfg.Introducer != "" {
		intro, err = introducer.NewIntroducerFromURL(cfg.Introducer)
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
		client: client,
		config: cfg,
	}

	// Introducer has precedence over the Proxy parameter, unless it fails.
	// Above we've parsed the introducer URL, here we try to get an http client
	// configured to use it.
	// In the future, we might want to add a timeout and mark it as unusable if it fails.
	if intro != nil {
		client, err := introducer.NewHTTPClientFromIntroducer(intro)
		if err != nil {
			return nil, err
		}
		log.Info().
			Str("type", intro.Type).
			Str("addr", intro.Addr).
			Bool("UseKCP", intro.KCP).
			Msg("Using introducer")
		api.httpClient = client
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
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
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

// DoGeolocationLookup will try to fetch a valid country code.This country
// code will be stored and sent in any subsequent resource queries to menshen
// This method should be called only once, right after initializing the API object.
// The VPN must be turned off when calling this function. When the geolocation
// lookup succeeds, the country code is saved on disk for the future, in case
// the geolcation lookup fails.

func (api *API) DoGeolocationLookup() error {
	log.Debug().Msg("Doing geolocataion lookup")

	storage, err := bitmask_storage.GetStorage()
	if err != nil {
		return fmt.Errorf("Could not get storage to load/save country code fallback: %s", err)
	}

	geolocator := geolocate.NewGeolocator(api.config.CountryCodeLookupURL)
	geo, err := geolocator.FindCurrentHostGeolocationWithSTUN(api.config.STUNServers)

	if err == nil {
		api.config.CountryCode = geo.CC
		storage.SaveFallbackCountryCode(geo.CC)
		return nil
	}

	log.Warn().
		Err(err).
		Str("stunServers", strings.Join(api.config.STUNServers, ",")).
		Str("countryCodeLookupURL", api.config.CountryCodeLookupURL).
		Msg("Could not get country code using STUN servers. " +
			"Trying to use previously fetched country code")

	cc := storage.GetFallbackCountryCode()
	if cc == "" {
		log.Warn().Msg("No fallback country code was saved. Proceeding without country code")
	} else {
		log.Info().
			Str("countryCode", cc).
			Msg("Using fallback country code")
		api.config.CountryCode = cc
	}

	return nil
}

func (api *API) GetProvider() (*models.ModelsProvider, error) {
	params := provisioning.NewGetProviderJSONParams()
	if api.httpClient != nil {
		params = params.WithHTTPClient(api.httpClient)
	}
	providerResponse, err := api.client.Provisioning.GetProviderJSON(params)
	if err != nil {
		return nil, err
	}
	return providerResponse.Payload, nil
}

// call menshen endpoint /service and return response
// locations, openvpn arguments, serial+version, auth
func (api *API) GetService() (*models.ModelsEIPService, error) {
	params := provisioning.NewGetAPI5ServiceParams()
	if api.httpClient != nil {
		params = params.WithHTTPClient(api.httpClient)
	}

	service, err := api.client.Provisioning.GetAPI5Service(params)
	if err != nil {
		return nil, err
	}
	return service.Payload, nil
}

// RunBootstrap initializes the provider setup by sequentially running various update methods.
// It can also force updates based on the `force` parameter.
// The method optionally reports progress through a channel.
//
// Parameters:
//   - force (bool): If set to true, forces updates to be applied even if existing data may be valid.
//   - progress (chan int): A channel used to report progress as an integer percentage in case it's not nil.
//
// Returns:
//   - *ProviderSetup: A pointer to the populated ProviderSetup object.
//   - error: An error object, if any occurred during execution.
func (api *API) RunBootstrap(force bool, progress chan int) (*ProviderSetup, error) {
	providerSetup := &ProviderSetup{}
	maybeUpdate(progress, 0)
	if err := api.UpdateProvider(providerSetup, force); err != nil {
		return nil, err
	}
	maybeUpdate(progress, 10)

	err := api.DoGeolocationLookup()
	if err != nil {
		log.Warn().
			Str("err", err.Error()).
			Msgf("Could not do geolocation lookup")
	}
	maybeUpdate(progress, 25)

	if err = api.UpdateEIPService(providerSetup, force); err != nil {
		return nil, err
	}
	maybeUpdate(progress, 40)

	if err = api.UpdateGateways(providerSetup, force); err != nil {
		return nil, err
	}
	maybeUpdate(progress, 55)

	if err = api.UpdateBridges(providerSetup, force); err != nil {
		return nil, err
	}
	maybeUpdate(progress, 70)

	if err = api.UpdateCredentials(providerSetup, force); err != nil {
		return nil, err
	}
	maybeUpdate(progress, 85)

	if err = api.saveProviderSetup(providerSetup); err != nil {
		return nil, err
	}
	maybeUpdate(progress, 100)

	return providerSetup, nil
}

func maybeUpdate(progress chan int, value int) {
	if progress != nil {
		progress <- value
	}
}

// UpdateCredentials updates the OpenVPN credentials in the given ProviderSetup struct.
// It checks the current credentials against an expecteed certificate fingerprint to determine if an update
// is necessary, or just proceeds if forced.
//
// Parameters:
//   - providerSetup (*ProviderSetup): The provider setup instance containing current credentials
//     and associated data.
//   - force (bool): Forces an update of the OpenVPN credentials.
//
// Returns:
//   - error: An error object, if any occurred during the process.
func (api *API) UpdateCredentials(providerSetup *ProviderSetup, force bool) error {
	storage, err := storage.GetStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage")
	}

	caCertFingerprint := providerSetup.Provider.CaCertFingerprint
	credentials := storage.GetOpenVPNCredentials()
	if ShouldUpdateOpenVPNCredentials(credentials, caCertFingerprint) || force {
		credentials, err = api.GetOpenVPNCert()
		if err != nil {
			return fmt.Errorf("failed to fetch openvpn credentials %v", err)
		}
	}
	providerSetup.OpenvpnCredentials = credentials
	return nil
}

// UpdateProvider fetches and updates the provider details into the ProviderSetup.
// The method checks storage for existing data and determines the need for updates based on the
// force parameter.
//
// Parameters:
//   - providerSetup (*ProviderSetup): The provider setup instance where the provider data will be stored.
//   - force (bool): Forces an update of the provider details.
//
// Returns:
//   - error: An error object, if any occurred during the process.
func (api *API) UpdateProvider(providerSetup *ProviderSetup, force bool) error {
	storage, err := storage.GetStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage")
	}

	var provider *models.ModelsProvider
	err = json.Unmarshal([]byte(storage.GetModelsProvider()), &provider)
	if err != nil || force {
		provider, err = api.GetProvider()
		if err != nil {
			return fmt.Errorf("fetching provider failed %v", err)
		}
	}

	if !SupportsApiv5(provider) {
		return fmt.Errorf("this backend seems not to support APIv5")
	}

	providerSetup.Provider = provider
	return nil
}

// UpdateBridges updates the bridges in the ProviderSetup.
// The method checks the stored bridges and updates them if they are outdated or if a force update is requested.
//
// Parameters:
//   - providerSetup (*ProviderSetup): The provider setup instance containing bridge data.
//   - force (bool): Forces an update of the bridge details.
//
// Returns:
//   - error: An error object, if any occurred during the process.
func (api *API) UpdateBridges(providerSetup *ProviderSetup, force bool) error {
	storage, err := storage.GetStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage")
	}

	var bridges []*models.ModelsBridge
	err = json.Unmarshal([]byte(storage.GetModelsBridges()), &bridges)
	if err != nil || ShouldUpdateBridges(storage.GetBridgesTimestamp()) || force {
		bridgeParams := &BridgeParams{
			CC: storage.GetFallbackCountryCode(),
		}
		bridges, err = api.GetAllBridges(bridgeParams)
		if err != nil {
			return fmt.Errorf("failed to fetch gateways endpoint: %v", err)
		}
	}
	providerSetup.Bridges = bridges
	return nil
}

// UpdateGateways updates the gateways in the ProviderSetup.
// Checks stored gateways and updates them based on freshness or if a force update is required.
//
// Parameters:
//   - providerSetup (*ProviderSetup): The provider setup instance containing gateway data.
//   - force (bool): Forces an update of the gateway details.
//
// Returns:
//   - error: An error object, if any occurred during the process.
func (api *API) UpdateGateways(providerSetup *ProviderSetup, force bool) error {
	storage, err := storage.GetStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage")
	}

	var gateways []*models.ModelsGateway
	err = json.Unmarshal([]byte(storage.GetModelsGateways()), &gateways)
	if err != nil || ShouldUpdateGateways(storage.GetGatewaysTimestamp()) || force {
		gatewayParams := &GatewayParams{
			CC: storage.GetFallbackCountryCode(),
		}
		gateways, err = api.GetGateways(gatewayParams)
		if err != nil {
			return fmt.Errorf("failed to fetch gateways endpoint: %v", err)
		}
	}
	providerSetup.Gateways = gateways
	return nil
}

// UpdateEIPService fetches and updates the EIP service details in the ProviderSetup.
// The update checks the freshness of existing data against stored timestamps and the force parameter.
//
// Parameters:
//   - providerSetup (*ProviderSetup): The provider setup instance containing service data.
//   - force (bool): Forces an update of the service details.
//
// Returns:
//   - error: An error object, if any occurred during the process.
func (api *API) UpdateEIPService(providerSetup *ProviderSetup, force bool) error {
	storage, err := storage.GetStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage")
	}

	var service *models.ModelsEIPService
	err = json.Unmarshal([]byte(storage.GetModelsEIPService()), &service)
	if err != nil || ShouldUpdateEIPService(storage.GetEIPServiceTimestamp()) || force {
		service, err = api.GetService()
		if err != nil {
			return fmt.Errorf("failed to fetch service endpoint: %v", err)
		}
	}
	providerSetup.Service = service
	return nil
}

func (api *API) saveProviderSetup(providerSetup *ProviderSetup) error {
	storage, err := storage.GetStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage")
	}

	result, _ := ToJson(providerSetup.Provider)
	storage.SaveModelsProvider(result)

	result, _ = ToJson(providerSetup.Service)
	storage.SaveModelsEIPService(result)

	result, _ = ToJson(providerSetup.Bridges)
	storage.SaveModelsBridges(result)

	result, _ = ToJson(providerSetup.Gateways)
	storage.SaveModelsGateways(result)

	storage.SaveOpenVPNCredentials(providerSetup.OpenvpnCredentials)
	return nil
}

// GatewayParams contains the fields that can be used to filter the listing of available gateways.
type GatewayParams struct {
	Location  string
	Port      string
	Transport string
	CC        string
}

type BridgeParams struct {
	Location  string
	Port      string
	Transport string
	CC        string
	Type      string
}

// GetGateways returns a list of gateways (it it's enabled by the menshen
// API). It optionally accepts a GatewayParams object where you can set
// different filters.
func (api *API) GetGateways(p *GatewayParams) ([]*models.ModelsGateway, error) {
	params := provisioning.NewGetAPI5GatewaysParams()
	if p != nil {
		params.Loc = &p.Location
		params.Port = &p.Port
		params.Tr = &p.Transport
		params.Cc = &p.CC
	}

	if api.config.CountryCode != "" {
		log.Debug().
			Str("countryCode", api.config.CountryCode).
			Msg("Setting country code")
		params.Cc = &api.config.CountryCode
	}

	if api.httpClient != nil {
		params = params.WithHTTPClient(api.httpClient)
	}
	authHeader := api.getInviteTokenAuth()

	gateways, err := api.client.Provisioning.GetAPI5Gateways(params, authHeader)
	if err != nil {
		return nil, err
	}
	return gateways.Payload, err
}

func (api *API) GetAllBridges(p *BridgeParams) ([]*models.ModelsBridge, error) {
	params := provisioning.NewGetAPI5BridgesParams()
	if p != nil {
		params.Loc = &p.Location
		params.Port = &p.Port
		params.Tr = &p.Transport
		params.Type = &p.Type
		params.Cc = &p.CC
	}
	if api.httpClient != nil {
		params = params.WithHTTPClient(api.httpClient)
	}
	authHeader := api.getInviteTokenAuth()
	bridges, err := api.client.Provisioning.GetAPI5Bridges(params, authHeader)
	if err != nil {
		return nil, err
	}
	return bridges.Payload, nil
}

// GetOpenVPNCert returns valid OpenVPN client credentials (CA certificate, openvpn certificate and
// private key)
func (api *API) GetOpenVPNCert() (string, error) {
	params := provisioning.NewGetAPI5OpenvpnCertParams()
	if api.httpClient != nil {
		params = params.WithHTTPClient(api.httpClient)
	}

	cert, err := api.client.Provisioning.GetAPI5OpenvpnCert(params)
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

	key := GetKeyFromApi5OpenvpnCertResponse(rawCert)
	crt := GetCertFromApi5OpenvpnCertResponse(rawCert)
	ca := GetCAFromApi5OpenvpnCertResponse(rawCert)
	gateways, err := api.GetGateways(params)
	if err != nil {
		return "", err
	}

	// TODO we can loop for a maximum of gateways
	gw := gateways[0]

	vars := configVars{
		CA:        ca,
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

package menshen

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"0xacab.org/leap/bitmask-core/models"
	"0xacab.org/leap/bitmask-core/pkg/bootstrap"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/snowflake"

	"github.com/rs/zerolog/log"
)

type Menshen struct {
	apiConfig *bootstrap.Config // information about the menshen (host, port, tls, DoH)
	api       *bootstrap.API    // http client that exports functions to get gateways, certificate, OpenVPN parameters, ...
	// SnowflakeCh chan *snowflake.StatusEvent //TODO: Snowflake support
	// snowflakeProgress int
	// snowflake bool
	Gateways           []*models.ModelsGateway            // list of gateways offered by menshen
	gwsByLocation      map[string][]*models.ModelsGateway // map with gateways per location (Paris: [gw1, gw2, ...])
	gwLocations        []string                           // list of locations (Paris, Seattle, ...)
	userChoice         string                             // remote selection by the user in the GUI (empty string for automatic/best gateway/location, "Paris" for gateways located Paris)
	locationQualityMap map[string]float64                 // quality for each location (locationQualityMap["Paris"] = 0.4 (values beteen 0 and 1)

}

// Parses API URL of menshen (config.APIURL). Can be overwritten during runtime by setting env API_URL.
// Returns hostname/ip, port, useTLS of menshen
func parseApiURL() (string, int, bool) {
	if os.Getenv("API_URL") != "" {
		config.APIURL = os.Getenv("API_URL")
		log.Debug().
			Str("apiUrl", config.APIURL).
			Msg("Using API URL from env")
	}

	url, err := url.Parse(config.APIURL)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("apiUrl", config.APIURL).
			Msg("Could not parse API URL")
	}

	hostname := url.Hostname()
	useTLS := url.Scheme != "http"

	var port int
	if url.Port() == "" {
		port = 443
	} else {
		port, err = strconv.Atoi(url.Port())
		if err != nil {
			log.Fatal().
				Err(err).
				Str("port", url.Port()).
				Msg("Could not parse port to int")
		}
	}

	log.Trace().
		Bool("useTLS", useTLS).
		Str("hostname", hostname).
		Int("port", port).
		Msg("Parsed API URL")

	return hostname, port, useTLS
}

func New() (*Menshen, error) {
	hostname, port, useTLS := parseApiURL()
	// URL schema parsing can be confusing. Parsing an empty string results in empty hostname
	if hostname == "" {
		return nil, errors.New("Could not initialize menshen object. Hostname is empty")
	}
	cfg := bootstrap.NewConfig()
	cfg.Host = hostname
	cfg.Port = port
	cfg.UseTLS = useTLS

	api, err := bootstrap.NewAPI(cfg)
	if err != nil {
		return nil, err
	}

	m := &Menshen{
		apiConfig:          cfg,
		api:                api,
		Gateways:           []*models.ModelsGateway{},
		gwsByLocation:      make(map[string][]*models.ModelsGateway),
		gwLocations:        []string{},
		userChoice:         "",
		locationQualityMap: make(map[string]float64),
	}
	return m, nil
}

// Asks menshen for OpenVPN arguments
// Returns a list of arguments that can be passed over as command line arguments
// There are key-value arguments like "--dev tun" and boolean arguments like
// "--persisst-key" without additional value
// Currently, there is no caching implemented
func (m *Menshen) GetOpenvpnArgs() ([]string, error) {
	log.Trace().Msg("Getting OpenVPN arguments")

	service, err := m.api.GetService()
	if err != nil {
		return []string{}, err
	}

	// openVpnArgsInterface is of type interface{}
	openVpnArgsInterface := service.OpenvpnConfiguration

	// openVpnArgsArrayInterface is of type map[string]interface{}
	//   openVpnArgsArrayInterface["dev"] = "tun" (string)
	//   openVpnArgsArrayInterface["persist-key"] = true (bool)
	openVpnArgsArrayInterface, ok := openVpnArgsInterface.(map[string]interface{})
	if !ok {
		return []string{},
			fmt.Errorf("OpenVPN arguments returned by menshen have an invalid type: %v", openVpnArgsInterface)
	}
	openVpnArgs := []string{}
	for arg, value := range openVpnArgsArrayInterface {
		// e.g.: arg=dev value=tun, arg=persist-key value=true
		switch v := value.(type) {
		case string:
			if arg == "cipher" {
				arg = "data-ciphers"
			}
			// add "--dev tun" to openVpnArgs (v is "tun", but as string)
			openVpnArgs = append(openVpnArgs, "--"+arg)
			openVpnArgs = append(openVpnArgs, strings.Split(v, " ")...)
		case bool:
			if v {
				// just add --persist-key without value
				openVpnArgs = append(openVpnArgs, "--"+arg)
			}
		default:
			log.Warn().
				Str("argument", arg).
				Msgf("Unkown OpenVPN argument type (not bool/string): %v", value)
		}
	}
	log.Debug().
		Str("arguments", strings.Join(openVpnArgs, " ")).
		Msg("Got OpenVPN arguments from menshen")
	return openVpnArgs, nil
}

// Asks menshen for valid client credentials (certificate + key)
// Currently, there is no caching implemented
func (m *Menshen) GetPemCertificate() ([]byte, error) {
	log.Trace().Msg("Getting OpenVPN client certificate")
	cert, err := m.api.GetOpenVPNCert()
	if err != nil {
		return []byte{}, err
	}
	return []byte(cert), nil
}

// Returns true if at least one gateway supports udp
func (m *Menshen) IsUDPAvailable() bool {
	for _, gw := range m.Gateways {
		if gw.Type == "openvpn" {
			if gw.Transport == "udp" {
				return true
			}
		}
	}
	return false
}

func (m *Menshen) DoLogin(username, password string) (bool, error) {
	// TODO: implement if menshen supports auth
	return true, nil
}

func (m *Menshen) NeedsCredentials() bool {
	// TODO: implement if menshen supports auth
	return false
}

func (m *Menshen) GetSnowflakeCh() chan *snowflake.StatusEvent {
	return nil
}

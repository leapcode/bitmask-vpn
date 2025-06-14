package bonafide

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

type eipService struct {
	Gateways             []gatewayV3
	defaultGateway       string
	Locations            map[string]Location
	OpenvpnConfiguration openvpnConfig `json:"openvpn_configuration"`
	auth                 string
}

type eipServiceV1 struct {
	Gateways             []gatewayV1
	defaultGateway       string
	Locations            map[string]Location
	OpenvpnConfiguration openvpnConfig `json:"openvpn_configuration"`
}

type gatewayV1 struct {
	Capabilities struct {
		Ports     []string
		Protocols []string
	}
	Host      string
	IPAddress string `json:"ip_address"`
	Location  string
}

type gatewayV3 struct {
	Capabilities struct {
		Transport []transportV3
	}
	Host      string
	IPAddress string `json:"ip_address"`
	Location  string
}

type Location struct {
	CountryCode string `json:"country_code"`
	Hemisphere  string
	Name        string
	Timezone    string
}

type transportV3 struct {
	Type      string
	Protocols []string
	Ports     []string
	Options   map[string]string
}

func (b *Bonafide) setupAuthentication(i interface{}) {
	switch i.(type) {
	case eipService:
		switch auth := b.eip.auth; auth {
		case "anon":
			// Do nothing, we're set on initialization.
		case "sip":
			b.auth = &sipAuthentication{b.client, b.getURL("auth")}
		default:
			log.Warn().
				Str("auth", auth).
				Msg("Unknown authentication method")
		}
	case eipServiceV1:
		// Do nothing, no auth on v1.
	}
}

func (b *Bonafide) IsUDPAvailable() bool {
	if b.eip == nil {
		return false
	}
	for _, gw := range b.eip.Gateways {
		for _, t := range gw.Capabilities.Transport {
			if t.Type == "openvpn" {
				for _, proto := range t.Protocols {
					if proto == "udp" {
						return true
					}
				}
			}

		}

	}
	return false
}

func (b *Bonafide) fetchEipJSON() error {
	eip3API, err := url.JoinPath(config.ProviderConfig.APIURL, "3", "config", "eip-service.json")
	if err != nil {
		return err
	}
	log.Debug().Any("config.ProviderConfig", config.ProviderConfig)

	resp, err := b.client.Post(eip3API, "", nil)

	for err != nil {
		resp, err = b.client.Post(eip3API, "", nil)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not fetch eip v3 json")
			time.Sleep(retryFetchJSONSeconds * time.Second)
		}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		b.eip, err = decodeEIP3(resp.Body)
	case 404:
		buf := make([]byte, 128)
		resp.Body.Read(buf)
		log.Warn().Msg("Error fetching eip v3 json (status code 404)")
		eip1API := config.ProviderConfig.APIURL + "1/config/eip-service.json"
		resp, err = b.client.Post(eip1API, "", nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("Get eip json has failed with status: %s", resp.Status)
		}

		b.eip, err = decodeEIP1(resp.Body)
	default:
		return fmt.Errorf("Get eip json has failed with status: %s", resp.Status)
	}
	if err != nil {
		return err
	}

	b.setupAuthentication(b.eip)
	return nil
}

func (b *Bonafide) parseEipJSONFromFile() error {
	provider := strings.ToLower(config.ProviderConfig.Provider)
	eipFile := filepath.Join(config.Path, provider+"-eip.json")
	f, err := os.Open(eipFile)
	if err != nil {
		return err
	}
	b.eip, err = decodeEIP3(f)
	return err
}

func decodeEIP3(body io.Reader) (*eipService, error) {
	var eip eipService
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&eip)
	return &eip, err
}

func decodeEIP1(body io.Reader) (*eipService, error) {
	var eip1 eipServiceV1
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&eip1)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not fetching eip v1 json")
		return nil, err
	}

	eip3 := eipService{
		Gateways:             make([]gatewayV3, len(eip1.Gateways)),
		Locations:            eip1.Locations,
		OpenvpnConfiguration: eip1.OpenvpnConfiguration,
	}
	for _, g := range eip1.Gateways {
		gateway := gatewayV3{
			Host:      g.Host,
			IPAddress: g.IPAddress,
			Location:  g.Location,
		}
		gateway.Capabilities.Transport = []transportV3{
			{
				Type:      "openvpn",
				Ports:     g.Capabilities.Ports,
				Protocols: g.Capabilities.Protocols,
			},
		}
		eip3.Gateways = append(eip3.Gateways, gateway)
	}
	return &eip3, nil
}

func (eip eipService) getGateways() []Gateway {
	gws := []Gateway{}
	for _, g := range eip.Gateways {
		for _, t := range g.Capabilities.Transport {
			gateway := Gateway{
				Host:         g.Host,
				IPAddress:    g.IPAddress,
				Location:     g.Location,
				Ports:        t.Ports,
				Protocols:    t.Protocols,
				Options:      t.Options,
				Transport:    t.Type,
				LocationName: eip.Locations[g.Location].Name,
				CountryCode:  eip.Locations[g.Location].CountryCode,
			}
			gws = append(gws, gateway)
		}
	}
	return gws
}

func (eip eipService) getOpenvpnArgs() []string {
	args := []string{}
	var cfg = eip.OpenvpnConfiguration

	// for debug purposes, we allow parsing an extra block of openvpn configurations.
	if openvpnExtra := os.Getenv("LEAP_OPENVPN_EXTRA_CONFIG"); openvpnExtra != "" {
		extraConfig, err := parseOpenvpnArgsFromFile(openvpnExtra)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not parse extra config:")
		} else {
			cfg = *extraConfig
		}
	}

	for arg, value := range cfg {
		switch v := value.(type) {
		case string:
			// this is a transitioning hack for the transition to float deployment,
			// assuming we're using openvpn 2.5. We're treating the "cipher"
			// string that the platform sends us as the newer data-cipher
			// which includes colon-separated ciphers.
			if arg == "cipher" {
				arg = "data-ciphers"
			}
			args = append(args, "--"+arg)
			args = append(args, strings.Split(v, " ")...)
		case bool:
			if v {
				args = append(args, "--"+arg)
			}
		default:
			log.Warn().
				Str("arg", arg).
				Msgf("Unknown openvpn argument type (value=%v)", value)
		}
	}
	return args
}

func parseOpenvpnArgsFromFile(path string) (*openvpnConfig, error) {
	// TODO sanitize options: check keys against array of allowed options
	f, err := os.Open(path)
	defer f.Close()

	if err != nil {
		return nil, err
	}
	byteValue, _ := ioutil.ReadAll(f)
	var cfg openvpnConfig
	json.Unmarshal([]byte(byteValue), &cfg)
	return &cfg, nil
}

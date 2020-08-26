package bonafide

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

type eipService struct {
	Gateways             []gatewayV3
	defaultGateway       string
	Locations            map[string]location
	OpenvpnConfiguration openvpnConfig `json:"openvpn_configuration"`
	auth                 string
}

type eipServiceV1 struct {
	Gateways             []gatewayV1
	defaultGateway       string
	Locations            map[string]location
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

type location struct {
	CountryCode string
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
			log.Printf("BUG: unknown authentication method %s", auth)
		}
	case eipServiceV1:
		// Do nothing, no auth on v1.
	}
}

func (b *Bonafide) fetchEipJSON() error {
	eip3API := config.APIURL + "3/config/eip-service.json"
	resp, err := b.client.Post(eip3API, "", nil)
	for err != nil {
		log.Printf("Error fetching eip v3 json: %v", err)
		// TODO why exactly 1 retry? Make it configurable, for tests
		time.Sleep(retryFetchJSONSeconds * time.Second)
		resp, err = b.client.Post(eip3API, "", nil)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		b.eip, err = decodeEIP3(resp.Body)
	case 404:
		buf := make([]byte, 128)
		resp.Body.Read(buf)
		log.Printf("Error fetching eip v3 json")
		eip1API := config.APIURL + "1/config/eip-service.json"
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
		log.Printf("Error fetching eip v1 json: %v", err)
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
			transportV3{
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
				Host:      g.Host,
				IPAddress: g.IPAddress,
				Location:  g.Location,
				Ports:     t.Ports,
				Protocols: t.Protocols,
				Options:   t.Options,
				Transport: t.Type,
			}
			gws = append(gws, gateway)
		}
	}
	return gws
}

func (eip eipService) getOpenvpnArgs() []string {
	args := []string{}
	for arg, value := range eip.OpenvpnConfiguration {
		switch v := value.(type) {
		case string:
			args = append(args, "--"+arg)
			args = append(args, strings.Split(v, " ")...)
		case bool:
			if v {
				args = append(args, "--"+arg)
			}
		default:
			log.Printf("Unknown openvpn argument type: %s - %v", arg, value)
		}
	}
	return args
}

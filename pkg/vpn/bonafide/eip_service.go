package bonafide

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

type eipService struct {
	Gateways             []gatewayV3
	SelectedGateways     []gatewayV3
	Locations            map[string]location
	defaultGateway       string
	OpenvpnConfiguration openvpnConfig `json:"openvpn_configuration"`
	auth                 string
}

type eipServiceV1 struct {
	Gateways             []gatewayV1
	SelectedGateways     []gatewayV1
	Locations            map[string]location
	OpenvpnConfiguration openvpnConfig `json:"openvpn_configuration"`
}

type location struct {
	CountryCode string
	Hemisphere  string
	Name        string
	Timezone    string
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
	/* TODO we could launch the looping call from here.
	but smells: calls a bonafide method that in turn calls methods in this file
	*/
	b.sortGateways()
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

func (eip eipService) getGateways(transport string) []Gateway {
	gws := []Gateway{}
	// TODO check that len(selected) != 0
	for _, g := range eip.SelectedGateways {
		for _, t := range g.Capabilities.Transport {
			if t.Type != transport {
				continue
			}

			gateway := Gateway{
				Host:      g.Host,
				IPAddress: g.IPAddress,
				Location:  g.Location,
				Ports:     t.Ports,
				Protocols: t.Protocols,
				Options:   t.Options,
			}
			gws = append(gws, gateway)
		}
	}
	// TODO return only top 3, at least for openvpn
	return gws
}

func (eip *eipService) setManualGateway(name string) {
	eip.defaultGateway = name

	gws := make([]gatewayV3, 0)
	for _, gw := range eip.Gateways {
		if gw.Location == eip.defaultGateway {
			gws = append(gws, gw)
			break
		}
	}
	eip.SelectedGateways = gws
}

func (eip *eipService) autoSortGateways(serviceSelection []string) {
	gws := make([]gatewayV3, 0)

	for _, host := range serviceSelection {
		for _, gw := range eip.Gateways {
			if gw.Host == host {
				gws = append(gws, gw)
			}
		}
	}

	if len(gws) == 0 {
		// this can happen if a misconfigured geoip service does not match the
		// providers list we got.
		log.Println("ERROR: did not get any useful selection. Is the geolocation service properly configured?")
		eip.SelectedGateways = eip.Gateways
	} else {
		eip.SelectedGateways = gws
	}
}

func (eip *eipService) sortGatewaysByTimezone(tzOffsetHours int) {
	gws := []gatewayDistance{}

	for _, gw := range eip.Gateways {
		distance := 13
		if gw.Location == eip.defaultGateway {
			distance = -1
		} else {
			gwOffset, err := strconv.Atoi(eip.Locations[gw.Location].Timezone)
			if err != nil {
				log.Printf("Error sorting gateways: %v", err)
			} else {
				distance = tzDistance(tzOffsetHours, gwOffset)
			}
		}
		gws = append(gws, gatewayDistance{gw, distance})
	}
	rand.Seed(time.Now().UnixNano())
	cmp := func(i, j int) bool {
		if gws[i].distance == gws[j].distance {
			return rand.Intn(2) == 1
		}
		return gws[i].distance < gws[j].distance
	}
	sort.Slice(gws, cmp)

	eip.SelectedGateways = make([]gatewayV3, len(eip.Gateways))
	for i, gw := range gws {
		eip.SelectedGateways[i] = gw.gateway
	}
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

type gatewayDistance struct {
	gateway  gatewayV3
	distance int
}

func tzDistance(offset1, offset2 int) int {
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	distance := abs(offset1 - offset2)
	if distance > 12 {
		distance = 24 - distance
	}
	return distance
}

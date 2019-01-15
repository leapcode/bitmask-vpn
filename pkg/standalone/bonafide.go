// Copyright (C) 2018 LEAP
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package standalone

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"0xacab.org/leap/bitmask-systray/pkg/config"
)

const (
	certAPI        = config.APIURL + "1/cert"
	eipAPI         = config.APIURL + "1/config/eip-service.json"
	secondsPerHour = 60 * 60
)

type bonafide struct {
	client         httpClient
	tzOffsetHours  int
	eip            *eipService
	defaultGateway string
}

type httpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type eipService struct {
	Gateways  []gateway
	Locations map[string]struct {
		CountryCode string
		Hemisphere  string
		Name        string
		Timezone    string
	}
	OpenvpnConfiguration map[string]interface{} `json:"openvpn_configuration"`
}

type gateway struct {
	Capabilities struct {
		Ports     []string
		Protocols []string
	}
	Host      string
	IPAddress string `json:"ip_address"`
	Location  string
}

type gatewayDistance struct {
	gateway  gateway
	distance int
}

type geoLocation struct {
	IPAddress      string   `json:"ip"`
	Country        string   `json:"cc"`
	City           string   `json:"city"`
	Latitude       float64  `json:"lat"`
	Longitude      float64  `json:"lon"`
	SortedGateways []string `json:"gateways"`
}

func newBonafide() *bonafide {
	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM(config.CaCert)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
	}
	_, tzOffsetSeconds := time.Now().Zone()
	tzOffsetHours := tzOffsetSeconds / secondsPerHour

	return &bonafide{
		client:         client,
		tzOffsetHours:  tzOffsetHours,
		eip:            nil,
		defaultGateway: "",
	}
}

func (b *bonafide) getCertPem() ([]byte, error) {
	resp, err := b.client.Post(certAPI, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get vpn cert has failed with status: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func (b *bonafide) getGateways() ([]gateway, error) {
	if b.eip == nil {
		err := b.fetchEipJSON()
		if err != nil {
			return nil, err
		}
	}

	return b.eip.Gateways, nil
}

func (b *bonafide) setDefaultGateway(name string) {
	b.defaultGateway = name
	b.sortGateways()
}

func (b *bonafide) getOpenvpnArgs() ([]string, error) {
	if b.eip == nil {
		err := b.fetchEipJSON()
		if err != nil {
			return nil, err
		}
	}

	args := []string{}
	for arg, value := range b.eip.OpenvpnConfiguration {
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
	return args, nil
}

func (b *bonafide) fetchGeolocation() ([]string, error) {
	resp, err := b.client.Post(config.GeolocationAPI, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get geolocation failed with status: %s", resp.Status)
	}

	geo := &geoLocation{}
	dataJSON, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(dataJSON, &geo)
	if err != nil {
		_ = fmt.Errorf("get vpn cert has failed with status: %s", resp.Status)
		return nil, err
	}

	return geo.SortedGateways, nil

}

func (b *bonafide) fetchEipJSON() error {
	resp, err := b.client.Post(eipAPI, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("get eip json has failed with status: %s", resp.Status)
	}

	var eip eipService
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&eip)
	if err != nil {
		return err
	}

	b.eip = &eip
	b.sortGateways()
	return nil
}

func (b *bonafide) sortGatewaysByGeolocation(geolocatedGateways []string) []gatewayDistance {
	gws := []gatewayDistance{}

	for i, host := range geolocatedGateways {
		for _, gw := range b.eip.Gateways {
			if gw.Host == host {
				gws = append(gws, gatewayDistance{gw, i})
			}
		}
	}
	return gws
}

func (b *bonafide) sortGatewaysByTimezone() []gatewayDistance {
	gws := []gatewayDistance{}

	for _, gw := range b.eip.Gateways {
		distance := 13
		if gw.Location == b.defaultGateway {
			distance = -1
		} else {
			gwOffset, err := strconv.Atoi(b.eip.Locations[gw.Location].Timezone)
			if err != nil {
				log.Printf("Error sorting gateways: %v", err)
			} else {
				distance = tzDistance(b.tzOffsetHours, gwOffset)
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
	return gws
}

func (b *bonafide) sortGateways() {
	gws := []gatewayDistance{}

	geolocatedGateways, _ := b.fetchGeolocation()

	if len(geolocatedGateways) > 0 {
		gws = b.sortGatewaysByGeolocation(geolocatedGateways)
	} else {
		log.Printf("Falling back to timezone heuristic for gateway selection")
		gws = b.sortGatewaysByTimezone()
	}

	for i, gw := range gws {
		b.eip.Gateways[i] = gw.gateway
	}
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

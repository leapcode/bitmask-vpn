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

package bitmask

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
)

const (
	certAPI        = "https://api.black.riseup.net/1/cert"
	eipAPI         = "https://api.black.riseup.net/1/config/eip-service.json"
	geolocationAPI = "https://api.black.riseup.net:9001/json"
	secondsPerHour = 60 * 60
)

var (
	caCert = []byte(`-----BEGIN CERTIFICATE-----
MIIFjTCCA3WgAwIBAgIBATANBgkqhkiG9w0BAQ0FADBZMRgwFgYDVQQKDA9SaXNl
dXAgTmV0d29ya3MxGzAZBgNVBAsMEmh0dHBzOi8vcmlzZXVwLm5ldDEgMB4GA1UE
AwwXUmlzZXVwIE5ldHdvcmtzIFJvb3QgQ0EwHhcNMTQwNDI4MDAwMDAwWhcNMjQw
NDI4MDAwMDAwWjBZMRgwFgYDVQQKDA9SaXNldXAgTmV0d29ya3MxGzAZBgNVBAsM
Emh0dHBzOi8vcmlzZXVwLm5ldDEgMB4GA1UEAwwXUmlzZXVwIE5ldHdvcmtzIFJv
b3QgQ0EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC76J4ciMJ8Sg0m
TP7DF2DT9zNe0Csk4myoMFC57rfJeqsAlJCv1XMzBmXrw8wq/9z7XHv6n/0sWU7a
7cF2hLR33ktjwODlx7vorU39/lXLndo492ZBhXQtG1INMShyv+nlmzO6GT7ESfNE
LliFitEzwIegpMqxCIHXFuobGSCWF4N0qLHkq/SYUMoOJ96O3hmPSl1kFDRMtWXY
iw1SEKjUvpyDJpVs3NGxeLCaA7bAWhDY5s5Yb2fA1o8ICAqhowurowJpW7n5ZuLK
5VNTlNy6nZpkjt1QycYvNycffyPOFm/Q/RKDlvnorJIrihPkyniV3YY5cGgP+Qkx
HUOT0uLA6LHtzfiyaOqkXwc4b0ZcQD5Vbf6Prd20Ppt6ei0zazkUPwxld3hgyw58
m/4UIjG3PInWTNf293GngK2Bnz8Qx9e/6TueMSAn/3JBLem56E0WtmbLVjvko+LF
PM5xA+m0BmuSJtrD1MUCXMhqYTtiOvgLBlUm5zkNxALzG+cXB28k6XikXt6MRG7q
hzIPG38zwkooM55yy5i1YfcIi5NjMH6A+t4IJxxwb67MSb6UFOwg5kFokdONZcwj
shczHdG9gLKSBIvrKa03Nd3W2dF9hMbRu//STcQxOailDBQCnXXfAATj9pYzdY4k
ha8VCAREGAKTDAex9oXf1yRuktES4QIDAQABo2AwXjAdBgNVHQ4EFgQUC4tdmLVu
f9hwfK4AGliaet5KkcgwDgYDVR0PAQH/BAQDAgIEMAwGA1UdEwQFMAMBAf8wHwYD
VR0jBBgwFoAUC4tdmLVuf9hwfK4AGliaet5KkcgwDQYJKoZIhvcNAQENBQADggIB
AGzL+GRnYu99zFoy0bXJKOGCF5XUXP/3gIXPRDqQf5g7Cu/jYMID9dB3No4Zmf7v
qHjiSXiS8jx1j/6/Luk6PpFbT7QYm4QLs1f4BlfZOti2KE8r7KRDPIecUsUXW6P/
3GJAVYH/+7OjA39za9AieM7+H5BELGccGrM5wfl7JeEz8in+V2ZWDzHQO4hMkiTQ
4ZckuaL201F68YpiItBNnJ9N5nHr1MRiGyApHmLXY/wvlrOpclh95qn+lG6/2jk7
3AmihLOKYMlPwPakJg4PYczm3icFLgTpjV5sq2md9bRyAg3oPGfAuWHmKj2Ikqch
Td5CHKGxEEWbGUWEMP0s1A/JHWiCbDigc4Cfxhy56CWG4q0tYtnc2GMw8OAUO6Wf
Xu5pYKNkzKSEtT/MrNJt44tTZWbKV/Pi/N2Fx36my7TgTUj7g3xcE9eF4JV2H/sg
tsK3pwE0FEqGnT4qMFbixQmc8bGyuakr23wjMvfO7eZUxBuWYR2SkcP26sozF9PF
tGhbZHQVGZUTVPyvwahMUEhbPGVerOW0IYpxkm0x/eaWdTc4vPpf/rIlgbAjarnJ
UN9SaWRlWKSdP4haujnzCoJbM7dU9bjvlGZNyXEekgeT0W2qFeGGp+yyUWw8tNsp
0BuC1b7uW/bBn/xKm319wXVDvBgZgcktMolak39V7DVO
-----END CERTIFICATE-----`)
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
	certs.AppendCertsFromPEM(caCert)
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
	resp, err := b.client.Post(geolocationAPI, "", nil)
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

// SPDX-FileCopyrightText: 2018 LEAP
// SPDX-License-Identifier: GPL-3.0-or-later
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

package bonafide

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

const (
	certAPI               = config.APIURL + "1/cert"
	certAPI3              = config.APIURL + "3/cert"
	secondsPerHour        = 60 * 60
	retryFetchJSONSeconds = 15
)

type Bonafide struct {
	client        httpClient
	eip           *eipService
	tzOffsetHours int
}

type Gateway struct {
	Host      string
	IPAddress string
	Location  string
	Ports     []string
	Protocols []string
	Options   map[string]string
}

type openvpnConfig map[string]interface{}

type httpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

type geoLocation struct {
	IPAddress      string   `json:"ip"`
	Country        string   `json:"cc"`
	City           string   `json:"city"`
	Latitude       float64  `json:"lat"`
	Longitude      float64  `json:"lon"`
	SortedGateways []string `json:"gateways"`
}

func New() *Bonafide {
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

	return &Bonafide{
		client:        client,
		eip:           nil,
		tzOffsetHours: tzOffsetHours,
	}
}

func (b *Bonafide) GetCertPem() ([]byte, error) {
	resp, err := b.client.Post(certAPI, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		resp, err = b.client.Post(certAPI3, "", nil)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get vpn cert has failed with status: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func (b *Bonafide) GetGateways(transport string) ([]Gateway, error) {
	if b.eip == nil {
		err := b.fetchEipJSON()
		if err != nil {
			return nil, err
		}
	}

	return b.eip.getGateways(transport), nil
}

func (b *Bonafide) SetDefaultGateway(name string) {
	b.eip.setDefaultGateway(name)
	b.sortGateways()
}

func (b *Bonafide) GetOpenvpnArgs() ([]string, error) {
	if b.eip == nil {
		err := b.fetchEipJSON()
		if err != nil {
			return nil, err
		}
	}
	return b.eip.getOpenvpnArgs(), nil
}

func (b *Bonafide) fetchGeolocation() ([]string, error) {
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

func (b *Bonafide) sortGateways() {
	geolocatedGateways, _ := b.fetchGeolocation()

	if len(geolocatedGateways) > 0 {
		b.eip.sortGatewaysByGeolocation(geolocatedGateways)
	} else {
		log.Printf("Falling back to timezone heuristic for gateway selection")
		b.eip.sortGatewaysByTimezone(b.tzOffsetHours)
	}
}

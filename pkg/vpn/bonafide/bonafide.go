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
	secondsPerHour        = 60 * 60
	retryFetchJSONSeconds = 15
)

const (
	certPathv1 = "1/cert"
	certPathv3 = "3/cert"
	authPathv3 = "3/auth"

	certAPI  = config.APIURL + certPathv1
	certAPI3 = config.APIURL + certPathv3
	authAPI  = config.APIURL + authPathv3
)

type Bonafide struct {
	client        httpClient
	eip           *eipService
	tzOffsetHours int
	auth          Authentication
	credentials   *Credentials
	apiURL        string
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
	Do(req *http.Request) (*http.Response, error)
}

// The Authentication interface allows to get a Certificate in Pem format.
// We implement Anonymous Authentication (Riseup et al), and Sip (Libraries).
type Authentication interface {
	GetPemCertificate() ([]byte, error)
}

type geoLocation struct {
	IPAddress      string   `json:"ip"`
	Country        string   `json:"cc"`
	City           string   `json:"city"`
	Latitude       float64  `json:"lat"`
	Longitude      float64  `json:"lon"`
	SortedGateways []string `json:"gateways"`
}

// New Bonafide: Initializes a Bonafide object. By default, no Credentials are passed.
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

	b := &Bonafide{
		client:        client,
		eip:           nil,
		tzOffsetHours: tzOffsetHours,
	}
	auth := AnonymousAuthentication{b}
	b.auth = &auth
	return b
}

func (b *Bonafide) SetCredentials(username, password string) {
	b.credentials = &Credentials{username, password}
}

func (b *Bonafide) GetURL(object string) (string, error) {
	if b.apiURL == "" {
		switch object {
		case "cert":
			return certAPI, nil
		case "certv3":
			return certAPI3, nil
		case "auth":
			return authAPI, nil
		}
	} else {
		switch object {
		case "cert":
			return b.apiURL + certPathv1, nil
		case "certv3":
			return b.apiURL + certPathv3, nil
		case "auth":
			return b.apiURL + authPathv3, nil
		}
	}
	return "", fmt.Errorf("ERROR: unknown object for api url")
}

func (b *Bonafide) GetPemCertificate() ([]byte, error) {
	if b.auth == nil {
		log.Fatal("ERROR: bonafide did not initialize auth")
	}
	cert, err := b.auth.GetPemCertificate()
	return cert, err
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

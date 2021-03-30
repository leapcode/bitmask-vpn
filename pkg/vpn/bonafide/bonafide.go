// Copyright (C) 2018-2020 LEAP
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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
)

type Bonafide struct {
	client        httpClient
	eip           *eipService
	tzOffsetHours int
	gateways      *gatewayPool
	maxGateways   int
	auth          authentication
	token         []byte
}

type openvpnConfig map[string]interface{}

type httpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

type geoGateway struct {
	Host     string  `json:"host"`
	Fullness float64 `json:"fullness"`
	Overload bool    `json:"overload"`
}

type geoLocation struct {
	IPAddress      string       `json:"ip"`
	Country        string       `json:"cc"`
	City           string       `json:"city"`
	Latitude       float64      `json:"lat"`
	Longitude      float64      `json:"lon"`
	Gateways       []string     `json:"gateways"`
	SortedGateways []geoGateway `json:"sortedGateways"`
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
		Timeout: time.Second * 30,
	}
	_, tzOffsetSeconds := time.Now().Zone()
	tzOffsetHours := tzOffsetSeconds / secondsPerHour

	b := &Bonafide{
		client:        client,
		eip:           nil,
		tzOffsetHours: tzOffsetHours,
	}
	switch auth := config.Auth; auth {
	case "sip":
		log.Println("Client expects sip auth")
		b.auth = &sipAuthentication{client, b.getURL("auth")}
	case "anon":
		log.Println("Client expects anon auth")
		b.auth = &anonymousAuthentication{}
	default:
		log.Println("Client expects invalid auth", auth)
		b.auth = &anonymousAuthentication{}
	}

	return b
}

/* NeedsCredentials signals if we have to ask user for credentials. If false, it can be that we have a cached token */
func (b *Bonafide) NeedsCredentials() bool {
	if !b.auth.needsCredentials() {
		return false
	}
	/* try cached */
	/* TODO cleanup this call - maybe expose getCachedToken instead of relying on empty creds? */
	_, err := b.auth.getToken("", "")
	if err != nil {
		return true
	}
	return false
}

func (b *Bonafide) DoLogin(username, password string) (bool, error) {
	if !b.auth.needsCredentials() {
		return false, errors.New("Auth method does not need login")
	}

	var err error

	log.Println("Bonafide: getting token...")
	b.token, err = b.auth.getToken(username, password)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (b *Bonafide) GetPemCertificate() ([]byte, error) {
	if b.auth == nil {
		log.Fatal("ERROR: bonafide did not initialize auth")
	}
	if b.auth.needsCredentials() {
		/* try cached token */
		token, err := b.auth.getToken("", "")
		if err != nil {
			return nil, errors.New("bug: this service needs login, but we were not logged in")
		}
		b.token = token
	}

	req, err := http.NewRequest("POST", b.getURL("certv3"), strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	if b.token != nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", b.token))
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		resp, err = b.client.Post(b.getURL("cert"), "", nil)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Get vpn cert has failed with status: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func (b *Bonafide) getURL(object string) string {
	switch object {
	case "cert":
		return config.APIURL + certPathv1
	case "certv3":
		return config.APIURL + certPathv3
	case "auth":
		return config.APIURL + authPathv3
	}
	log.Println("BUG: unknown url object")
	return ""
}

func (b *Bonafide) maybeInitializeEIP() error {
	if b.eip == nil {
		err := b.fetchEipJSON()
		if err != nil {
			return err
		}
		b.gateways = newGatewayPool(b.eip)
	}

	// FIXME: let's update the menshen gateways every time we 'maybe initilize EIP'
	//        in a future we might want to be more clever on when to do that
	//        (when opening the locations tab in the UI, only on reconnects, ...)
	b.fetchGatewaysFromMenshen()
	return nil
}

// GetGateways filters by transport, and will return the maximum number defined
// in bonafide.maxGateways, or the maximum by default (3).
func (b *Bonafide) GetGateways(transport string) ([]Gateway, error) {
	err := b.maybeInitializeEIP()
	if err != nil {
		return nil, err
	}
	max := maxGateways
	if b.maxGateways != 0 {
		max = b.maxGateways
	}

	gws, err := b.gateways.getBest(transport, b.tzOffsetHours, max)
	return gws, err
}

// GetAllGateways only filters gateways by transport.
// TODO could pass "any" instead?
func (b *Bonafide) GetAllGateways(transport string) ([]Gateway, error) {
	err := b.maybeInitializeEIP()
	if err != nil {
		return nil, err
	}
	gws, err := b.gateways.getAll(transport, b.tzOffsetHours)
	return gws, err
}

func (b *Bonafide) ListLocationFullness(transport string) map[string]float64 {
	return b.gateways.listLocationFullness(transport)
}

func (b *Bonafide) SetManualGateway(label string) {
	b.gateways.setUserChoice(label)
}

func (b *Bonafide) SetAutomaticGateway() {
	b.gateways.setAutomaticChoice()
}

func (b *Bonafide) IsManualLocation() bool {
	if b.gateways == nil {
		return false
	}
	return b.gateways.isManualLocation()
}

func (b *Bonafide) GetGatewayByIP(ip string) (Gateway, error) {
	return b.gateways.getGatewayByIP(ip)
}

func (b *Bonafide) fetchGatewaysFromMenshen() error {
	/* FIXME in float deployments, geolocation is served on gemyip.domain/json, with a LE certificate, but in riseup is served behind the api certificate.
	So this is a workaround until we streamline that behavior */
	resp, err := b.client.Post(config.GeolocationAPI, "", nil)
	if err != nil {
		client := &http.Client{}
		_resp, err := client.Post(config.GeolocationAPI, "", nil)
		if err != nil {
			log.Printf("ERROR: could not fetch geolocation: %s\n", err)
			return err
		}
		resp = _resp
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("ERROR: bad status code while fetching geolocation:", resp.StatusCode)
		return fmt.Errorf("Get geolocation failed with status: %d", resp.StatusCode)
	}

	geo := &geoLocation{}
	dataJSON, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(dataJSON, &geo)
	if err != nil {
		log.Printf("ERROR: cannot parse geolocation json: %s\n", err)
		log.Println(string(dataJSON))
		_ = fmt.Errorf("bad json")
		return err
	}

	log.Println("Got sorted gateways:", geo.Gateways)
	b.gateways.setRecommendedGateways(geo)
	return nil
}

func (b *Bonafide) GetOpenvpnArgs() ([]string, error) {
	err := b.maybeInitializeEIP()
	if err != nil {
		return nil, err
	}
	return b.eip.getOpenvpnArgs(), nil
}

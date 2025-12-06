// Copyright (C) 2018-2021 LEAP
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
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-core/pkg/introducer"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/snowflake"
)

const (
	secondsPerHour        = 60 * 60
	retryFetchJSONSeconds = 15
	winUserAgent          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36 Edg/95.0.1020.53"
)

const (
	certPathv1 = "1/cert"
	certPathv3 = "3/cert"
	authPathv3 = "3/auth"
)

type Bonafide struct {
	client            httpClient
	eip               *eipService
	tzOffsetHours     int
	gateways          *gatewayPool
	maxGateways       int
	auth              authentication
	token             []byte
	snowflakeCh       chan *snowflake.StatusEvent // only used by the GUI to show the progress (but does not work?)
	snowflakeProgress int
	snowflake         bool
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
	certs, err := x509.SystemCertPool()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Error loading SystemCertPool, falling back to empty pool")
		certs = x509.NewCertPool()
	}
	certs.AppendCertsFromPEM(config.ProviderConfig.CaCert)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
		Timeout: time.Second * 30,
	}

	// experimental introducer
	if introURL := os.Getenv("LEAP_INTRODUCER_URL"); introURL != "" {
		inr, err := introducer.NewIntroducerFromURL(introURL)
		if err != nil {
			log.Debug().
				Err(err).
				Str("introducer URL", introURL).
				Msg("failed to create introducer from URL")
		}

		client, err = introducer.NewHTTPClientFromIntroducer(inr)
		if err != nil {
			log.Debug().
				Err(err).
				Msg("failed to create http client from introducer")
		}
	}
	client.Timeout = time.Minute
	_, tzOffsetSeconds := time.Now().Zone()
	tzOffsetHours := tzOffsetSeconds / secondsPerHour

	b := &Bonafide{
		client:        client,
		eip:           nil,
		tzOffsetHours: tzOffsetHours,
		snowflakeCh:   make(chan *snowflake.StatusEvent, 20),
	}
	switch auth := config.ProviderConfig.Auth; auth {
	case "sip":
		log.Debug().Msg("Client expects sip auth")
		b.auth = &sipAuthentication{client, b.getURL("auth")}
	case "anon":
		log.Debug().Msg("Client expects anon auth")
		b.auth = &anonymousAuthentication{}
	default:
		log.Debug().
			Str("auth", auth).
			Msg("Client expects invalid auth")
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

	log.Debug().Msg("Bonafide: getting token...")
	b.token, err = b.auth.getToken(username, password)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (b *Bonafide) GetPemCertificate() ([]byte, error) {
	if b.auth == nil {
		log.Fatal().Msg("ERROR: bonafide did not initialize auth")
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
	if runtime.GOOS == "windows" {
		req.Header.Add("User-Agent", winUserAgent)
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

	return io.ReadAll(resp.Body)
}

func (b *Bonafide) getURL(object string) string {
	switch object {
	case "cert":
		if link, err := url.JoinPath(config.ProviderConfig.APIURL, certPathv1); err == nil {
			return link
		}
		return ""
	case "certv3":
		u, err := url.Parse(config.ProviderConfig.APIURL)
		if err != nil {
			log.Debug().
				Err(err)
			return ""
		}
		apiUrl, err := url.Parse(u.Hostname())
		if err != nil {
			log.Debug().
				Err(err)
			return ""
		}
		apiUrl.Scheme = u.Scheme
		if link, err := url.JoinPath(apiUrl.String(), certPathv3); err == nil {
			log.Debug().
				Str("cert_url", link).
				Msg("v3 openvpn cert url")
			return link
		}
		return ""
	case "auth":
		if link, err := url.JoinPath(config.ProviderConfig.APIURL, authPathv3); err == nil {
			return link
		}
		return ""
	}
	log.Warn().Msg("BUG: unknown url object")
	return ""
}

func (b *Bonafide) watchSnowflakeProgress(ch chan *snowflake.StatusEvent) {
	// We need to keep track of the bootstrap process here, and then we
	// pass to the channel that is observed by the backend
	log.Debug().Msg("Waiting for snowflake events")
	go func() {
		for {
			select {
			case evt := <-ch:
				log.Debug().
					Str("tag", evt.Tag).
					Str("progress", fmt.Sprintf("%02d%%", evt.Progress)).
					Msg("Snowflake progress")
				b.snowflakeProgress = evt.Progress
				b.snowflakeCh <- evt
			}
		}

	}()
}

func (b *Bonafide) maybeInitializeEIP() error {
	if b.gateways != nil && len(b.gateways.available) > 0 {
		return nil
	}
	// FIXME - use config/bitmask flag
	if os.Getenv("SNOWFLAKE") == "1" {
		log.Info().Msg("Snowflake is enabled. Fetching eip json and certificate via snowflake (SNOWFLAKE=1)")
		p := strings.ToLower(config.ProviderConfig.Provider)
		if b.snowflakeProgress != 100 {
			ch := make(chan *snowflake.StatusEvent, 20)
			b.watchSnowflakeProgress(ch)
			err := snowflake.BootstrapWithSnowflakeProxies(p, ch)
			if err != nil {
				return fmt.Errorf("Could not initialize snowflake: %s", err.Error())
			}
		}
		err := b.parseEipJSONFromFile()
		if err != nil {
			return err
		}
		b.gateways = newGatewayPool(b.eip)
	} else {
		err := b.fetchEipJSON()
		if err != nil {
			return err
		}
		b.gateways = newGatewayPool(b.eip)

		// XXX For now, we just initialize once per session.
		// We might update the menshen gateways every time we 'maybe initilize EIP'
		// We might also want to be more clever on when to do that
		// (when opening the locations tab in the UI, only on reconnects, ...)
		// or just periodically - but we need to modify menshen api to
		// pass a location parameter.
		if len(b.gateways.recommended) == 0 {
			b.fetchGatewaysFromMenshen()
		}
	}
	return nil
}

// GetBestGateways filters by transport, and will return the maximum number defined
// in bonafide.maxGateways, or the maximum by default (3).
func (b *Bonafide) GetBestGateways(transport string) ([]Gateway, error) {
	log.Info().Str("transport", transport).Msg("Getting gateways for")
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

// FetchGateways only filters gateways by transport.
// if "any" is provided it will return all gateways for all transports
func (b *Bonafide) FetchAllGateways(transport string) error {
	err := b.maybeInitializeEIP()
	// XXX needs to wait for bonafide too
	if err != nil {
		return err
	}
	_, err = b.gateways.GetGatewaysByTimezone(transport, b.tzOffsetHours, 999)
	return err
}

func (b *Bonafide) GetLocationQualityMap(transport string) map[string]float64 {
	return b.gateways.getLocationQualityMap(transport)
}

func (b *Bonafide) GetLocationLabels(transport string) map[string][]string {
	return b.gateways.getLocationLabels(transport)
}

func (b *Bonafide) SetManualGateway(label string) {
	log.Debug().Str("location", label).Msg("manual location selection")
	b.gateways.setUserChoice(label)
}

func (b *Bonafide) SetAutomaticGateway() {
	b.gateways.setAutomaticChoice()
}

// This function is part of the apiInterface
// In the v5 implementation, some errors can happen
// In this case we always return nil as error
func (b *Bonafide) GetBestLocation(transport string) (string, error) {
	if b.gateways == nil {
		return "", nil
	}
	return b.gateways.getBestLocation(transport, b.tzOffsetHours), nil
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
	/* FIXME in float deployments, geolocation is served on
	* gemyip.domain/json, with a LE certificate, but in riseup is served
	* behind the api certificate.  So this is a workaround until we
	* streamline that behavior */
	resp, err := b.client.Post(config.ProviderConfig.GeolocationAPI, "", nil)
	if err != nil {
		client := &http.Client{Timeout: time.Second * 30}
		_resp, err := client.Post(config.ProviderConfig.GeolocationAPI, "", nil)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not fetch geolocation")
			return err
		}
		resp = _resp
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Warn().
			Err(err).
			Int("statusCode", resp.StatusCode).
			Msg("Bad status code while fetching geolocation")
		return fmt.Errorf("Get geolocation failed with status: %d", resp.StatusCode)
	}

	geo := &geoLocation{}
	dataJSON, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(dataJSON, &geo)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not parse geolocation json")
		log.Warn().Msgf("%s", string(dataJSON))
		_ = fmt.Errorf("bad json")
		return err
	}

	log.Info().Msgf("Got sorted gateways: %v", geo.Gateways)
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

func (b *Bonafide) GetSnowflakeCh() chan *snowflake.StatusEvent {
	return b.snowflakeCh
}

func (b *Bonafide) DoGeolocationLookup() error {
	return errors.New("DoGeolocationLookup is not supported in v3 (only implemented in bitmask-core)")
}

func (b *Bonafide) SupportsObfs4() bool {
	return b.supportsTransport("obfs4")
}

func (b *Bonafide) SupportsKCP() bool {
	return b.supportsTransport("kcp")
}

func (b *Bonafide) SupportsQUIC() bool {
	return b.supportsTransport("quic")
}

func (b *Bonafide) SupportsHopping() bool {
	return b.supportsTransport("obfs4-hop")
}

func (b *Bonafide) supportsTransport(transport string) bool {
	if b.eip == nil {
		return false
	}
	for _, gw := range b.gateways.available {
		if gw.isTransport(transport) {
			return true
		}
	}
	return false
}

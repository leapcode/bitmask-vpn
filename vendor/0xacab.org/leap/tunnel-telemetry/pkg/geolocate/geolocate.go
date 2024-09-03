package geolocate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/rs/zerolog/log"
)

var (
	defaultGeolocationAPI = "https://api.dev.ooni.io/api/v1/geolookup"
)

// FindCurrentHostGeolocation will make a best-effor attempt at discovering the public IP
// of the vantage point where the software is running, and obtain geolocation metadata for it.
// This function currently uses a single endpoint for geolocation (in the OONI API).
func FindCurrentHostGeolocation() (*GeoInfo, error) {
	ip, err := AttemptFetchingPublicIP()
	if err != nil {
		return nil, err
	}

	// TODO: use smart-dialer here.
	geo := NewGeolocator()
	info, err := geo.Geolocate(ip)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// FindCurrentHostGeolocationWithSTUN first trys to get the current public ip address by
// using the given STUN servers. It then uses countryCodeLookupURL to convert the ip address
// into a country code. If countryCodeLookupURL is empty, then defaultGeolocationAPI (OONI) is used
func FindCurrentHostGeolocationWithSTUN(stunServers []string, countryCodeLookupURL string) (*GeoInfo, error) {
	var ip string
	var err error

	if len(stunServers) == 0 {
		return nil, errors.New("Could not get country code. The list of STUN servers is empty")
	}

	for _, server := range stunServers {
		log.Trace().
			Str("server", server).
			Msg("Trying STUN server")

		ip, err = FetchIPFromSTUNCall(server)
		if err == nil {
			break
		} else {
			log.Warn().
				Str("server", server).
				Err(err).
				Msg("Could not get ip using STUN server")
		}
	}

	if ip == "" {
		return nil, errors.New("Could not get ip address with STUN servers. All STUN servers failed")
	}

	// TODO: use smart-dialer here.
	geo := NewGeolocator()
	if countryCodeLookupURL == "" {
		log.Trace().
			Str("countryCodeLookupURL", defaultGeolocationAPI).
			Msg("Using default country code lookup url (OONI)")
	} else {
		geo.API = countryCodeLookupURL
		log.Trace().
			Str("countryCodeLookupURL", countryCodeLookupURL).
			Msg("Using custom country code lookup url")
	}
	info, err := geo.Geolocate(ip)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// AttemptFetchingPublicIP will attempt to get our public IP by exhausting
// all the available sources; the order is stun > https. It will return
// an error if all the sources are used and we still don't have a result.
func AttemptFetchingPublicIP() (string, error) {

	log.Trace().Msg("Trying to get current ip address using STUN servers")
	shuffleServers(stunServers)

	for _, server := range stunServers {
		log.Trace().
			Str("server", server).
			Msg("Trying STUN server")

		ip, err := FetchIPFromSTUNCall(server)
		if err != nil {
			log.Warn().
				Str("server", server).
				Err(err).
				Msg("Could not get ip using STUN server")
			continue
		}
		return ip, nil
	}

	log.Warn().Msg("Could not get current ip address using STUN servers. Using public WebAPIs")

	shuffleServers(httpsServers)
	for _, provider := range httpsServers {

		log.Trace().
			Str("provider", provider).
			Msg("Using Provider")

		ip, err := FetchIPFromHTTPSAPICall(provider)
		if err != nil {
			log.Warn().
				Str("provider", provider).
				Err(err).
				Msg("Could not get ip using WebAPI")
			continue
		} else {
			return ip, nil
		}
	}
	return "", errors.New("Could not get ip address by using STUN/WebAPIs")
}

// A Geolocator is able to geolocate IPs, using a specific http.Client.
type Geolocator struct {
	API    string
	Client *http.Client
}

// TODO: add NewGeolocationWithHTTPClient
func NewGeolocator() *Geolocator {
	return &Geolocator{
		API:    defaultGeolocationAPI,
		Client: defaultHTTPClient,
	}
}

// GeoInfo contains the minimal metadata that we need for annotating
// reports.
type GeoInfo struct {
	ASName string `json:"as_name"`
	ASN    int    `json:"asn"`
	CC     string `json:"cc"`
}

type geoLocationFromOONI struct {
	Geolocation map[string]GeoInfo `json:"geolocation"`
	Version     int                `json:"v"`
}

func (g *Geolocator) Geolocate(ip string) (*GeoInfo, error) {
	resp := &geoLocationFromOONI{}
	query := fmt.Sprintf(`{"addresses": ["%s"]}`, ip)
	if err := g.doPostJSON(g.API, []byte(query), resp); err != nil {
		return nil, fmt.Errorf("Could not use country code lookup server: %s", err)
	}
	geoinfo := resp.Geolocation[ip]
	return &geoinfo, nil
}

func (g *Geolocator) doPostJSON(url string, data []byte, jd any) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if jd != nil {
		if err := json.NewDecoder(resp.Body).Decode(jd); err != nil {
			return err
		}
	}
	return nil
}

func shuffleServers(ss []string) {
	rand.Shuffle(len(ss), func(i, j int) {
		ss[i], ss[j] = ss[j], ss[i]
	})
}

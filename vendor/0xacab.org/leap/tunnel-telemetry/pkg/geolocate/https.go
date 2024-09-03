package geolocate

import (
	"net/http"
	"time"
)

var (
	// TODO: can add ubuntu, cloudflare here
	httpsServers = []string{"bdc", "ipify", "ipinfo"}

	defaultHTTPClient = &http.Client{Timeout: 10 * time.Second}
)

type apiIP struct {
	uri    string
	lookup func() (string, error)
}

var ipProviders = map[string]apiIP{
	"bdc": func() apiIP {
		uri := "https://api-bdc.net/data/client-ip"
		type r struct {
			IP   string `json:"ipString"`
			Type string `json:"ipType"`
		}
		return apiIP{
			uri: uri,
			lookup: func() (string, error) {
				v := &r{}
				if err := getJSON(uri, v); err != nil {
					return "", err
				}
				return v.IP, nil
			},
		}
	}(),
	"ipify": func() apiIP {
		uri := "https://api.ipify.org?format=json"
		type r struct {
			IP string `json:"ip"`
		}
		return apiIP{
			uri: uri,
			lookup: func() (string, error) {
				v := &r{}
				if err := getJSON(uri, v); err != nil {
					return "", err
				}
				return v.IP, nil
			},
		}
	}(),
	"ipinfo": func() apiIP {
		uri := "https://ipinfo.io/json"
		type r struct {
			IP      string `json:"ip"`
			Country string `json:"country"`
		}
		return apiIP{
			uri: uri,
			lookup: func() (string, error) {
				v := &r{}
				if err := getJSON(uri, v); err != nil {
					return "", err
				}
				return v.IP, nil
			},
		}
	}(),
}

// FetchIPFromHTTPSAPICall tries to get the public IP via the passed HTTPS provider label.
func FetchIPFromHTTPSAPICall(provider string) (string, error) {
	return ipProviders[provider].lookup()
}

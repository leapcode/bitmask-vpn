package bootstrap

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"
)

var ubuntuGeo = "https://geoip.ubuntu.com/lookup"

type ubuntuResponse struct {
	XMLName xml.Name `xml:"Response"`
	CC      string   `xml:"CountryCode"`
}

// ubuntuGeoLookup will attempt to fetch geolocation info from ubuntu's service,
// which is contained in an xml document. We do not care about network or IP at
// this moment, and it's probably better not to log/store that info.
func ubuntuGeoLookup(client *http.Client) (string, error) {
	resp, err := client.Get(ubuntuGeo)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var v ubuntuResponse
	err = xml.Unmarshal(data, &v)
	if err != nil {
		return "", err
	}
	return strings.ToLower(v.CC), nil
}

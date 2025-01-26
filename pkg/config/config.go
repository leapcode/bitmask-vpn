package config

type ProviderCfg struct {
	Provider             string
	ApplicationName      string
	BinaryName           string
	Auth                 string
	APIURL               string
	GeolocationAPI       string
	ApiVersion           int
	CaCert               []byte
	STUNServers          []string
	CountryCodeLookupURL string
}

var ProviderConfig = &ProviderCfg{}

func init() {
	if ProviderConfig == nil {
		ProviderConfig = &ProviderCfg{}
	}
}

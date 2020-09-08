package config

var (
	Provider        = ""
	ProviderName    = ""
	ApplicationName = ""
	BinaryName      = ""
	Auth            = ""
	AuthEmptyPass   = false
	DonateURL       = ""
	AskForDonations = true
	HelpURL         = ""
	TosURL          = ""
	APIURL          = ""
	GeolocationAPI  = ""
)

var Version string

/*

CaCert : a string containing a representation of the provider CA, used to
        sign the webapp and openvpn certificates. should be placed in
        config/[provider]-ca.crt

*/
var CaCert = []byte("")

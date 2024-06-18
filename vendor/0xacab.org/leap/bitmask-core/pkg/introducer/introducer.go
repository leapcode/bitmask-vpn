// introducer is a simple obfuscated proxy that points to an instance of the menshen API.
// The upstream API can be publicly available or not.
package introducer

import (
	"fmt"
	"net/url"
	"strings"
)

// Introducer has everything we need to instantiate a http.Client that dials to the obfuscated introducer.
type Introducer struct {
	Type string
	Addr string
	Cert string
	FQDN string
	KCP  bool
}

// Validate returns true if all the fields in this Introducer are like expected.
func (i *Introducer) Validate() error {
	if i.Type != "obfsvpnintro" {
		return fmt.Errorf("unknown type: %s", i.Type)
	}
	if len(strings.Split(i.Addr, ":")) != 2 {
		return fmt.Errorf("expected address in format ipaddr:port")
	}
	if len(i.Cert) != 70 {
		return fmt.Errorf("wrong certificate len = %d", len(i.Cert))
	}
	if len(strings.Split(i.FQDN, ".")) < 2 {
		return fmt.Errorf("expected a FQDN, got: %s", i.FQDN)
	}
	return nil
}

// URL produces the canonical URL for this introducer. We need to make sure to use this URL
// in the internal storage, so that we can ensure equality regardless of parameter order.
func (i *Introducer) URL() string {
	var kcp string
	switch {
	case i.KCP:
		kcp = "1"
	default:
		kcp = "0"
	}
	return fmt.Sprintf("%s://%s/?cert=%s&fqdn=%s&kcp=%s",
		i.Type,
		i.Addr,
		i.Cert,
		i.FQDN,
		kcp,
	)
}

// NewIntroducerFromURL returns a new Introducer after parsing the passed URL. It will also return an error
// if it was not possible to parse the URL correctly.
func NewIntroducerFromURL(introducerURL string) (*Introducer, error) {
	// Parse the introducer URL string
	u, err := url.Parse(introducerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse introducer URL: %w", err)
	}

	// Extract FQDN from query parameters
	fqdn := u.Query().Get("fqdn")
	if fqdn == "" {
		return nil, fmt.Errorf("FQDN not found in the introducer URL")
	}

	// Extract KCP from query parameters. It defaults to false.
	kcp := false
	if kcpValue := u.Query().Get("kcp"); kcpValue == "1" {
		kcp = true
	}

	// Extract Cert from query parameters
	cert := u.Query().Get("cert")
	if cert == "" {
		return nil, fmt.Errorf("cert not found in the introducer URL")
	}

	introducer := &Introducer{
		Type: u.Scheme,
		Addr: u.Host,
		FQDN: fqdn,
		KCP:  kcp,
		Cert: cert,
	}

	return introducer, nil
}

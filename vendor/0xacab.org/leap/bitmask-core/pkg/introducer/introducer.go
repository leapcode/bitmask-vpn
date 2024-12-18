// introducer is a simple obfuscated proxy that points to an instance of the menshen API.
// The upstream API can be publicly available or not.
package introducer

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

// Introducer has everything we need to instantiate a http.Client that dials to the obfuscated introducer.
type Introducer struct {
	Type string // currently only obfsvpnintro is supported
	Addr string // host/ip of proxy - can have a port, e.g. proxy.menshen.net:4430
	Cert string // obfs4 proxy cert
	FQDN string // hostname of menshen hostname the proxy server should connect to
	KCP  bool   // use obfs4 with kcp
	Auth string // credentials for menshen authentication to get private gateways/bridges
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
	if i.FQDN != "localhost" && len(strings.Split(i.FQDN, ".")) < 2 {
		return fmt.Errorf("expected a FQDN, got: %s", i.FQDN)
	}
	if len(i.Auth) == 0 {
		log.Warn().Msg("Invite token of introducer url is empty")
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
	return fmt.Sprintf("%s://%s/?cert=%s&fqdn=%s&kcp=%s&auth=%s",
		i.Type,
		i.Addr,
		url.QueryEscape(i.Cert),
		url.QueryEscape(i.FQDN),
		kcp,
		url.QueryEscape(i.Auth),
	)
}

// NewIntroducerFromURL returns a new Introducer after parsing the passed URL. It will also return an error
// if it was not possible to parse the URL correctly.
func NewIntroducerFromURL(introducerURL string) (*Introducer, error) {
	// Parse the introducer URL string. Get parameters are automatically url decoded
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

	// Extract Auth from query parameters (can be empty)
	auth := u.Query().Get("auth")

	introducer := &Introducer{
		Type: u.Scheme,
		Addr: u.Host,
		FQDN: fqdn,
		KCP:  kcp,
		Cert: cert,
		Auth: auth,
	}

	return introducer, nil
}

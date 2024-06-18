package bootstrap

import (
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/babolivier/go-doh-client"
)

var (
	// we can use ip addresses or hostnames here
	// doh will connect to tcp:443 and verify the certificate
	// if you specify an ip address, make sure that the ip
	// is part of the common name
	defaultResolver = []string{
		"208.67.222.222",  // OpenDNS https://www.opendns.com/setupguide/
		"9.9.9.9",         // quad9 https://www.quad9.net/
		"dns.mullvad.net", // https://mullvad.net/en/help/dns-over-https-and-dns-over-tls
		"dns.njal.la",     // A free non logging and uncensored public DNS with DNS-over-TLS/HTTPS https://dns.njal.la/
	}
)

func dohQuery(domain string) (string, error) {

	for _, dnsServer := range defaultResolver {
		log.Debug().
			Str("dnsServer", dnsServer).
			Msg("Selected DoH provider")

		resolver := doh.Resolver{
			Host:       dnsServer,
			Class:      doh.IN,
			HTTPClient: &http.Client{Timeout: 10 * time.Second},
		}

		ips, _, err := resolver.LookupA(domain)
		if err != nil {
			log.Warn().
				Str("resolver", dnsServer).
				Str("domain", domain).
				Err(err).
				Msg("Could not resolve host with DNS over HTTPs")
			continue
		}
		return ips[0].IP4, nil
	}
	return "", errors.New("Could not resolve ip with DNS over HTTPS. Tried all resolvers")

}

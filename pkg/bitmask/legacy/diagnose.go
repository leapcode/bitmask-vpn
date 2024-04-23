package legacy

import (
	"net"

	"github.com/rs/zerolog/log"
)

func logDnsLookup(domain string) {
	addrs, err := net.LookupHost(domain)
	if err != nil {
		log.Warn().
			Err(err).
			Str("domain", domain).
			Msg("Could not resolve address")
	}

	log.Debug().
		Str("domain", domain).
		Msg("Resolving domain ...")
	for _, addr := range addrs {
		log.Debug().
			Str("domain", domain).
			Str("addr", addr).
			Msg("Resolved to ip")
	}
}

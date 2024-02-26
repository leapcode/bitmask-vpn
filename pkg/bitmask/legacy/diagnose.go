package legacy

import (
	"log"
	"net"
)

func logDnsLookup(domain string) {
	addrs, err := net.LookupHost(domain)
	if err != nil {
		log.Println("ERROR cannot resolve address:", domain)
		log.Println(err)
	}

	log.Println("From here,", domain, "resolves to:")
	for _, addr := range addrs {
		log.Println(addr)
	}
}

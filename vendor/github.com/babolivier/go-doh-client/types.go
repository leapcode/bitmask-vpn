package doh

import (
	"net"
)

// DNSType implements DNS values.
type DNSType uint16

const (
	// A implements the DNS A type.
	A DNSType = 1
	// NS implements the DNS NS type.
	NS = 2
	// CNAME implements the DNS CNAME type.
	CNAME = 5
	// SOA implements the DNS SOA type.
	SOA = 6
	// PTR implements the DNS PTR type.
	PTR = 12
	// MX implements the DNS MX type.
	MX = 15
	// TXT implements the DNS TXT type.
	TXT = 16
	// AAAA implements the DNS AAAA type.
	AAAA = 28
	// SRV implements the DNS SRV type.
	SRV = 33
)

// DNSClass implements DNS classes.
type DNSClass uint16

const (
	// IN implement the DNS Internet class.
	IN DNSClass = 1
	// CS implements the DNS CSNET class.
	CS = 2
	// CH implements the DNS CH class.
	CH = 3
	// HS implements the DNS Hesiod class.
	HS = 4
	// ANYCLASS implements the DNS * QCLASS.
	ANYCLASS = 255
)

// ARecord implements the DNS A record.
type ARecord struct {
	IP4 string
}

// AAAARecord implements the DNS AAAA record.
type AAAARecord struct {
	IP6 string
}

// CNAMERecord implements the DNS CNAME record.
type CNAMERecord struct {
	CNAME string
}

// TXTRecord implements the DNS TXT record.
type TXTRecord struct {
	TXT string
}

// SOARecord implements the DNS SOA record.
type SOARecord struct {
	PrimaryNS   string
	RespMailbox string
	Serial      uint32
	Refresh     int32
	Retry       int32
	Expire      int32
	Minimum     uint32
}

// PTRRecord implements the DNS PTR record.
type PTRRecord struct {
	PTR string
}

// MXRecord implements the DNS MX record.
type MXRecord net.MX

// SRVRecord implements the DNS SRV record.
type SRVRecord net.SRV

// NSRecord implements the DNS NS record.
type NSRecord net.NS

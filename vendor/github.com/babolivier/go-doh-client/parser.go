package doh

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// parser is the internal structure handling parsing of RDATA fields in DNS
// responses.
type parser struct {
	// res is the full response (including headers), which is needed in order to
	// properly parse domain names when compressed as described in section 4.1.4
	// of RFC 1035.
	res []byte
}

// parse is a generic function which calls the right function for a given DNS
// type in order to parse an answer's data.
func (p *parser) parse(t DNSType, c DNSClass, rdata []byte) interface{} {
	// Types compatible with all classes.
	switch t {
	case CNAME:
		return p.parseCNAME(rdata)
	case MX:
		return p.parseMX(rdata)
	case SRV:
		return p.parseSRV(rdata)
	case NS:
		return p.parseNS(rdata)
	case TXT:
		return p.parseTXT(rdata)
	case SOA:
		return p.parseSOA(rdata)
	case PTR:
		return p.parsePTR(rdata)
	}

	// Internet-specific types.
	if c == IN || c == ANYCLASS {
		switch t {
		case A:
			return p.parseA(rdata)
		case AAAA:
			return p.parseAAAA(rdata)
		}
	}

	return nil
}

// parseA parses A records.
func (p *parser) parseA(rdata []byte) *ARecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    ADDRESS                    |
		|                                               |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/
	var ip []string
	for i := 0; i < len(rdata); i++ {
		ip = append(ip, strconv.Itoa(int(rdata[i])))
	}

	a := new(ARecord)
	a.IP4 = strings.Join(ip, ".")

	return a
}

// parseAAAA parses AAAA records.
func (p *parser) parseAAAA(rdata []byte) *AAAARecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                                               |
		|                                               |
		|                                               |
		|                    ADDRESS                    |
		|                                               |
		|                                               |
		|                                               |
		|                                               |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/
	var ip []string
	for i := 0; i < len(rdata); i += 2 {
		ip = append(ip, fmt.Sprintf("%x", binary.BigEndian.Uint16(rdata[i:i+2])))
	}

	// TODO: Compress e.g. a:0:0:0:b into a::b
	aaaa := new(AAAARecord)
	aaaa.IP6 = strings.Join(ip, ":")

	return aaaa
}

// parseCNAME parses CNAME records.
func (p *parser) parseCNAME(rdata []byte) *CNAMERecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		/                     NAME                      /
		/                                               /
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/
	cname := new(CNAMERecord)
	cname.CNAME, _ = p.parseName(rdata)

	return cname
}

// parseMX parses MX records.
func (p *parser) parseMX(rdata []byte) *MXRecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                  PREFERENCE                   |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		/                   EXCHANGE                    /
		/                                               /
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/
	mx := new(MXRecord)
	mx.Pref = binary.BigEndian.Uint16(rdata[0:2])
	mx.Host, _ = p.parseName(rdata[2:])

	return mx
}

// parseSRV parses SRV records.
func (p *parser) parseSRV(rdata []byte) *SRVRecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                   PRIORITY                    |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    WEIGHT                     |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                     PORT                      |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    TARGET                     |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/
	srv := new(SRVRecord)
	srv.Priority = binary.BigEndian.Uint16(rdata[0:2])
	srv.Weight = binary.BigEndian.Uint16(rdata[2:4])
	srv.Port = binary.BigEndian.Uint16(rdata[4:6])
	srv.Target, _ = p.parseName(rdata[6:])
	return srv
}

// parseNS parses NS records.
func (p *parser) parseNS(rdata []byte) *NSRecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		/                   NSDNAME                     /
		/                                               /
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/
	ns := new(NSRecord)
	ns.Host, _ = p.parseName(rdata)
	return ns
}

// parseTXT parses TXT records.
func (p *parser) parseTXT(rdata []byte) *TXTRecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+
		|         LENGTH        |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		/                   TXT-DATA                    /
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/

	length := int(rdata[0])

	txt := new(TXTRecord)
	txt.TXT = string(rdata[1 : length+1])

	return txt
}

// parseSOA parses SOA records.
func (p *parser) parseSOA(rdata []byte) *SOARecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		/                     MNAME                     /
		/                                               /
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		/                     RNAME                     /
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    SERIAL                     |
		|                                               |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    REFRESH                    |
		|                                               |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                     RETRY                     |
		|                                               |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    EXPIRE                     |
		|                                               |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    MINIMUM                    |
		|                                               |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/

	var offset int

	soa := new(SOARecord)
	soa.PrimaryNS, offset = p.parseName(rdata)
	rdata = rdata[offset:]

	soa.RespMailbox, offset = p.parseName(rdata)
	rdata = rdata[offset:]

	soa.Serial = binary.BigEndian.Uint32(rdata[0:4])
	soa.Refresh = int32(binary.BigEndian.Uint32(rdata[4:8]))
	soa.Retry = int32(binary.BigEndian.Uint32(rdata[8:12]))
	soa.Expire = int32(binary.BigEndian.Uint32(rdata[12:16]))
	soa.Minimum = binary.BigEndian.Uint32(rdata[16:20])

	return soa
}

// parsePTR parses PTR records.
func (p *parser) parsePTR(rdata []byte) *PTRRecord {
	/*
		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		/                   PTRDNAME                    /
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/

	ptr := new(PTRRecord)
	ptr.PTR, _ = p.parseName(rdata)

	return ptr
}

// parseName parses a domain name as described in the QNAME definition of
// section 4.1.2 of RFC 1035, with support for compression.
// Returns the domain name with points as the separator between labels, as well
// as the number of bytes the name represents in the payload it's been given.
func (p *parser) parseName(b []byte) (name string, offset int) {
	var labels []string
	for {
		length := int(b[offset])
		// A length of 0 means we've reached the end of the domain name.
		if length == 0 {
			offset++
			break
		}

		// If the two most significant bits of the first byte are both 1, it
		// means compression is used for the rest of the domain name.
		if length>>6 == 3 {
			// 16383 is b10 for b2 00111111 11111111, which matches with the
			// pointer to the next labels without the two "11" most significant
			// bits.
			ptr := binary.BigEndian.Uint16(b[offset:offset+2]) & 16383
			label, _ := p.parseName(p.res[ptr:])
			labels = append(labels, label)
			offset += 2
			// RFC says the pointer points to "an entire domain name or a list
			// of labels at the end of a domain name", so we can safely assume
			// that's the end of the name.
			break
		} else {
			labels = append(labels, string(b[offset+1:offset+length+1]))
			offset += length + 1
		}

	}

	return strings.Join(labels, "."), offset
}

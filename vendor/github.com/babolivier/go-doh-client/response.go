package doh

import (
	"encoding/binary"
)

// answer describes a parsed answer from the response message.
type answer struct {
	name   string
	t      DNSType
	class  DNSClass
	ttl    uint32
	parsed interface{}
}

// parseResponse parses the message the resolver responded with.
// Returns all of the answers included in the message.
// Returns an error if the message isn't a response, if the message includes
// header values that are not currently supported, or if the message includes an
// error code.
func parseResponse(res []byte) ([]answer, error) {
	p := new(parser)
	p.res = res

	if len(res) < DNSMsgHeaderLen {
		return nil, ErrCorrupted
	}

	/*
		DNS HEADER

		                               1  1  1  1  1  1
		 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                      ID                       |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|QR|   Opcode  |AA|TC|RD|RA| Z|AD|CD|   RCODE   |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    QDCOUNT                    |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    ANCOUNT                    |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    NSCOUNT                    |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		|                    ARCOUNT                    |
		+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	*/

	// Check QR == 1 (is response)
	qr := res[2] >> 7
	if qr != 1 {
		return nil, ErrNotAResponse
	}

	// Check Opcode == 0 (is standard query)
	// TODO: Support other values.
	opcode := res[2] >> 3 & 15
	if opcode != 0 {
		return nil, ErrNotStandardQuery
	}

	// Check TC == 0 (not truncated)
	// TODO: Support truncated messages.
	tc := res[2] >> 1 & 1
	if tc != 0 {
		return nil, ErrTruncated
	}

	// Check RCODE == 0 (no error)
	rcode := res[3] & 15
	if rcode != 0 {
		return nil, dnsErrors[rcode]
	}

	qdcount := binary.BigEndian.Uint16(res[4:6])
	ancount := binary.BigEndian.Uint16(res[6:8])

	// Get to the very first byte after decoding headers.
	buf := res[DNSMsgHeaderLen:]
	var i uint16
	for i = 0; i < qdcount; i++ {
		/*
			Parse queries
			We only process them in order to reach the answers section.

			                               1  1  1  1  1  1
			 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|                                               |
			/                     QNAME                     /
			/                                               /
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|                     QTYPE                     |
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|                     QCLASS                    |
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		*/
		if len(buf) == 0 {
			return nil, ErrCorrupted
		}
		_, offset := p.parseName(buf)
		buf = buf[offset+4:]
	}

	// Now buf should be at the first byte of the first answer.
	answers := make([]answer, 0)
	for i = 0; i < ancount; i++ {
		/*
			Parse answers

			                               1  1  1  1  1  1
			 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|                                               |
			/                                               /
			/                      NAME                     /
			|                                               |
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|                      TYPE                     |
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|                     CLASS                     |
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|                      TTL                      |
			|                                               |
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			|                   RDLENGTH                    |
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
			/                     RDATA                     /
			/                                               /
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+

			NAME (or some labels) can be compressed as:

			                               1  1  1  1  1  1
			 0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
			| 1  1|                OFFSET                   |
			+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
		*/

		if len(buf) == 0 {
			return nil, ErrCorrupted
		}
		name, offset := p.parseName(buf)
		t := DNSType(binary.BigEndian.Uint16(buf[offset : offset+2]))
		class := DNSClass(binary.BigEndian.Uint16(buf[offset+2 : offset+4]))
		ttl := binary.BigEndian.Uint32(buf[offset+4 : offset+8])
		rdlength := binary.BigEndian.Uint16(buf[offset+8 : offset+10])
		rdata := buf[offset+10 : offset+10+int(rdlength)]

		// Set buffer value for next occurrence.
		buf = buf[offset+10+int(rdlength):]

		// Parse the answer.
		parsed := p.parse(t, class, rdata)
		answers = append(answers, answer{
			name:   name,
			t:      t,
			class:  class,
			ttl:    ttl,
			parsed: parsed,
		})
	}

	return answers, nil
}

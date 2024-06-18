package doh

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"strings"
	"time"
)

// encodeQuery creates a DNS query message from the given fqdn, type and class.
func encodeQuery(fqdn string, t DNSType, c DNSClass) []byte {
	q := bytes.NewBuffer(nil)

	reqID := []byte{0, 0}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	binary.BigEndian.PutUint16(reqID, uint16(r.Int31()))

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
	q.Write([]byte{
		reqID[0], reqID[1],
		// QR = 0 (query)
		// OPCODE = 0 (standard query)
		// AA ignored
		// TC = 0 (not truncated)
		// RD = 1 (recursion desired)
		(0 << 7) | (0 << 3) | (0 << 1) | 1,
		// RA ignored
		// Z = 0 (reserved)
		// AD = 0
		// CD = 1
		// RCODE ignored
		(1 << 4),
		// QDCOUNT = 1
		byte(0), byte(1),
		// ANCOUNT = 0
		byte(0), byte(0),
		// NSCOUNT = 0
		byte(0), byte(0),
		// ARCOUNT = 0
		byte(0), byte(0),
	})

	qtype := []byte{0, 0}
	binary.BigEndian.PutUint16(qtype, uint16(t))
	qclass := []byte{0, 0}
	binary.BigEndian.PutUint16(qclass, uint16(c))

	/*
		DNS QUERY

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
	labels := strings.Split(fqdn, ".")
	for _, l := range labels {
		q.Write([]byte{byte(len(l))})
		q.Write([]byte(l))
	}
	q.Write([]byte{0})
	q.Write(qtype)
	q.Write(qclass)

	return q.Bytes()
}

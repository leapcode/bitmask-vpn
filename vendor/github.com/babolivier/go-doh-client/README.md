# go-doh-client

[![Build Status](https://travis-ci.org/babolivier/go-doh-client.svg?branch=master)](https://travis-ci.org/babolivier/go-doh-client) [![GoDoc](https://godoc.org/github.com/babolivier/go-doh-client?status.svg)](https://godoc.org/github.com/babolivier/go-doh-client) [![Go Report Card](https://goreportcard.com/badge/github.com/babolivier/go-doh-client)](https://goreportcard.com/report/github.com/babolivier/go-doh-client) [![codecov](https://codecov.io/gh/babolivier/go-doh-client/branch/master/graph/badge.svg)](https://codecov.io/gh/babolivier/go-doh-client)

This is a Go client library implementation of DNS over HTTPS
([RFC8484](https://tools.ietf.org/html/rfc8484)).

## Compliance with DNS specifications

This client library doesn't currently implement all of the DNS specifications.

It implements looking up the following records:

* A
* AAAA
* CNAME
* MX
* NS
* TXT
* SRV
* SOA
* PTR

It also currently doesn't implement other query types than standard query, nor
support for truncated messages. Full compliance, at least with [RFC
1035](https://tools.ietf.org/html/rfc1035), is something I'd like, though, so
all of that should come in the future.

## Usage

This client library should be as easy to use as any other DNS client library.
The only difference is the transport layer it uses to perform lookups.

Here's a quick example:

```go
package main

import (
	"log"

	"github.com/babolivier/go-doh-client"
)

func main() {
	resolver := doh.Resolver{
		Host:  "9.9.9.9", // Change this with your favourite DoH-compliant resolver.
		Class: doh.IN,
	}

	// Perform a A lookup on example.com
	a, _, err := resolver.LookupA("example.com")
	if err != nil {
		panic(err)
	}
	println(a[0].IP4) // 93.184.216.34

	// Perform a SRV lookup for e.g. a Matrix homeserver
	srv, _, err := resolver.LookupService("matrix", "tcp", "example.com")
	if err != nil {
		panic(err)
	}
	println(srv[0].Target) // matrix.example.com
}
```

## Why?

I grew quite interested in how the Internet works lately, which implies spending
some time reading DNS-related RFCs. On top of that, DNS over HTTPS is something
I'm interested in quite a lot for privacy reasons and because of how harder it
is to censor than classic DNS, so I decided to give it a go. And also because my
definition of "having fun during holidays" obviously involves implementing part
of the DNS RFC.

## Contribute

Contributions are more than welcome. I tried to make this library as friendly to
hack on as possible, especially when the said hack aims to implement support for
a new DNS record type.
[Here](https://github.com/babolivier/go-doh-client/commit/e64451280e70778bf8d95ea1f23e86d047a80222)'s
an example of how to do so, which is the exhaustive changeset for the
implementation of SOA records.

And of course, if you have any issue or feedback you want to report on, feel
free to [open an issue](https://github.com/babolivier/go-doh-client/issues/new).

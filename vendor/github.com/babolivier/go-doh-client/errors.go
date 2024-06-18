package doh

import (
	"errors"
)

// ErrFormatError means that the name server was unable to interpret the query.
var ErrFormatError = errors.New("Format error")

// ErrServerFailure means that The name server was unable to process this query
// due to a problem with the name server.
var ErrServerFailure = errors.New("Server failure")

// ErrNameError means that the domain name referenced in the query does not
// exist.
var ErrNameError = errors.New("Name error")

// ErrNotImplemented means that the name server does not support the requested
// kind of query.
var ErrNotImplemented = errors.New("Not implemented")

// ErrRefused means that The name server refuses to perform the specified
// operation for policy reasons. For example, a name server may not wish to
// provide the information to the particular requester, or a name server may not
// wish to perform a particular operation (e.g., zone transfer) for particular
// data.
var ErrRefused = errors.New("Refused")

var dnsErrors = []error{
	nil,
	ErrFormatError,
	ErrServerFailure,
	ErrNameError,
	ErrNotImplemented,
	ErrRefused,
}

// ErrNotAResponse means that the server responded with a message that isn't a
// response.
var ErrNotAResponse = errors.New("the message the server sent us isn't a response")

// ErrNotIN means that the lookup can only be performed with the DNS class IN
// (e.g. A, AAAA).
var ErrNotIN = errors.New("class must be IN (Internet) (or ANYCLASS (*), which includes IN)")

// ErrNotStandardQuery means that the server responded with an OPCODE header
// that isn't a standard query, which is the only value currently supported.
var ErrNotStandardQuery = errors.New("only standard queries are supported")

// ErrTruncated means that the message is truncated, which isn't currently
// supported.
var ErrTruncated = errors.New("truncated messages aren't supported")

// ErrCorrupted means that the message sent back by the server is either empty,
// incomplete, or corrupted.
var ErrCorrupted = errors.New("the message the server sent is empty, incomplete, or corrupted")

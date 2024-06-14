package vpn

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"math"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Validate the correctness of a client credential pem file. The file should contain a private
// key (-----BEGIN RSA PRIVATE KEY-----) and a certificate (-----BEGIN CERTIFICATE-----)
// It also checks if the certificate is expired. It does not check if the certificate is signed
// by config.CaCert.
func isValidCert(path string) bool {
	log.Trace().
		Str("path", path).
		Msg("Checking for a valid OpenVPN client credentials (key and certificate)")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Debug().
			Str("path", path).
			Msg("Could not read certificate file")
		return false
	}

	beginRsaKey := "-----BEGIN RSA PRIVATE KEY-----"
	if !strings.Contains(string(data), beginRsaKey) {
		log.Debug().
			Str("pem", string(data)).
			Msg("Certificate file does not contain a private key")
		return false
	}

	_, rest := pem.Decode(data)
	if rest == nil {
		log.Warn().
			Str("data", string(data)).
			Msg("Could not decode pem data")
		return false
	}

	certBlock, rest := pem.Decode(rest)
	if certBlock == nil || rest == nil {
		log.Warn().Msg("Invalid result after decoding of pem data")
		return false
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not parse x509 certificate")
		return false
	}
	loc, _ := time.LoadLocation("UTC")
	expires := cert.NotAfter
	tomorrow := time.Now().In(loc).Add(24 * time.Hour)

	if !expires.After(tomorrow) {
		log.Debug().
			Str("path", path).
			Msg("Certificate is expired")
		return false
	}

	log.Debug().
		Str("path", path).
		Msg("Successfully verified certificate")
	return true
}

// Generate a random password with len l
func getRandomPass(l int) string {
	buff := make([]byte, int(math.Round(float64(l)/float64(1.33333333333))))
	rand.Read(buff)
	str := base64.RawURLEncoding.EncodeToString(buff)
	return str[:l] // strip 1 extra character we get from odd length results
}

// Resolve host and log - used for analyzing blocked clients
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

package vpn

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"math"
	"net"
	"os"
	"strings"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
	"github.com/rs/zerolog/log"
)

const (
	privateKey = "PRIVATE KEY"
	cert       = "CERTIFICATE"
)

func isUpgradeAvailable() bool {

	// SNAPS have their own way of upgrading. We probably should also try to detect
	// if we've been installed via another package manager.
	// For now, it's maybe a good idea to disable the UI check in linux, and be
	// way more strict in windows/osx.
	if os.Getenv("SNAP") != "" {
		return false
	}
	return version.CanUpgrade()
}

// Validate the correctness of a client credential pem file. The file should contain a private
// key (-----BEGIN RSA PRIVATE KEY-----) and a certificate (-----BEGIN CERTIFICATE-----)
// It also checks if the certificate is expired. It does not check if the certificate is signed
// by config.CaCert.
func isValidCert(path string) bool {
	log.Trace().
		Str("path", path).
		Msg("Checking for valid OpenVPN client credentials (key and certificate)")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Debug().
			Str("path", path).
			Str("err", err.Error()).
			Msg("Could not read certificate file")
		return false
	}

	pkBlock, rest := pem.Decode(data)
	if rest == nil || pkBlock == nil {
		log.Warn().
			Str("data", string(data)).
			Msg("Could not decode pem data")
		return false
	}

	if !strings.Contains(pkBlock.Type, privateKey) {
		log.Debug().
			Str("pem", string(data)).
			Msg("Certificate file does not contain a private key")
		return false
	}

	certBlock, _ := pem.Decode(rest)
	if certBlock == nil || certBlock.Type != cert {
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
		Msg("Resolving domain")
	for _, addr := range addrs {
		log.Debug().
			Str("domain", domain).
			Str("addr", addr).
			Msg("Resolved to ip")
	}
}

package legacy

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"time"

	"github.com/rs/zerolog/log"
)

func isValidCert(path string) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	// skip private key, but there should be one
	_, rest := pem.Decode(data)
	certBlock, rest := pem.Decode(rest)
	if len(rest) != 0 {
		log.Warn().Msg("ERROR bad cert data")
		return false
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	loc, _ := time.LoadLocation("UTC")
	expires := cert.NotAfter
	tomorrow := time.Now().In(loc).Add(24 * time.Hour)

	if !expires.After(tomorrow) {
		return false
	} else {
		log.Debug().
			Str("path", path).
			Msg("We have a valid cert")
		return true
	}
}

package legacy

import (
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"time"
)

func isValidCert(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	// skip private key, but there should be one
	_, rest := pem.Decode(data)
	certBlock, rest := pem.Decode(rest)
	if len(rest) != 0 {
		log.Println("ERROR bad cert data")
		return false
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	loc, _ := time.LoadLocation("UTC")
	expires := cert.NotAfter
	tomorrow := time.Now().In(loc).Add(24 * time.Hour)

	if !expires.After(tomorrow) {
		return false
	} else {
		log.Println("DEBUG We have a valid cert:", path)
		return true
	}
}

package bitmask

import (
	"bytes"
	"testing"
)

var (
	privateKeyHeader = []byte("-----BEGIN RSA PRIVATE KEY-----")
	certHeader       = []byte("-----BEGIN CERTIFICATE-----")
)

func TestGetCert(t *testing.T) {
	cert, err := getCertPem()
	if err != nil {
		t.Fatal("get_cert returned an error: ", err)
	}

	if !bytes.Contains(cert, privateKeyHeader) {
		t.Errorf("No private key present: \n%q", cert)
	}

	if !bytes.Equal(cert, certHeader) {
		t.Errorf("No cert present: \n%q", cert)
	}
}

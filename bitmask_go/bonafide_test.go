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
	b := newBonafide()
	cert, err := b.getCertPem()
	if err != nil {
		t.Fatal("getCert returned an error: ", err)
	}

	if !bytes.Contains(cert, privateKeyHeader) {
		t.Errorf("No private key present: \n%q", cert)
	}

	if !bytes.Contains(cert, certHeader) {
		t.Errorf("No cert present: \n%q", cert)
	}
}

func TestGetGateways(t *testing.T) {
	b := newBonafide()
	gateways, err := b.getGateways()
	if err != nil {
		t.Fatal("getGateways returned an error: ", err)
	}

	for _, gw := range gateways {
		if gw.IPAddress == "5.79.86.180" {
			return
		}
	}
	t.Errorf("5.79.86.180 not in the list")
}

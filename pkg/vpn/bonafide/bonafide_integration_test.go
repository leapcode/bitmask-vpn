// +build integration
// Copyright (C) 2018 LEAP
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bonafide

import (
	"bytes"
	"testing"
)

const (
	gwIP = "199.58.81.145"
)

var (
	privateKeyHeader = []byte("-----BEGIN RSA PRIVATE KEY-----")
	certHeader       = []byte("-----BEGIN CERTIFICATE-----")
)

func TestIntegrationGetCert(t *testing.T) {
	initTestConfig()
	b := New()
	cert, err := b.GetPemCertificate()
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

func _TestGetGateways(t *testing.T) {
	// FIXME: we return only 3 gateways now
	initTestConfig()
	b := New()
	gateways, err := b.GetGateways("openvpn")
	if err != nil {
		t.Fatal("getGateways returned an error: ", err)
	}

	for _, gw := range gateways {
		if gw.IPAddress == gwIP {
			return
		}
	}
	t.Errorf("%s not in the list", gwIP)
}

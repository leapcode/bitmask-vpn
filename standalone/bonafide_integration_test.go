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

package bitmask

import (
	"bytes"
	"testing"
)

var (
	privateKeyHeader = []byte("-----BEGIN RSA PRIVATE KEY-----")
	certHeader       = []byte("-----BEGIN CERTIFICATE-----")
)

func TestIntegrationGetCert(t *testing.T) {
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

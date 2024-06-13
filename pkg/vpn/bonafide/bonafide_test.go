// Copyright (C) 2018-2020 LEAP
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
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const (
	certPath   = "testdata/cert"
	geoPath    = "testdata/geoloc.json"
	eip1Path   = "testdata/eip-service.json"
	eipPath    = "testdata/eip-service3.json"
	eipPathSip = "testdata/eip-service3-sip.json"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type mockClient struct {
	path string
	geo  string
}

func (c mockClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	/* FIXME - should get the mocked geolocation configured too */
	if strings.Contains(url, "api.black.riseup.net:9001/json") {
		f, err := os.Open(c.geo)
		return &http.Response{
			Body:       f,
			StatusCode: 200,
		}, err
	} else {
		f, err := os.Open(c.path)
		return &http.Response{
			Body:       f,
			StatusCode: 200,
		}, err
	}
}

func (c mockClient) Do(req *http.Request) (*http.Response, error) {
	f, err := os.Open(c.path)
	return &http.Response{
		Body:       f,
		StatusCode: 200,
	}, err
}

func TestAnonGetCert(t *testing.T) {
	b := Bonafide{client: mockClient{certPath, geoPath}}
	b.auth = &anonymousAuthentication{}
	cert, err := b.GetPemCertificate()
	if err != nil {
		t.Fatal("getCert returned an error: ", err)
	}

	f, err := os.Open(certPath)
	if err != nil {
		t.Fatal("Can't open ", certPath, ": ", err)
	}
	defer f.Close()

	certData, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("Can't read all: ", err)
	}
	if !reflect.DeepEqual(certData, cert) {
		t.Errorf("cert doesn't match")
	}
}

func TestGatewayTzLocation(t *testing.T) {
	// tzOffset -> location
	values := map[int]string{
		-12: "c",
		-10: "a",
		-5:  "a",
		-3:  "a",
		-1:  "b",
		0:   "b",
		2:   "b",
		5:   "c",
		8:   "c",
		12:  "c",
	}

	for tzOffset, location := range values {
		b := Bonafide{
			client:        mockClient{eipPath, geoPath},
			tzOffsetHours: tzOffset,
		}
		b.maxGateways = 99
		gateways, err := b.GetBestGateways("openvpn")

		if err != nil {
			t.Errorf("getGateways returned an error: %v", err)
			continue
		}
		if len(gateways) < 4 {
			t.Errorf("Wrong number of gateways for tz %d: %d", tzOffset, len(gateways))
			t.Fatal("aborting")
		}
		if gateways[0].Location != location {
			t.Errorf("Wrong location for tz %d: %s, expected: %s", tzOffset, gateways[0].Location, location)
		}
	}
}

func TestOpenvpnGateways(t *testing.T) {
	b := Bonafide{
		client: mockClient{eipPath, geoPath},
	}
	b.maxGateways = 10
	gateways, err := b.GetBestGateways("openvpn")
	if err != nil {
		t.Fatalf("getGateways returned an error: %v", err)
	}
	if len(gateways) == 0 {
		t.Fatalf("No openvpn gateways found")
	}

	present := make([]bool, 6)
	for _, g := range gateways {
		i, err := strconv.Atoi(g.Host[0:1])
		if err != nil {
			t.Fatalf("unkown host %s: %v", g.Host, err)
		}
		present[i] = true
	}
	for i, p := range present {
		switch i {
		case 0:
			continue
		case 5:
			if p {
				t.Errorf("Host %d should not have openvpn transport", i)
			}
		default:
			if !p {
				log.Println(">> present", present)
				t.Errorf("Host %d should have openvpn transport", i)
			}
		}
	}
}

func TestObfs4Gateways(t *testing.T) {
	b := Bonafide{
		client: mockClient{eipPath, geoPath},
	}
	b.maxGateways = 10
	gateways, err := b.GetBestGateways("obfs4")
	if err != nil {
		t.Fatalf("getGateways returned an error: %v", err)
	}
	if len(gateways) == 0 {
		t.Fatalf("No obfs4 gateways found")
	}

	present := make([]bool, 6)
	for _, g := range gateways {
		i, err := strconv.Atoi(g.Host[0:1])
		if err != nil {
			t.Fatalf("unkonwn host %s: %v", g.Host, err)
		}
		present[i] = true

		cert, ok := g.Options["cert"]
		if !ok {
			t.Fatalf("No cert in options (%s): %v", g.Host, g.Options)
		}
		if cert != "obfs-cert" {
			t.Errorf("No valid options given (%s): %v", g.Host, g.Options)
		}
	}
	for i, p := range present {
		switch i {
		case 0:
			continue
		case 2, 4:
			if p {
				t.Errorf("Host %d should not have obfs4 transport", i)
			}
		default:
			if !p {
				t.Errorf("Host %d should have obfs4 transport", i)
			}
		}
	}
}

type failingClient struct {
	path string
}

func (c failingClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	statusCode := 200
	if strings.Contains(url, "3/config/eip-service.json") {
		statusCode = 404
	}
	f, err := os.Open(c.path)
	return &http.Response{
		Body:       f,
		StatusCode: statusCode,
	}, err
}

func (c failingClient) Do(req *http.Request) (*http.Response, error) {
	f, err := os.Open(c.path)
	return &http.Response{
		Body:       f,
		StatusCode: 401,
	}, err
}

func TestEipServiceV1Fallback(t *testing.T) {
	b := Bonafide{
		client: failingClient{eip1Path},
	}
	b.maxGateways = 10
	gateways, err := b.GetBestGateways("obfs4")
	if err != nil {
		t.Fatalf("getGateways obfs4 returned an error: %v", err)
	}
	if len(gateways) != 0 {
		t.Fatalf("It found some obfs4 gateways: %v", gateways)
	}

	gateways, err = b.GetBestGateways("openvpn")
	if err != nil {
		t.Fatalf("getGateways openvpn returned an error: %v", err)
	}
	if len(gateways) != 4 {
		t.Fatalf("Got wrong number of gateways: %v", gateways)
	}
}

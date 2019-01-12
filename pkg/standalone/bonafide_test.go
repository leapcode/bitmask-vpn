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

package standalone

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
)

const (
	certPath = "testdata/cert"
	eipPath  = "testdata/eip-service.json"
)

type client struct {
	path string
}

func (c client) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	f, err := os.Open(c.path)
	return &http.Response{
		Body:       f,
		StatusCode: 200,
	}, err
}

func TestGetCert(t *testing.T) {
	b := bonafide{client: client{certPath}}
	cert, err := b.getCertPem()
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
		b := bonafide{
			client:        client{eipPath},
			tzOffsetHours: tzOffset,
		}
		gateways, err := b.getGateways()
		if err != nil {
			t.Errorf("getGateways returned an error: %v", err)
			continue
		}
		if len(gateways) < 4 {
			t.Errorf("Wrong number of gateways: %d", len(gateways))
			continue

		}
		if gateways[0].Location != location {
			t.Errorf("Wrong location for tz %d: %s, expected: %s", tzOffset, gateways[0].Location, location)
		}
	}
}

package bonafide

import (
	"reflect"
	"sort"
	"testing"
)

const (
	eipGwTestPath = "testdata/eip-service3.json"
)

func TestGatewayPool(t *testing.T) {
	b := Bonafide{client: mockClient{eipGwTestPath, geoPath}}
	err := b.fetchEipJSON()
	if err != nil {
		t.Fatal("fetchEipJSON returned an error: ", err)
	}

	g := gatewayPool{available: b.eip.getGateways()}
	if len(g.available) != 7 {
		/* just to check that the dataset has not changed */
		t.Fatal("Expected 7 initial gateways, got", len(g.available))
	}

	/* now we initialize a pool the proper way */
	pool := newGatewayPool(b.eip)
	if len(pool.available) != 7 {
		t.Fatal("Expected 7 initial gateways, got", len(g.available))
	}
	expectedLabels := []string{"a", "b", "c"}
	sort.Strings(expectedLabels)

	labels := pool.getLocations()
	sort.Strings(labels)
	if !reflect.DeepEqual(expectedLabels, labels) {
		t.Fatal("gatewayPool labels not what expected. Got:", labels)
	}

	if pool.userChoice != "" {
		t.Fatal("userChoice should be empty by default")
	}

	err = pool.setUserChoice("foo")
	if err == nil {
		t.Fatal("gatewayPool should not let you set a foo gateway")
	}
	err = pool.setUserChoice("a")
	if err != nil {
		t.Fatal("location 'a' should be a valid label")
	}
	err = pool.setUserChoice("c")
	if err != nil {
		t.Fatal("location 'c' should be a valid label")
	}
	if string(pool.userChoice) != "c" {
		t.Fatal("userChoice should be c")
	}

	pool.setAutomaticChoice()
	if string(pool.userChoice) != "" {
		t.Fatal("userChoice should be empty after auto selection")
	}

	_, err = pool.getRandomGatewaysByLocation("foo", "openvpn")
	if err == nil {
		t.Fatal("should get an error with invalid label")
	}

	gws, err := pool.getRandomGatewaysByLocation("a", "openvpn")
	if gws[0].IPAddress != "1.1.1.1" {
		t.Fatal("expected to get gw 1.1.1.1 with label a")
	}

	gw, err := pool.getGatewayByIP("1.1.1.1")
	if err != nil {
		t.Fatal("expected to get gw a with ip 1.1.1.1")
	}
	if gw.Host != "1.example.com" {
		t.Fatal("expected to get gw 1.example.com with ip 1.1.1.1")
	}

	// TODO test getBest

}

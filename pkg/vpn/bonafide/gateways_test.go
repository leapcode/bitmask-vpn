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
	expectedLabels := []string{"a-1", "a-2", "b-1", "b-2", "b-3", "c-1", "c-2"}
	sort.Strings(expectedLabels)

	labels := pool.getLabels()
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
	err = pool.setUserChoice("a-1")
	if err != nil {
		t.Fatal("location 'a-1' should be a valid label")
	}
	err = pool.setUserChoice("c-2")
	if err != nil {
		t.Fatal("location 'c-2' should be a valid label")
	}
	if pool.userChoice != "c-2" {
		t.Fatal("userChoice should be c-2")
	}

	pool.setAutomaticChoice()
	if pool.userChoice != "" {
		t.Fatal("userChoice should be empty after auto selection")
	}

	gw, err := pool.getGatewayByLabel("foo")
	if err == nil {
		t.Fatal("should get an error with invalid label")
	}

	gw, err = pool.getGatewayByLabel("a-1")
	if gw.IPAddress != "1.1.1.1" {
		t.Fatal("expected to get gw 1.1.1.1 with label a-1")
	}

	gw, err = pool.getGatewayByIP("1.1.1.1")
	if err != nil {
		t.Fatal("expected to get gw a with ip 1.1.1.1")
	}
	if gw.Host != "1.example.com" {
		t.Fatal("expected to get gw 1.example.com with ip 1.1.1.1")
	}

	// TODO test getBest

}

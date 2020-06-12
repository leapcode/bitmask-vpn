package bonafide

import (
	"log"
	"testing"
)

func TestBonafideAPI(t *testing.T) {
	b := New()
	cert, err := b.GetCertPem()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(cert))
}

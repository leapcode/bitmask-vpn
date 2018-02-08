package main

import (
	"log"

	"0xacab.org/leap/bitmask-systray/bitmask"
)

const (
	provider = "riseup.net"
)

func main() {
	conf, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	go notificate(conf)

	b, err := bitmask.Init()
	if err != nil {
		log.Fatal(err)
	}

	run(b, conf)
}

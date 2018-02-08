package main

import (
	"log"

	"0xacab.org/leap/bitmask-systray/bitmask"
)

const (
	provider = "riseup.net"
)

func main() {
	// TODO: do I need to bootstrap the provider?
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

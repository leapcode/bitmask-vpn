package main

import (
	"log"

	"0xacab.org/leap/bitmask-systray/bitmask"
)

const (
	provider = "riseup.net"
)

func main() {
	go notificate()

	b, err := bitmask.Init()
	if err != nil {
		log.Fatal(err)
	}

	run(b)
}

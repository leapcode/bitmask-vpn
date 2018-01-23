package main

import (
	"log"

	"0xacab.org/meskio/bitmask-systray/bitmask"
)

const (
	provider = "demo.bitmask.net"
)

func main() {
	go notificate()

	b, err := bitmask.Init()
	if err != nil {
		log.Fatal(err)
	}

	run(b)
}

package main

import (
	"time"
)

var ch chan string

func main() {
	go notificate()

	ch = make(chan string)
	run(ch)
}

func startVPN() {
	go func() {
		ch <- "starting"
		time.Sleep(time.Second * 5)
		ch <- "on"
	}()
}

func cancelVPN() {
	go func() {
		ch <- "stopping"
		time.Sleep(time.Second * 5)
		ch <- "off"
	}()
}

func stopVPN() {
	go func() {
		ch <- "failed"
	}()
}

func getVPNStatus() string {
	return "off"
}

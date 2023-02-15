package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"0xacab.org/leap/bitmask-vpn/pkg/backend"
)

func main() {
	var c string
	var installHelpers bool

	flag.StringVar(&c, "c", "", "Config file")
	flag.BoolVar(&installHelpers, "i", false, "Install helpers (asks for sudo)")
	flag.Parse()

	if installHelpers {
		backend.InstallHelpers()
		os.Exit(0)
	}

	if len(c) == 0 {
		fmt.Println("Please setup a config file with -c")
		os.Exit(1)
	}

	if _, err := os.Stat(c); err == nil {
		log.Println("Loading config file from", c)
		// all good. we could validate the json.
	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Cannot find file:", c)
		os.Exit(1)
	} else {
		// Schrodinger: file may or may not exist.
		log.Println("Error:", err)
	}

	providerDefinitionJSON, err := ioutil.ReadFile(c)
	if err != nil {
		fmt.Println("Error reading config file")
		os.Exit(1)
	}

	// TODO daemonize, or run in foreground to debug.
	log.Println("Starting bitmaskd...")

	opts := backend.InitOptsFromJSON("riseup", string(providerDefinitionJSON))
	opts.DisableAutostart = true
	opts.Obfs4 = false
	opts.StartVPN = "off"
	backend.EnableWebAPI("8000")
	backend.InitializeBitmaskContext(opts)

	log.Println("Backend initialized")

	runtime.Goexit()
	fmt.Println("Exit")
}

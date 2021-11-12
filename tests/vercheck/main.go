package main

import (
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
	"fmt"
)

func main() {
	fmt.Println("Testing version upgrade (checks network)")
	fmt.Println("-> set DEBUG=1 for details")
	u := version.CanUpgrade()
	switch {
	case u:
		fmt.Println("can upgrade")
	case !u:
		fmt.Println("no new version available")
	}
}

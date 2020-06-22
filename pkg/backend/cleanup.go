package backend

import (
	"0xacab.org/leap/bitmask-vpn/pkg/vpn"
)

func cleanup() {
	vpn.Cleanup()
}

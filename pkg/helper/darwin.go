//go:build darwin
// +build darwin

// Copyright (C) 2018-2020 LEAP
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

/*

This module holds some specific constants for osx, and it also contains the implementation of the pf firewall.

To inspect the rules in the firewall manually, use the bitmask anchor:

  sudo pfctl -s rules -a com.apple/250.BitmaskFirewall

*/

package helper

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	bitmask_anchor = "com.apple/250.BitmaskFirewall"
	gateways_table = "bitmask_gateways"
	pfctl          = "/sbin/pfctl"
	LogFolder      = "/var/log/"
)

func _getExecPath() string {
	ex, err := os.Executable()
	if err != nil {
		log.Warn().Msg("Could not get executable path")
	}
	return filepath.Dir(ex)
}

func getHelperDir() string {
	d := _getExecPath()
	return d
}

func getPlatformOpenvpnFlags() []string {
	helperDir := getHelperDir()
	return []string{
		"--script-security", "2",
		"--up", filepath.Join(helperDir, "client.up.sh"),
		"--down", filepath.Join(helperDir, "client.down.sh"),
	}
}

func parseCliArgs() {
	// OSX helper does not respond to arguments
}

func initializeService(port int) {}

func daemonize() {
}

func getOpenvpnPath() string {
	openvpnPath := filepath.Join(getHelperDir(), "openvpn.leap")
	log.Debug().
		Str("path", openvpnPath).
		Msg("Got OpenVPN path")
	return openvpnPath
}

func kill(cmd *exec.Cmd) error {
	log.Info().
		Int("pid", cmd.Process.Pid).
		Msg("Sending kill signal to pid")
	err := cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}
	return nil
}

func firewallStart(gateways []string, mode string) error {
	enablePf()
	err := resetGatewaysTable(gateways, mode)
	if err != nil {
		return err
	}

	return loadBitmaskAnchor()
}

func firewallStop() error {
	out, err := exec.Command(pfctl, "-a", bitmask_anchor, "-F", "all").Output()
	if err != nil {
		log.Warn().
			Err(err).
			Str("cmdOut", string(out)).
			Msg("Could not stop firewall")
		/* TODO return error if different from anchor not exists */
		/*return errors.New("Error while stopping firewall")*/
	}
	for range [50]int{} {
		if firewallIsUp() {
			log.Debug().Msg("Firewall still up, waiting...")
			time.Sleep(200 * time.Millisecond)
		} else {
			return nil
		}
	}
	return errors.New("Could not stop the firewall")
}

func firewallIsUp() bool {
	out, err := exec.Command(pfctl, "-a", bitmask_anchor, "-sr").Output()
	if err != nil {
		log.Warn().
			Err(err).
			Str("cmdOut", string(out)).
			Msg("Could not get status the firewall")
		return false
	}
	return strings.Contains(string(out), "block drop out proto udp from any to any port = 53")
}

func enablePf() {
	cmd := exec.Command(pfctl, "-e")
	cmd.Run()
}

func resetGatewaysTable(gateways []string, mode string) error {
	log.Debug().Msg("Resetting gateways")
	cmd := exec.Command(pfctl, "-a", bitmask_anchor, "-t", gateways_table, "-T", "delete")
	err := cmd.Run()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not delete table in firewall")
	}

	for _, gateway := range gateways {
		log.Debug().
			Str("gateway", gateway).
			Msg("Adding gateway to table")
		cmd = exec.Command(pfctl, "-a", bitmask_anchor, "-t", gateways_table, "-T", "add", gateway)
		err = cmd.Run()
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not add gateway to table")
		}
	}

	nameserver := nameserverTCP
	if mode == "udp" {
		nameserver = nameserverUDP
	}

	cmd = exec.Command(pfctl, "-a", bitmask_anchor, "-t", gateways_table, "-T", "add", nameserver)
	return cmd.Run()

}

func getDefaultDevice() string {
	out, err := exec.Command("/bin/sh", "-c", "/sbin/route -n get -net default | /usr/bin/grep interface | /usr/bin/awk '{print $2}'").Output()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not get default service")
	}
	return strings.TrimSpace(bytesToString(out))
}

func loadBitmaskAnchor() error {
	dev := getDefaultDevice()
	rulePath, err := getRulefilePath()
	if err != nil {
		return err
	}
	cmdline := fmt.Sprintf("%s -D default_device=%s -a %s -f %s", pfctl, dev, bitmask_anchor, rulePath)

	log.Debug().
		Str("cmd", cmdline).
		Msg("Loading Bitmask Anchor")

	_, err = exec.Command("/bin/sh", "-c", cmdline).Output()
	return err
}

func getRulefilePath() (string, error) {
	rulefilePath := filepath.Join(getHelperDir(), "helper", "bitmask.pf.conf")
	log.Debug().
		Str("ruleFilePath", rulefilePath).
		Msg("Got rule file path")

	if _, err := os.Stat(rulefilePath); !os.IsNotExist(err) {
		return rulefilePath, nil
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = path.Join(os.Getenv("HOME"), "go")
	}
	rulefile := path.Join(gopath, "0xacab.org", "leap", "riseup_vpn", "osx", "bitmask.pf.conf")

	if _, err := os.Stat(rulefile); !os.IsNotExist(err) {
		return rulefile, nil
	}
	return "", errors.New("Can't find rule file for the firewall")
}

func bytesToString(data []byte) string {
	return string(data[:])
}

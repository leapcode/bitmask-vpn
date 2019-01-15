// +build darwin
// Copyright (C) 2018 LEAP
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
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sevlyar/go-daemon"
)

const (
	appPath     = "/Applications/RiseupVPN.app/"
	helperPath  = appPath + "Contents/helper/"
	LogFolder   = helperPath
	openvpnPath = appPath + "Contents/Resources/openvpn.leap"

	rulefilePath   = helperPath + "bitmask.pf.conf"
	bitmask_anchor = "com.apple/250.BitmaskFirewall"
	gateways_table = "bitmask_gateways"

	pfctl = "/sbin/pfctl"
)

var (
	platformOpenvpnFlags = []string{
		"--script-security", "2",
		"--up", helperPath + "client.up.sh",
		"--down", helperPath + "client.down.sh",
	}
)

func daemonize() {
	cntxt := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "bitmask-helper.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[bitmask-helper]"},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()
	log.Print("bitmask-helper daemon started")
}

func getOpenvpnPath() string {
	return openvpnPath
}

func kill(cmd *exec.Cmd) error {
	return cmd.Process.Signal(os.Interrupt)
}

func firewallStart(gateways []string) error {
	enablePf()
	err := resetGatewaysTable(gateways)
	if err != nil {
		return err
	}

	return loadBitmaskAnchor()
}

func firewallStop() error {
	return exec.Command(pfctl, "-a", bitmask_anchor, "-F", "all").Run()
}

func firewallIsUp() bool {
	out, err := exec.Command(pfctl, "-a", bitmask_anchor, "-sr").Output()
	if err != nil {
		log.Printf("An error ocurred getting the status of the firewall: %v", err)
		return false
	}
	return bytes.Contains(out, []byte("block out proto udp to any port 53"))
}

func enablePf() {
	cmd := exec.Command(pfctl, "-e")
	cmd.Run()
}

func resetGatewaysTable(gateways []string) error {
	log.Println("Resetting gateways")
	cmd := exec.Command(pfctl, "-a", bitmask_anchor, "-t", gateways_table, "-T", "delete")
	err := cmd.Run()
	if err != nil {
		log.Printf("Can't delete table: %v", err)
	}

	for _, gateway := range gateways {
		log.Println("Adding Gateway:", gateway)
		cmd = exec.Command(pfctl, "-a", bitmask_anchor, "-t", gateways_table, "-T", "add", gateway)
		err = cmd.Run()
		if err != nil {
			log.Printf("Error adding gateway to table: %v", err)
		}
	}

	cmd = exec.Command(pfctl, "-a", bitmask_anchor, "-t", gateways_table, "-T", "add", nameserver)
	return cmd.Run()

}

func getDefaultDevice() string {
	out, err := exec.Command("/bin/sh", "-c", "/sbin/route -n get -net default | /usr/bin/grep interface | /usr/bin/awk '{print $2}'").Output()
	if err != nil {
		log.Printf("Error getting default device")
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

	log.Println("Loading Bitmask Anchor:", cmdline)

	_, err = exec.Command("/bin/sh", "-c", cmdline).Output()
	return err
}

func getRulefilePath() (string, error) {
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

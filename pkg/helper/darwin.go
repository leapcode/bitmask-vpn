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
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sevlyar/go-daemon"
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
		log.Print("ERROR: cannot find executable path")
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
	cntxt := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "bitmask-helper.log",
		LogFilePerm: 0640,
		WorkDir:     filepath.Join(getHelperDir(), "helper"),
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

func runServer(preferredPort int) {
	port := getFirstAvailablePortFrom(preferredPort)
	writePortToFile(port)
	bindAddr := "localhost:" + strconv.Itoa(port)
	serveHTTP(bindAddr)
}

func getOpenvpnPath() string {
	openvpnPath := filepath.Join(getHelperDir(), "openvpn.leap")
	log.Println("openvpn path:", openvpnPath)
	return openvpnPath
}

func kill(cmd *exec.Cmd) error {
	log.Printf("Sending kill signal to pid: %v", cmd.Process.Pid)
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
		log.Printf("An error ocurred stopping the firewall: %v", out)
		/* TODO return error if different from anchor not exists */
		/*return errors.New("Error while stopping firewall")*/
	}
	for range [50]int{} {
		if firewallIsUp() {
			log.Printf("Firewall still up, waiting...")
			time.Sleep(200 * time.Millisecond)
		} else {
			return nil
		}
	}
	return errors.New("Could not stop firewall")
}

func firewallIsUp() bool {
	out, err := exec.Command(pfctl, "-a", bitmask_anchor, "-sr").Output()
	if err != nil {
		log.Printf("An error ocurred getting the status of the firewall: %v", err)
		log.Printf(string(out))
		return false
	}
	return strings.Contains(string(out), "block drop out proto udp from any to any port = 53")
}

func enablePf() {
	cmd := exec.Command(pfctl, "-e")
	cmd.Run()
}

func resetGatewaysTable(gateways []string, mode string) error {
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
	rulefilePath := filepath.Join(getHelperDir(), "helper", "bitmask.pf.conf")
	log.Println("DEBUG: rule file path", rulefilePath)

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

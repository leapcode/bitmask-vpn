// +build linux
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

package bitmask

import (
	"errors"
	"log"
	"os"
	"os/exec"
)

const (
	systemOpenvpnPath = "/usr/sbin/openvpn"
	snapOpenvpnPath   = "/snap/bin/riseup-vpn.openvpn"
)

var bitmaskRootPaths = []string{
	"/usr/sbin/bitmask-root",
	"/usr/local/sbin/bitmask-root",
	"/snap/bin/riseup-vpn.bitmask-root",
}

type launcher struct {
	openvpnCh chan []string
}

func newLauncher() *launcher {
	l := launcher{make(chan []string, 1)}
	go l.openvpnRunner()
	return &l
}

func (l *launcher) openvpnStart(flags ...string) error {
	log.Println("openvpn start: ", flags)
	arg := []string{"openvpn", "start", getOpenvpnPath()}
	arg = append(arg, flags...)
	l.openvpnCh <- arg
	return nil
}

func (l *launcher) openvpnStop() error {
	l.openvpnCh <- nil
	log.Println("openvpn stop")
	return runBitmaskRoot("openvpn", "stop")
}

func (l *launcher) firewallStart(gateways []string) error {
	log.Println("firewall start")
	arg := []string{"firewall", "start"}
	arg = append(arg, gateways...)
	return runBitmaskRoot(arg...)
}

func (l *launcher) firewallStop() error {
	log.Println("firewall stop")
	return runBitmaskRoot("firewall", "stop")
}

func (l *launcher) openvpnRunner(arg ...string) {
	running := false
	runOpenvpn := func(arg []string) {
		for running {
			err := runBitmaskRoot(arg...)
			if err != nil {
				log.Printf("An error ocurred running openvpn: %v", err)
			}
		}
	}

	for arg := range l.openvpnCh {
		if arg == nil {
			running = false
		} else {
			running = true
			go runOpenvpn(arg)
		}
	}
}

func runBitmaskRoot(arg ...string) error {
	bitmaskRoot, err := bitmaskRootPath()
	if err != nil {
		return err
	}
	arg = append([]string{bitmaskRoot}, arg...)

	cmd := exec.Command("pkexec", arg...)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func bitmaskRootPath() (string, error) {
	for _, path := range bitmaskRootPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return path, nil
		}
	}
	return "", errors.New("No bitmask-root found")
}

func getOpenvpnPath() string {
	if os.Getenv("SNAP") != "" {
		return snapOpenvpnPath
	}
	return systemOpenvpnPath
}

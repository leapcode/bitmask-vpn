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

package standalone

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/mitchellh/go-ps"
)

const (
	systemOpenvpnPath   = "/usr/sbin/openvpn"
	snapOpenvpnPath     = "/snap/bin/" + config.BinaryName + ".openvpn"
	snapBitmaskRootPath = "/snap/bin/" + config.BinaryName + ".bitmask-root"
)

var bitmaskRootPaths = []string{
	"/usr/sbin/bitmask-root",
	"/usr/local/sbin/bitmask-root",
}

type launcher struct {
	openvpnCh chan []string
}

func newLauncher() (*launcher, error) {
	l := launcher{make(chan []string, 1)}
	go l.openvpnRunner()
	return &l, nil
}

func (l *launcher) close() error {
	return nil
}

func (l *launcher) check() (helpers bool, priviledge bool, err error) {

	/*
		isRunning, err := isPolkitRunning()
		if err != nil {
			return
		}
		if !isRunning {
			polkitPath := getPolkitPath()
			if polkitPath == "" {
				return true, false, nil
			}
			cmd := exec.Command("setsid", polkitPath)
			err = cmd.Start()
			if err != nil {
				return
			}
			isRunning, err = isPolkitRunning()
			return true, isRunning, err
		}
	*/

	return true, true, nil
}

func isPolkitRunning() (bool, error) {
	var polkitProcNames = [...]string{
		"polkit-gnome-authentication-agent-1",
		"polkit-kde-authentication-agent-1",
		"polkit-mate-authentication-agent-1",
		"lxpolkit",
		"lxsession",
		"gnome-shell",
		"gnome-flashback",
		"fingerprint-polkit-agent",
		"xfce-polkit",
	}

	processes, err := ps.Processes()
	if err != nil {
		return false, err
	}

	for _, proc := range processes {
		executable := proc.Executable()
		for _, name := range polkitProcNames {
			if strings.Contains(executable, name) {
				return true, nil
			}
		}
	}
	return false, nil
}

func getPolkitPath() string {
	var polkitPaths = [...]string{
		"/usr/bin/lxpolkit",
		"/usr/bin/lxqt-policykit-agent",
		"/usr/lib/policykit-1-gnome/polkit-gnome-authentication-agent-1",
		"/usr/lib/x86_64-linux-gnu/polkit-mate/polkit-mate-authentication-agent-1",
		"/usr/lib/mate-polkit/polkit-mate-authentication-agent-1",
		"/usr/lib/x86_64-linux-gnu/libexec/polkit-kde-authentication-agent-1",
		"/usr/lib/kde4/libexec/polkit-kde-authentication-agent-1",
		// now we get weird
		"/usr/libexec/policykit-1-pantheon/pantheon-agent-polkit",
		"/usr/lib/polkit-1-dde/dde-polkit-agent",
		// do you know some we"re still missing? :)
	}

	for _, polkit := range polkitPaths {
		_, err := os.Stat(polkit)
		if err == nil {
			return polkit
		}
	}
	return ""
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

func (l *launcher) firewallStart(gateways []gateway) error {
	log.Println("firewall start")
	arg := []string{"firewall", "start"}
	for _, gw := range gateways {
		arg = append(arg, gw.IPAddress)
	}
	return runBitmaskRoot(arg...)
}

func (l *launcher) firewallStop() error {
	log.Println("firewall stop")
	return runBitmaskRoot("firewall", "stop")
}

func (l *launcher) firewallIsUp() bool {
	err := runBitmaskRoot("firewall", "isup")
	return err == nil
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

	out, err := exec.Command("pkexec", arg...).Output()
	if err != nil {
		log.Println("Error while running bitmask-root:")
		log.Println("args: ", arg)
		log.Println("output: ", string(out))
		return err
	}
	return nil
}

func bitmaskRootPath() (string, error) {
	if os.Getenv("SNAP") != "" {
		path := snapBitmaskRootPath
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return path, nil
		}
	}
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

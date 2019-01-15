// +build windows
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

package helper

import (
	"log"
	"os"
	"os/exec"

	"0xacab.org/leap/bitmask-systray/pkg/config"
)

const (
	appPath          = `C:\Program Files\` + config.ApplicationName + `\`
	LogFolder        = appPath
	openvpnPath      = appPath + `openvpn.exe`
	chocoOpenvpnPath = `C:\Program Files\OpenVPN\bin\openvpn.exe`
)

var (
	platformOpenvpnFlags = []string{
		"--script-security", "1",
	}
)

func daemonize() {}

func getOpenvpnPath() string {
	if _, err := os.Stat(openvpnPath); !os.IsNotExist(err) {
		return openvpnPath
	} else if _, err := os.Stat(chocoOpenvpnPath); !os.IsNotExist(err) {
		return chocoOpenvpnPath
	}
	return "openvpn.exe"
}

func kill(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}

func firewallStart(gateways []string) error {
	log.Println("Start firewall: do nothing, not implemented")
	return nil
}

func firewallStop() error {
	log.Println("Stop firewall: do nothing, not implemented")
	return nil
}

func firewallIsUp() bool {
	log.Println("IsUp firewall: do nothing, not implemented")
	return false
}

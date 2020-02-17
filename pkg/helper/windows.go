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
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

const (
	svcName          = config.BinaryName + `-helper`
	appPath          = `C:\Program Files\` + config.ApplicationName + `\`
	LogFolder        = appPath
	openvpnPath      = appPath + `openvpn.exe`
	chocoOpenvpnPath = `C:\Program Files\OpenVPN\bin\openvpn.exe`
)

var (
	platformOpenvpnFlags = []string{
		"--script-security", "1",
		"--block-outside-dns",
	}
	httpBindAddr string
)

func parseCliArgs() {
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("Failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		runService(svcName, false)
		return
	}
	admin := isAdmin()
	fmt.Printf("Running as admin: %v\n", admin)
	if !admin {
		log.Fatal("Needs to be run as administrator")
	}
	if len(os.Args) < 2 {
		usage("ERROR: no command specified")
	}
	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		runService(svcName, true)
		return
	case "install":
		// TODO get binary name
		err = installService(svcName, "bitmask-helper service")
	case "remove":
		err = removeService(svcName)
	case "start":
		err = startService(svcName)
	case "stop":
		err = controlService(svcName, svc.Stop, svc.Stopped)
	default:
		usage(fmt.Sprintf("ERROR: Invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("Failed to %s %s: %v", cmd, svcName, err)
	}
	return
}

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command>\n"+
			"	where <command> is one of\n"+
			"	install, remove, debug, start, stop\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func daemonize() {}

// http server is called from within Execute in windows
func doHandleCommands(bindAddr string) {
	httpBindAddr = bindAddr
}

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

func isAdmin() bool {
	var sid *windows.SID

	// Although this looks scary, it is directly copied from the
	// official windows documentation. The Go API for this is a
	// direct wrap around the official C++ API.
	// See https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		log.Fatalf("SID Error: %s", err)
		return false
	}

	// This appears to cast a null pointer so I'm not sure why this
	// works, but this guy says it does and it Works for Me™:
	// https://github.com/golang/go/issues/28804#issuecomment-438838144
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	//fmt.Println("Admin?", member)
	if err != nil {
		log.Fatalf("Token Membership Error: %s", err)
		return false
	}
	return member

	// Also note that an admin is _not_ necessarily considered
	// elevated.
	// For elevation see https://github.com/mozey/run-as-admin
	//fmt.Println("Elevated?", token.IsElevated())
}

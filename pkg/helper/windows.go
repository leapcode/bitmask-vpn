//go:build windows
// +build windows

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

package helper

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

var (
	svcName = BinaryName + `-helper-v2`

	// XXX this is set to c:\WINDOWS\system32 on initialization. Do not use it, use a function call instead.
	appPath          = getExecDir()
	LogFolder        = getExecDir()
	openvpnPath      = path.Join(appPath, "openvpn.exe")
	chocoOpenvpnPath = `C:\Program Files\OpenVPN\bin\openvpn.exe`
	httpServerConf   = &httpConf{}
)

func getPlatformOpenvpnFlags() []string {
	return []string{
		"--script-security", "0",
		"--block-outside-dns",
		"--redirect-gateway",
	}
}

func getExecDir() string {
	ex, err := os.Executable()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not get executable path")
		return ""
	}
	/* XXX filepath.Abs is buggy, maybe because of spaces in the path. fuck it, this is good enough for now */
	return strings.Replace(ex, "\\helper.exe", "", 1)
}

type httpConf struct {
	BindAddr string
}

// parseCliArgs allows the helper binary to install/uninstall itself. It requires admin privileges.
// However, be warned: if you intend to use it from the command line, you will have to compile it with the Go compiler yourself.
// the version we're shipping (ie, cross-compiled with the mingw compiler) apparently is not able to output to stdout/stderr properly.
// To compile a usable version, from the top of the repo you can do:
// "cd cmd/bitmask-helper && GOOS=windows GOARCH=i386 go build"
func parseCliArgs() {
	log.Debug().Msg("Parsing CLI args...")
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not determine if we are running in an interactive session")
	}
	if !isIntSess {
		runService(svcName, false)
		return
	}
	log.Debug().Msg("Checking for admin")
	admin := isAdmin()
	log.Info().
		Bool("isAdmin", admin).
		Msg("Checking if we are running as admin")
	if !admin {
		log.Error().Msg("Needs to be run as administrator")
		os.Exit(2)
	}
	if len(os.Args) < 2 {
		usage("ERROR: no command specified")
	}
	cmd := strings.ToLower(os.Args[1])
	log.Debug().
		Str("cmd", cmd).
		Msg("Helper command")
	switch cmd {
	case "debug":
		// run the service on the foreground, for debugging
		runService(svcName, true)
		return
	case "install":
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
		log.Fatal().
			Err(err).
			Str("cmd", cmd).
			Str("svcName", svcName).
			Msg("Could not run helper")
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

// initializeService only initializes the server.
// we expect serveHTTP to be called from within Execute in windows
func initializeService(preferredPort int) {
	port := getFirstAvailablePortFrom(preferredPort)
	writePortToFile(port)
	httpServerConf.BindAddr = "localhost:" + strconv.Itoa(port)
	log.Info().
		Str("bindAddr", httpServerConf.BindAddr).
		Msg("Successfully initialized command server")
}

func daemonize() {}

// runServer does nothing, serveHTTP is called from within Execute in windows
func runServer(uid, gid int) {}

func getOpenvpnPath() string {
	openvpn := path.Join(getExecDir(), "openvpn.exe")
	if _, err := os.Stat(openvpn); !os.IsNotExist(err) {
		log.Debug().
			Str("openvpnPath", openvpnPath).
			Msg("Found OpenVPN path")
		return openvpn
	} else if _, err := os.Stat(chocoOpenvpnPath); !os.IsNotExist(err) {
		log.Debug().
			Str("openvpnPath", chocoOpenvpnPath).
			Msg("Found choco OpenVPN path")

		return chocoOpenvpnPath
	}
	log.Debug().Msg("Could not find system-wide openvpn...")
	return "openvpn.exe"
}

func kill(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}

func firewallStart(gateways []string, mode string) error {
	log.Warn().Msg("Start firewall: do nothing, not implemented")
	return nil
}

func firewallStop() error {
	log.Warn().Msg("Stop firewall: do nothing, not implemented")
	return nil
}

func firewallIsUp() bool {
	log.Warn().Msg("IsUp firewall: do nothing, not implemented")
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
		log.Fatal().
			Err(err).
			Msg("SID Error")
		return false
	}

	// This appears to cast a null pointer so I'm not sure why this
	// works, but this guy says it does and it Works for Me™:
	// https://github.com/golang/go/issues/28804#issuecomment-438838144
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	//fmt.Println("Admin?", member)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Token Membership Error")
		return false
	}
	return member

	// Also note that an admin is _not_ necessarily considered
	// elevated.
	// For elevation see https://github.com/mozey/run-as-admin
	//fmt.Println("Elevated?", token.IsElevated())
}

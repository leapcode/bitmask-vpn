//go:build linux
// +build linux

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

package launcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"github.com/keybase/go-ps"
	"github.com/rs/zerolog/log"
)

const (
	systemOpenvpnPath = "/usr/sbin/openvpn"
)

var bitmaskRootPaths = []string{
	"/usr/sbin/bitmask-root",
	"/usr/local/sbin/bitmask-root",
}

type Launcher struct {
	OpenvpnCh chan []string
	Failed    bool
	MngPass   string
}

func NewLauncher() (*Launcher, error) {
	l := Launcher{make(chan []string, 1), false, ""}
	go l.openvpnRunner()
	return &l, nil
}

func (l *Launcher) Close() error {
	return nil
}

// Check returns: hasHelper, hashPolkitRoot, error
func (l *Launcher) Check() (bool, bool, error) {
	if hasHelpers := hasHelpers(); !hasHelpers {
		return false, true, nil
	}

	isRunning, err := isPolkitRunning()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not check if polkit is running")
		return true, false, err
	}

	if !isRunning {
		log.Debug().Msg("A polkit daemon is not running. Trying to start")

		polkitPath, err := getPolkitPath()
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not find any usable polkit")
			return true, false, nil
		}

		log.Debug().
			Str("polkitPath", polkitPath).
			Msg("Starting polkit daemon")
		cmd := exec.Command("setsid", polkitPath)
		err = cmd.Start()
		if err != nil {
			log.Warn().
				Err(err).
				Str("polkitPath", polkitPath).
				Msg("Could not run setsid")
			return true, false, err
		}

		log.Debug().Msg("Checking if polkit is running after attempted launch")
		isRunning, err = isPolkitRunning()
		if err != nil {
			log.Warn().
				Err(err).
				Bool("isRunning", isRunning).
				Msgf("Could not check if polkit is running")
		}
		return true, isRunning, err
	}

	log.Debug().Msg("A polkit daemon is running")
	return true, true, nil
}

func hasHelpers() bool {
	/* TODO add polkit file too */
	_, err := bitmaskRootPath()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not find bitmask-root helper")
		return false
	}
	return true
}

func isPolkitRunning() (bool, error) {
	// TODO shouldn't we also check for polkitd running?
	var polkitProcNames = [...]string{
		"polkit-gnome-authentication-agent-1",
		"polkit-kde-auth",
		"polkit-mate-authentication-agent-1",
		"polkit-ukui-authentication-agent-1",
		"lxpolkit",
		"lxqt-policykit-agent",
		"lxsession",
		"gnome-shell",
		"gnome-flashback",
		"fingerprint-polkit-agent",
		"xfce-polkit",
		"phosh",
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

func getPolkitPath() (string, error) {
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
		// do you know some we"re still missing? please send a merge request :)
	}

	for _, polkit := range polkitPaths {
		log.Trace().Str("polkitBinary", polkit).Msg("Checking if polkit binary exists")
		_, err := os.Stat(polkit)
		if err == nil {
			log.Debug().
				Str("polkitBinary", polkit).
				Msg("Found a polkit binary")
			return polkit, nil
		}
	}
	return "", errors.New("Could not find any usable polkit binary")
}

func (l *Launcher) OpenvpnStart(flags ...string) error {
	log.Info().Msg("Starting OpenVPN")
	arg := []string{"openvpn", "start", getOpenvpnPath()}
	arg = append(arg, flags...)
	l.OpenvpnCh <- arg
	return nil
}

func (l *Launcher) OpenvpnStop() error {
	l.OpenvpnCh <- nil
	log.Info().Msg("Stopping OpenVPN")
	return runBitmaskRoot("openvpn", "stop")
}

func (l *Launcher) FirewallStart(gateways []bonafide.Gateway) error {
	log.Info().Msg("Starting firewall")
	if len(gateways) == 0 {
		log.Warn().Msg("Need atleast one gateway for firewall allow list")
	}

	for _, gw := range gateways {
		log.Debug().
			Str("gatewayIP", gw.IPAddress).
			Msg("Whitelisting gateway ip in firewall")
	}

	if os.Getenv("LEAP_DRYRUN") == "1" {
		log.Debug().Msg("Not changing firewall rules (LEAP_DRYRUN=1)")
		return nil
	}

	arg := []string{"firewall", "start"}
	for _, gw := range gateways {
		arg = append(arg, gw.IPAddress)
	}
	return runBitmaskRoot(arg...)
}

func (l *Launcher) FirewallStop() error {
	log.Info().Msg("Stopping firewall")
	return runBitmaskRoot("firewall", "stop")
}

func (l *Launcher) FirewallIsUp() bool {
	err := runBitmaskRoot("firewall", "isup")
	return err == nil
}

func (l *Launcher) openvpnRunner(arg ...string) {
	running := false
	runOpenvpn := func(arg []string) {
		for running {
			err := runBitmaskRoot(arg...)
			if err != nil {
				log.Warn().
					Err(err).
					Msg("An error ocurred running OpenVPN")
				l.OpenvpnCh <- nil
				l.Failed = true
			}
		}
	}

	for arg := range l.OpenvpnCh {
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
	log.Debug().
		Str("cmd", strings.Join(arg, " ")).
		Msg("Executing bitmask-root")

	out, err := cmd.CombinedOutput()
	// 'firewall isup' is called often and "fails" often, as the firewall is not yet up, so don't log
	outTrimed := strings.TrimRight(strings.TrimRight(string(out), "\n"), "\r")
	if err != nil && arg[2] != "isup" {
		log.Warn().
			Err(err).
			Str("cmd", strings.Join(arg, " ")).
			Msgf("Error running bitmask-root: \"%s\"", outTrimed)
	} else {
		log.Trace().
			Str("cmd", strings.Join(arg, " ")).
			Msgf("Command exited gracefully: \"%s\"", outTrimed)
	}
	return err
}

func bitmaskRootPath() (string, error) {
	if os.Getenv("SNAP") != "" {
		path := "/snap/bin/" + config.BinaryName + ".bitmask-root"
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return path, nil
		}
	}
	for _, path := range bitmaskRootPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return path, nil
		}
	}
	return "", fmt.Errorf("bitmask-root not found in %q", strings.Join(bitmaskRootPaths, ", "))
}

func getOpenvpnPath() string {
	if os.Getenv("SNAP") != "" {
		return "/snap/bin/" + config.BinaryName + ".openvpn"
	}
	return systemOpenvpnPath
}

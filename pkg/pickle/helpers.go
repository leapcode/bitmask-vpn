// Copyright (C) 2020 LEAP
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

package pickle

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/rs/zerolog/log"
)

//go:embed helpers
var helpers embed.FS

const (
	bitmaskRoot = "/usr/sbin/bitmask-root"
	// TODO parametrize this with config.appName
	policyFile = "/usr/share/polkit-1/actions/se.leap.bitmask.riseupvpn.policy"
)

func check(err error) {
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not dump helper")
	}
}

func alreadyThere(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func isRoot() bool {
	uid := os.Getuid()
	return uid == 0
}

func copyAsRoot(orig, dest string, isExec bool) {
	if alreadyThere(dest) {
		log.Info().
			Str("outFile", dest).
			Msg("Helper file already exists")
		return
	}

	cmd := exec.Command("false")
	if isRoot() {
		cmd = exec.Command("cp", orig, dest)
	} else {
		var confirm string
		log.Info().
			Str("dest", dest).
			Msg("> About to write (with sudo):\n>ok? [y/N]")
		fmt.Scanln(&confirm)
		if confirm != "y" {
			log.Warn().Msg("Aborting")
			os.Exit(1)
		}
		cmd = exec.Command("sudo", "cp", orig, dest)
	}

	err := cmd.Run()
	check(err)

	if isExec {
		if isRoot() {
			cmd = exec.Command("chmod", "755", dest)
		} else {
			cmd = exec.Command("sudo", "chmod", "755", dest)
		}
		err = cmd.Run()
		check(err)
	} else {
		if isRoot() {
			cmd = exec.Command("chmod", "644", dest)
		} else {
			cmd = exec.Command("sudo", "chmod", "644", dest)
		}
		err = cmd.Run()
		check(err)
	}

	fmt.Println("> done")
}

func dumpHelper(fname, dest string, isExec bool) {
	// TODO win/mac implementation
	if runtime.GOOS != "linux" {
		log.Debug().
			Str("os", runtime.GOOS).
			Msg("Skipping OS. bitmask-root is only supported on Linux")
		return
	}

	fd, err := helpers.Open(path.Join("helpers", fname))
	check(err)
	log.Debug().
		Str("helper", fname).
		Msg("Checking helper file")

	tmpfile, err := os.CreateTemp("/dev/shm", "*")
	check(err)
	defer os.Remove(tmpfile.Name())

	_, err = io.Copy(tmpfile, fd)
	check(err)
	copyAsRoot(tmpfile.Name(), dest, isExec)
}

func InstallHelpers() {
	// logger is not configured at this point
	config.ConfigureLogger()
	defer config.CloseLogger()

	// this  function can be called by command line argument: riseup-vpn --install-helpers
	log.Info().Msg("Installing helpers")
	dumpHelper("bitmask-root", bitmaskRoot, true)
	dumpHelper("se.leap.bitmask.policy", policyFile, false)
	log.Info().Msg("All helpers are installed")
}

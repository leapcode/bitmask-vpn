//go:generate statik -src=../../helpers -include=*

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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"

	_ "0xacab.org/leap/bitmask-vpn/pkg/pickle/statik"
	"github.com/rakyll/statik/fs"
)

const (
	bitmaskRoot = "/usr/sbin/bitmask-root"
	// TODO parametrize this with config.appName
	policyFile = "/usr/share/polkit-1/actions/se.leap.bitmask.riseupvpn.policy"
)

func check(e error) {
	if e != nil {
		panic(e)
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
		fmt.Println("> File exists: ", dest)
		return
	}

	cmd := exec.Command("false")
	if isRoot() {
		cmd = exec.Command("cp", orig, dest)
	} else {
		var confirm string
		fmt.Println("> About to write (with sudo):", dest)
		fmt.Printf("> ok? [y/N] ")
		fmt.Scanln(&confirm)
		if confirm != "y" {
			fmt.Println("aborting")
			os.Exit(1)
		}
		cmd = exec.Command("sudo", "cp", orig, dest)
	}

	err := cmd.Run()
	check(err)

	if isExec {
		if isRoot() {
			cmd = exec.Command("chmod", "776", dest)
		} else {
			cmd = exec.Command("sudo", "chmod", "776", dest)
		}
		err = cmd.Run()
		check(err)
	}

	fmt.Println("> done")
}

func dumpHelper(fname, dest string, isExec bool) {
	// TODO win/mac implementation
	if runtime.GOOS != "linux" {
		fmt.Println("Only linux supported for now")
		return
	}
	stFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	r, err := stFS.Open("/" + fname)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	c, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile("/dev/shm", "*")
	check(err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(c)
	check(err)
	copyAsRoot(tmpfile.Name(), dest, isExec)
}

func InstallHelpers() {
	dumpHelper("bitmask-root", bitmaskRoot, true)
	dumpHelper("se.leap.bitmask.policy", policyFile, false)
}

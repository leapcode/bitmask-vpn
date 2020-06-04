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
	"math/rand"
	"os"
	"os/exec"
	"time"

	_ "0xacab.org/leap/bitmask-vpn/pkg/pickle/statik"
	"github.com/rakyll/statik/fs"
)

const (
	bitmaskRoot = "/usr/sbin/bitmask-root"
	// TODO parametrize this with config.appName
	policyFile = "/usr/share/polkit-1/actions/se.leap.bitmask.riseupvpn.policy"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

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

func copyAsRoot(orig, dest string, isExec bool) {
	if alreadyThere(dest) {
		fmt.Println("> File exists: ", dest)
		return
	}
	var confirm string
	fmt.Println("> About to write (as root):", dest)
	fmt.Printf("> Continue? [y/N] ")
	fmt.Scanln(&confirm)
	if confirm != "y" {
		fmt.Println("aborting")
		os.Exit(1)
	}
	cmd := exec.Command("sudo", "cp", orig, dest)
	err := cmd.Run()
	check(err)

	if isExec {
		cmd = exec.Command("sudo", "chmod", "776", dest)
		err = cmd.Run()
		check(err)
	}

	fmt.Println("> done")
}

/* dumpHelper works in linux only at the moment.
   TODO should separate implementations by platform */
func dumpHelper(fname, dest string, isExec bool) {
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
	tmp := "/dev/shm/" + randSeq(14)

	f, err := os.Create(tmp)
	check(err)
	defer os.Remove(tmp)

	_, err = f.Write(c)
	check(err)
	copyAsRoot(tmp, dest, isExec)
}

func InstallHelpers() {
	rand.Seed(time.Now().UnixNano())
	dumpHelper("bitmask-root", bitmaskRoot, true)
	dumpHelper("se.leap.bitmask.policy", policyFile, false)
}

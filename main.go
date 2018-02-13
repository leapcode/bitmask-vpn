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

package main

import (
	"log"

	"0xacab.org/leap/bitmask-systray/bitmask"
)

const (
	provider = "riseup.net"
)

func main() {
	conf, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	notify := newNotificator(conf)

	b, err := bitmask.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	checkAndInstallHelpers(b, notify)
	run(b, conf)
}

func checkAndInstallHelpers(b *bitmask.Bitmask, notify *notificator) {
	helpers, priviledge, err := b.VPNCheck()
	if err.Error() == "nopolkit" || (err == nil && !priviledge) {
		log.Printf("No polkit found")
		notify.authAgent()
	} else if err != nil {
		log.Fatal(err)
	}

	if !helpers {
		err = b.InstallHelpers()
		if err != nil {
			log.Println("Error installing helpers: ", err)
		}
	}
}

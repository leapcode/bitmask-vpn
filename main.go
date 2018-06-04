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
	"flag"
	"fmt"
	"log"
	"os"

	"0xacab.org/leap/bitmask-systray/bitmask"
	"github.com/jmshal/go-locale"
	"golang.org/x/text/message"
)

const (
	provider        = "riseup.net"
	applicationName = "RiseupVPN"
)

var version string
var printer *message.Printer

func main() {
	versionFlag := flag.Bool("version", false, "Version of the bitmask-systray")
	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if _, err := os.Stat(bitmask.ConfigPath); os.IsNotExist(err) {
		os.MkdirAll(bitmask.ConfigPath, os.ModePerm)
	}

	err := acquirePID()
	if err != nil {
		log.Fatal(err)
	}
	defer releasePID()

	conf := parseConfig()
	initPrinter()

	notify := newNotificator(conf)

	b, err := bitmask.Init()
	if err != nil {
		log.Print(err)
		notify.bitmaskNotRunning()
		return
	}
	defer b.Close()
	go checkAndStartBitmask(b, notify, conf)

	run(b, conf, notify)
}

func checkAndStartBitmask(b *bitmask.Bitmask, notify *notificator, conf *systrayConfig) {
	err := checkAndInstallHelpers(b, notify)
	if err != nil {
		log.Printf("Is bitmask running? %v", err)
		os.Exit(1)
	}
	maybeStartVPN(b, conf)
}

func checkAndInstallHelpers(b *bitmask.Bitmask, notify *notificator) error {
	helpers, priviledge, err := b.VPNCheck()
	if (err != nil && err.Error() == "nopolkit") || (err == nil && !priviledge) {
		log.Printf("No polkit found")
		notify.authAgent()
	} else if err != nil {
		notify.bitmaskNotRunning()
		return err
	}

	if !helpers {
		err = b.InstallHelpers()
		if err != nil {
			log.Println("Error installing helpers: ", err)
		}
	}
	return nil
}

func maybeStartVPN(b *bitmask.Bitmask, conf *systrayConfig) {
	if conf.UserStoppedVPN {
		return
	}

	err := b.StartVPN(provider)
	if err != nil {
		log.Println("Error starting VPN: ", err)
	}
	conf.setUserStoppedVPN(false)
}

func initPrinter() {
	locale, err := go_locale.DetectLocale()
	if err != nil {
		log.Println("Error detecting the system locale: ", err)
	}

	printer = message.NewPrinter(message.MatchLanguage(locale, "en"))
}

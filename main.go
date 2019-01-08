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
	"io"
	"log"
	"os"
	"path"
	"runtime"

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
	// on OSX sometimes the systray doesn't work (bitmask-systray#52)
	// locking the main thread into an OS thread fixes the problem
	runtime.LockOSThread()

	logger, err := configureLogger()
	if err != nil {
		log.Println("Can't configure logger: ", err)
	} else {
		defer logger.Close()
	}

	conf := parseConfig()
	initPrinter()

	flag.BoolVar(&conf.SelectGateway, "select-gateway", false, "Enable gateway selection")
	versionFlag := flag.Bool("version", false, "Version of the bitmask-systray")
	flag.Parse()
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	bt := bmTray{conf: conf}
	go initialize(conf, &bt)
	bt.start()
}

func initialize(conf *systrayConfig, bt *bmTray) {
	if _, err := os.Stat(bitmask.ConfigPath); os.IsNotExist(err) {
		os.MkdirAll(bitmask.ConfigPath, os.ModePerm)
	}

	err := acquirePID()
	if err != nil {
		log.Fatal(err)
	}
	defer releasePID()

	notify := newNotificator(conf)

	b, err := initBitmask()
	if err != nil {
		notify.initFailure(err)
		return
	}
	defer b.Close()
	go checkAndStartBitmask(b, notify, conf)
	go listenSignals(b)

	as := newAutostart(applicationName, getIconPath())
	err = as.Enable()
	if err != nil {
		log.Printf("Error enabling autostart: %v", err)
	}
	bt.loop(b, notify, as)
}

func checkAndStartBitmask(b bitmask.Bitmask, notify *notificator, conf *systrayConfig) {
	err := checkAndInstallHelpers(b, notify)
	if err != nil {
		log.Printf("Is bitmask running? %v", err)
		os.Exit(1)
	}
	err = maybeStartVPN(b, conf)
	if err != nil {
		log.Println("Error starting VPN: ", err)
		notify.errorStartingVPN(err)
	}
}

func checkAndInstallHelpers(b bitmask.Bitmask, notify *notificator) error {
	helpers, priviledge, err := b.VPNCheck()
	if (err != nil && err.Error() == "nopolkit") || (err == nil && !priviledge) {
		log.Printf("No polkit found")
		notify.authAgent()
	} else if err != nil {
		log.Printf("Error checking vpn: %v", err)
		notify.errorStartingVPN(err)
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

func maybeStartVPN(b bitmask.Bitmask, conf *systrayConfig) error {
	if conf.UserStoppedVPN {
		return nil
	}

	err := b.StartVPN(provider)
	conf.setUserStoppedVPN(false)
	return err
}

func configureLogger() (io.Closer, error) {
	logFile, err := os.OpenFile(path.Join(bitmask.ConfigPath, "systray.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	}
	return logFile, err
}

func initPrinter() {
	locale, err := go_locale.DetectLocale()
	if err != nil {
		log.Println("Error detecting the system locale: ", err)
	}

	printer = message.NewPrinter(message.MatchLanguage(locale, "en"))
}

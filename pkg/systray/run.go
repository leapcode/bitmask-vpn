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

package systray

import (
	"log"
	"os"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

func Run(conf *Config) {
	bt := bmTray{conf: conf}
	go initialize(conf, &bt)
	bt.start()
}

func initialize(conf *Config, bt *bmTray) {
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		os.MkdirAll(config.Path, os.ModePerm)
	}

	err := acquirePID()
	if err != nil {
		log.Fatal(err)
	}
	defer releasePID()

	notify := newNotificator(conf)

	b, err := bitmask.Init(conf.Printer)
	if err != nil {
		notify.initFailure(err)
		return
	}
	defer b.Close()
	go checkAndStartBitmask(b, notify, conf)
	go listenSignals(b)

	as := bitmask.NewAutostart(config.ApplicationName, getIconPath())
	err = as.Enable()
	if err != nil {
		log.Printf("Error enabling autostart: %v", err)
	}
	bt.loop(b, notify, as)
}

func checkAndStartBitmask(b bitmask.Bitmask, notify *notificator, conf *Config) {
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

func maybeStartVPN(b bitmask.Bitmask, conf *Config) error {
	if conf.UserStoppedVPN {
		return nil
	}

	err := b.StartVPN(config.Provider)
	conf.setUserStoppedVPN(false)
	return err
}

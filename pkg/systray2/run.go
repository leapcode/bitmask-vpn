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

/*
func initialize(conf *Config, bt *bmTray, finishedCh chan bool) {
	defer func() { finishedCh <- true }()
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		os.MkdirAll(config.Path, os.ModePerm)
	}

	err := acquirePID()
	if err != nil {
		log.Fatal(err)
	}
	defer releasePID()

	b, err := bitmask.Init(conf.Printer)
	if err != nil {
		// TODO notify failure
		return
	}
	defer b.Close()
	go checkAndStartBitmask(b, conf)
	go listenSignals(b)

	var as bitmask.Autostart
	if conf.DisableAustostart {
		as = &bitmask.DummyAutostart{}
	} else {
		as = bitmask.NewAutostart(config.ApplicationName, "")
	}
	err = as.Enable()
	if err != nil {
		log.Printf("Error enabling autostart: %v", err)
	}
}
*/

func checkAndStartBitmask(b bitmask.Bitmask, conf *Config) {
	if conf.Obfs4 {
		err := b.UseTransport("obfs4")
		if err != nil {
			log.Printf("Error setting transport: %v", err)
		}
	}
	err := checkAndInstallHelpers(b)
	if err != nil {
		log.Printf("Is bitmask running? %v", err)
		os.Exit(1)
	}
	err = maybeStartVPN(b, conf)
	if err != nil {
		log.Println("Error starting VPN: ", err)
	}
}

func checkAndInstallHelpers(b bitmask.Bitmask) error {
	helpers, priviledge, err := b.VPNCheck()
	if (err != nil && err.Error() == "nopolkit") || (err == nil && !priviledge) {
		log.Printf("No polkit found")
		os.Exit(1)
	} else if err != nil {
		log.Printf("Error checking vpn: %v", err)
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
	if !conf.StartVPN {
		return nil
	}

	err := b.StartVPN(config.Provider)
	conf.setUserStoppedVPN(false)
	return err
}

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

package bitmask

import (
	"errors"
	"log"
	"os"
	"path"

	"github.com/jmshal/go-locale"
	"golang.org/x/text/message"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/pid"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn"
)

type ProviderInfo struct {
	Provider string
	AppName  string
}

func GetConfiguredProvider() *ProviderInfo {
	provider := config.Provider
	appName := config.ApplicationName
	return &ProviderInfo{provider, appName}
}

func InitializeLogger() {
	_, err := config.ConfigureLogger(path.Join(config.LogPath))
	if err != nil {
		log.Println("Can't configure logger: ", err)
	}
}

func initBitmask(printer *message.Printer) (Bitmask, error) {
	b, err := vpn.Init()
	if err != nil {
		log.Printf("An error ocurred starting bitmask: %v", err)
		err = errors.New(printer.Sprintf(errorMsg, err))
	}
	return b, err
}

func InitializeBitmask() (Bitmask, error) {
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		os.MkdirAll(config.Path, os.ModePerm)
	}

	err := pid.AcquirePID()
	if err != nil {
		log.Fatal(err)
	}
	defer pid.ReleasePID()

	conf := config.ParseConfig()
	conf.Version = "unknown"
	conf.Printer = initPrinter()

	b, err := initBitmask(conf.Printer)
	if err != nil {
		// TODO notify failure
		log.Fatal(err)
	}
	go checkAndStartBitmask(b, conf)

	var as Autostart
	if conf.DisableAustostart {
		as = &dummyAutostart{}
	} else {
		as = newAutostart(config.ApplicationName, "")
	}
	err = as.Enable()
	if err != nil {
		log.Printf("Error enabling autostart: %v", err)
	}
	return b, nil
}

func initPrinter() *message.Printer {
	locale, err := go_locale.DetectLocale()
	if err != nil {
		log.Println("Error detecting the system locale: ", err)
	}

	return message.NewPrinter(message.MatchLanguage(locale, "en"))
}

func checkAndStartBitmask(b Bitmask, conf *config.Config) {
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

func checkAndInstallHelpers(b Bitmask) error {
	helpers, privilege, err := b.VPNCheck()
	if (err != nil && err.Error() == "nopolkit") || (err == nil && !privilege) {
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

func maybeStartVPN(b Bitmask, conf *config.Config) error {
	if !conf.StartVPN {
		return nil
	}

	err := b.StartVPN(config.Provider)
	conf.SetUserStoppedVPN(false)
	return err
}

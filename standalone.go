// +build !bitmaskd
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
	"errors"
	"log"
	"os"

	"0xacab.org/leap/bitmask-systray/bitmask"
	standalone "0xacab.org/leap/bitmask-systray/standalone"
	pmautostart "github.com/ProtonMail/go-autostart"
)

const (
	errorMsg = `An error has ocurred initializing %s: %v`
)

func initBitmask() (bitmask.Bitmask, error) {
	b, err := standalone.Init()
	if err != nil {
		log.Printf("An error ocurred starting standalone bitmask: %v", err)
		err = errors.New(printer.Sprintf(errorMsg, applicationName, err))
	}
	return b, err
}

func newAutostart(appName string, iconPath string) autostart {
	return &pmautostart.App{
		Name:        appName,
		Exec:        os.Args,
		DisplayName: appName,
		Icon:        iconPath,
	}
}

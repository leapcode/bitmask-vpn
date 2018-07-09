// +build standalone
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
	"os"

	"0xacab.org/leap/bitmask-systray/bitmask"
	standalone "0xacab.org/leap/bitmask-systray/standalone"
	pmautostart "github.com/ProtonMail/go-autostart"
)

func initBitmask() (bitmask.Bitmask, error) {
	return standalone.Init()
}

func newAutostart(appName string, iconPath string) autostart {
	return &pmautostart.App{
		Name:        appName,
		Exec:        os.Args,
		DisplayName: appName,
		Icon:        iconPath,
	}
}

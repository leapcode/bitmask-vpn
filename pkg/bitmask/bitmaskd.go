// +build bitmaskd
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

package bitmask

import (
	"errors"
	"log"

	"0xacab.org/leap/bitmask-systray/pkg/bitmaskd"
	"golang.org/x/text/message"
)

const (
	notRunning = `Is bitmaskd running? Start bitmask and try again.`
)

// Init bitmask
func Init(printer *message.Printer) (Bitmask, error) {
	b, err := bitmaskd.Init()
	if err != nil {
		log.Printf("An error ocurred starting bitmaskd: %v", err)
		err = errors.New(printer.Sprintf(notRunning))
	}
	return b, err
}

// NewAutostart creates a handler for the autostart of your platform
func NewAutostart(appName string, iconPath string) Autostart {
	return &dummyAutostart{}
}

// SPDX-FileCopyrightText: 2018 LEAP
// SPDX-License-Identifier: GPL-3.0-or-later
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
	"os/signal"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	logFile = "systray.log"
)

var version string

func main() {
	displayVersion := flag.Bool("version", false, "Display the version")
	flag.Parse()

	if *displayVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	start()
}

func start() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	b, err := bitmask.Init(message.NewPrinter(language.English))
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	err = b.StartVPN(config.Provider)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	<-signalCh
}

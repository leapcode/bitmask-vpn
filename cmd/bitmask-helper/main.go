// SPDX-FileCopyrightText: 2018-2020 LEAP
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
	"log"
	"path"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/helper"
)

const (
	preferredPort = 7171
	logFile       = "helper.log"
)

var version string

func main() {
	logger, err := config.ConfigureLogger(path.Join(helper.LogFolder, logFile))
	if err != nil {
		log.Println("Can't configure logger: ", err)
	} else {
		defer logger.Close()
	}
	config.Version = version

	// StartHelper is the main entry point - it also handles cli args in windows, and starts the http server.
	helper.StartHelper(preferredPort)
}

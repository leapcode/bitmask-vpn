// +build !windows,!darwin
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
	"os"

	"github.com/pebbe/zmq4"
)

const coreEndpoint = "ipc:///var/tmp/bitmask.core.sock"

var ConfigPath = os.Getenv("HOME") + "/.config/leap"

func hasCurve() bool {
	return zmq4.HasCurve()
}

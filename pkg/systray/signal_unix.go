// +build !windows
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
	"os"
	"os/signal"
	"syscall"

	"0xacab.org/leap/bitmask-systray/pkg/bitmask"
)

func listenSignals(bm bitmask.Bitmask) {
	sigusrCh := make(chan os.Signal, 1)
	signal.Notify(sigusrCh, syscall.SIGUSR1)

	for range sigusrCh {
		bm.ReloadFirewall()
	}
}

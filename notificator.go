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
	"time"

	"0xacab.org/leap/go-dialog"
	"github.com/skratchdot/open-golang/open"
)

const (
	donationText = `The %s service is expensive to run. Because we don't want to store personal information about you, there is no accounts or billing for this service. But if you want the service to continue, donate at least $5 each month.
	
Do you want to donate now?`
	missingAuthAgent = `Could not find a polkit authentication agent. Please run one and try again.`
	notRunning       = `Is bitmaskd running? Start bitmask and try again.`
)

type notificator struct {
	conf *systrayConfig
}

func newNotificator(conf *systrayConfig) *notificator {
	n := notificator{conf}
	go n.donations()
	return &n
}

func (n *notificator) donations() {
	time.Sleep(time.Minute * 5)
	for {
		if n.conf.needsNotification() {
			letsDonate := dialog.Message(printer.Sprintf(donationText, applicationName)).Title(printer.Sprintf("Donate")).YesNo()
			n.conf.setNotification()
			if letsDonate {
				open.Run("https://riseup.net/donate-vpn")
				n.conf.setDonated()
			}
		}
		time.Sleep(time.Hour)
	}
}

func (n *notificator) bitmaskNotRunning() {
	dialog.Message(printer.Sprintf(notRunning)).Title(printer.Sprintf("Can't contact bitmask")).Error()
}

func (n *notificator) authAgent() {
	dialog.Message(printer.Sprintf(missingAuthAgent)).Title(printer.Sprintf("Missing authentication agent")).Error()
}

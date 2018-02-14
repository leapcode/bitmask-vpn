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
	"path"
	"time"

	notif "github.com/0xAX/notificator"
)

const (
	donationText     = `The RiseupVPN service is expensive to run. Because we don't want to store personal information about you, there is no accounts or billing for this service. But if you want the service to continue, donate at least $5 each month at https://riseup.net/donate-vpn`
	missingAuthAgent = `Could not find a polkit authentication agent. Please run one and try again.`
	notRunning       = `Is bitmaskd running? Start bitmask and try again.`
)

type notificator struct {
	notify *notif.Notificator
	conf   *systrayConfig
}

func newNotificator(conf *systrayConfig) *notificator {
	wd, _ := os.Getwd()
	notify := notif.New(notif.Options{
		DefaultIcon: path.Join(wd, "riseupvpn.svg"),
		AppName:     "RiseupVPN",
	})
	n := notificator{notify, conf}
	go n.donations()
	return &n
}

func (n *notificator) donations() {
	time.Sleep(time.Minute * 5)
	for {
		if n.conf.needsNotification() {
			n.notify.Push("Donate to RiseupVPN", donationText, "", notif.UR_NORMAL)
			n.conf.setNotification()
		}
		time.Sleep(time.Hour)
	}
}

func (n *notificator) bitmaskNotRunning() {
	n.notify.Push("Can't contact bitmask", notRunning, "", notif.UR_CRITICAL)
}

func (n *notificator) authAgent() {
	n.notify.Push("Missing authentication agent", missingAuthAgent, "", notif.UR_CRITICAL)
}

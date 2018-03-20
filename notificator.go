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
	svgFileName      = "riseupvpn.svg"
)

type notificator struct {
	notify *notif.Notificator
	conf   *systrayConfig
}

func newNotificator(conf *systrayConfig) *notificator {
	notify := notif.New(notif.Options{
		DefaultIcon: getSVGPath(),
		AppName:     "RiseupVPN",
	})
	n := notificator{notify, conf}
	//go n.donations()
	return &n
}

func (n *notificator) donations() {
	time.Sleep(time.Minute * 5)
	for {
		if n.conf.needsNotification() {
			n.notify.Push(printer.Sprintf("Donate to RiseupVPN"), printer.Sprintf(donationText), "", notif.UR_NORMAL)
			n.conf.setNotification()
		}
		time.Sleep(time.Hour)
	}
}

func (n *notificator) bitmaskNotRunning() {
	n.notify.Push(printer.Sprintf("Can't contact bitmask"), printer.Sprintf(notRunning), "", notif.UR_CRITICAL)
}

func (n *notificator) authAgent() {
	n.notify.Push(printer.Sprintf("Missing authentication agent"), printer.Sprintf(missingAuthAgent), "", notif.UR_CRITICAL)
}

func getSVGPath() string {
	wd, _ := os.Getwd()
	svgPath := path.Join(wd, svgFileName)
	if fileExist(svgPath) {
		return svgPath
	}

	svgPath = "/usr/share/riseupvpn/riseupvpn.svg"
	if fileExist(svgPath) {
		return svgPath
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = path.Join(os.Getenv("HOME"), "go")
	}
	svgPath = path.Join(gopath, "src", "0xacab.org", "leap", "bitmask-systray", svgFileName)
	if fileExist(svgPath) {
		return svgPath
	}

	return ""
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

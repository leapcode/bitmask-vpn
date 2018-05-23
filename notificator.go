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

	"0xacab.org/leap/go-dialog"
	"github.com/skratchdot/open-golang/open"
)

const (
	donationText = `The %s service is expensive to run. Because we don't want to store personal information about you, there is no accounts or billing for this service. But if you want the service to continue, donate at least $5 each month.
	
Do you want to donate now?`
	aboutText = `%s is an easy, fast, and secure VPN service from riseup.net. %s does not require a user account, keep logs, or track you in any way.
	    
This service paid for entirely by donations from users like you. Please donate at riseup.net/vpn/donate.
		
By using this application, you agree to the Terms of Service available at riseup.net/tos. This service is provide as-is, without any warranty, and is intended for people who work to make the world a better place.`
	missingAuthAgent = `Could not find a polkit authentication agent. Please run one and try again.`
	notRunning       = `Is bitmaskd running? Start bitmask and try again.`
	svgFileName      = "riseupvpn.svg"
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
			letsDonate := dialog.Message(printer.Sprintf(donationText, applicationName)).
				Title(printer.Sprintf("Donate")).
				Icon(getSVGPath()).
				YesNo()
			n.conf.setNotification()
			if letsDonate {
				open.Run("https://riseup.net/donate-vpn")
				n.conf.setDonated()
			}
		}
		time.Sleep(time.Hour)
	}
}

func (n *notificator) about() {
	dialog.Message(printer.Sprintf(aboutText, applicationName, applicationName)).
		Title(printer.Sprintf("About")).
		Icon(getSVGPath()).
		Info()
}

func (n *notificator) bitmaskNotRunning() {
	dialog.Message(printer.Sprintf(notRunning)).
		Title(printer.Sprintf("Can't contact bitmask")).
		Icon(getSVGPath()).
		Error()
}

func (n *notificator) authAgent() {
	dialog.Message(printer.Sprintf(missingAuthAgent)).
		Title(printer.Sprintf("Missing authentication agent")).
		Icon(getSVGPath()).
		Error()
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

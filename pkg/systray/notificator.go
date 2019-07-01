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
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/go-dialog"
	"github.com/skratchdot/open-golang/open"
)

const (
	donationText = `The %s service is expensive to run. Because we don't want to store personal information about you, there are no accounts or billing for this service. But if you want the service to continue, donate at least $5 each month.
	
Do you want to donate now?`
	aboutText = `%[1]s is an easy, fast, and secure VPN service from %[2]s. %[1]s does not require a user account, keep logs, or track you in any way.
	    
This service is paid for entirely by donations from users like you. Please donate at %[3]s.
		
By using this application, you agree to the Terms of Service available at %[4]s. This service is provided as-is, without any warranty, and is intended for people who work to make the world a better place.


%[1]v version: %[5]s`
	missingAuthAgent = `Could not find a polkit authentication agent. Please run one and try again.`
	errorStartingVPN = `Can't connect to %s: %v`
	svgFileName      = "icon.svg"
)

type notificator struct {
	conf *Config
}

func newNotificator(conf *Config) *notificator {
	n := notificator{conf}
	go n.donations()
	return &n
}

func (n *notificator) donations() {
	for {
		time.Sleep(time.Hour)
		if n.conf.needsNotification() {
			letsDonate := dialog.Message(n.conf.Printer.Sprintf(donationText, config.ApplicationName)).
				Title(n.conf.Printer.Sprintf("Donate")).
				Icon(getIconPath()).
				YesNo()
			n.conf.setNotification()
			if letsDonate {
				open.Run(config.DonateURL)
				n.conf.setDonated()
			}
		}
	}
}

func (n *notificator) about(version string) {
	if version == "" && os.Getenv("SNAP") != "" {
		_version, err := ioutil.ReadFile(os.Getenv("SNAP") + "/snap/version.txt")
		if err == nil {
			version = string(_version)
		}
	}
	dialog.Message(n.conf.Printer.Sprintf(aboutText, config.ApplicationName, config.Provider, config.DonateURL, config.TosURL, version)).
		Title(n.conf.Printer.Sprintf("About")).
		Icon(getIconPath()).
		Info()
}

func (n *notificator) initFailure(err error) {
	dialog.Message(err.Error()).
		Title(n.conf.Printer.Sprintf("Initialization error")).
		Icon(getIconPath()).
		Error()
}

func (n *notificator) authAgent() {
	dialog.Message(n.conf.Printer.Sprintf(missingAuthAgent)).
		Title(n.conf.Printer.Sprintf("Missing authentication agent")).
		Icon(getIconPath()).
		Error()
}

func (n *notificator) errorStartingVPN(err error) {
	dialog.Message(n.conf.Printer.Sprintf(errorStartingVPN, config.ApplicationName, err)).
		Title(n.conf.Printer.Sprintf("Error starting VPN")).
		Icon(getIconPath()).
		Error()
}

func getIconPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = path.Join(os.Getenv("HOME"), "go")
	}

	if runtime.GOOS == "windows" {
		icoPath := `C:\Program Files\` + config.ApplicationName + `\icon.ico`
		if fileExist(icoPath) {
			return icoPath
		}
		icoPath = path.Join(gopath, "src", "0xacab.org", "leap", "riseup_vpn", "assets", "riseupvpn.ico")
		if fileExist(icoPath) {
			return icoPath
		}
		return ""
	}

	if runtime.GOOS == "darwin" {
		icnsPath := "/Applications/" + config.ApplicationName + ".app/Contents/Resources/app.icns"
		if fileExist(icnsPath) {
			return icnsPath
		}
		icnsPath = path.Join(gopath, "src", "0xacab.org", "leap", "riseup_vpn", "assets", "riseupvpn.icns")
		if fileExist(icnsPath) {
			return icnsPath
		}
		return ""
	}

	snapPath := os.Getenv("SNAP")
	if snapPath != "" {
		return snapPath + "/snap/meta/gui/icon.svg"
	}

	wd, _ := os.Getwd()
	svgPath := path.Join(wd, svgFileName)
	if fileExist(svgPath) {
		return svgPath
	}

	svgPath = "/usr/share/" + config.BinaryName + "/icon.svg"
	if fileExist(svgPath) {
		return svgPath
	}

	svgPath = path.Join(gopath, "src", "0xacab.org", "leap", "bitmask-vpn", svgFileName)
	if fileExist(svgPath) {
		return svgPath
	}

	return ""
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

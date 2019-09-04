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
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"0xacab.org/leap/bitmask-vpn/icon"
	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

type bmTray struct {
	bm            bitmask.Bitmask
	conf          *Config
	notify        *notificator
	waitCh        chan bool
	mStatus       *systray.MenuItem
	mTurnOn       *systray.MenuItem
	mTurnOff      *systray.MenuItem
	mHelp         *systray.MenuItem
	mDonate       *systray.MenuItem
	mAbout        *systray.MenuItem
	mQuit         *systray.MenuItem
	activeGateway *gatewayTray
	autostart     bitmask.Autostart
}

type gatewayTray struct {
	menuItem *systray.MenuItem
	name     string
}

func (bt *bmTray) start() {
	// XXX this removes the snap error message, but produces an invisible icon.
	// https://0xacab.org/leap/riseup_vpn/issues/44
	// os.Setenv("TMPDIR", "/var/tmp")
	systray.Run(bt.onReady, bt.onExit)
}

func (bt *bmTray) quit() {
	systray.Quit()
}

func (bt *bmTray) onExit() {
	log.Println("Closing systray")
}

func (bt *bmTray) onReady() {
	printer := bt.conf.Printer
	systray.SetIcon(icon.Off)

	bt.mStatus = systray.AddMenuItem(printer.Sprintf("Checking status..."), "")
	bt.mStatus.Disable()
	bt.waitCh <- true
}

func (bt *bmTray) setUpSystray() {
	printer := bt.conf.Printer
	bt.mTurnOn = systray.AddMenuItem(printer.Sprintf("Turn on"), "")
	bt.mTurnOn.Hide()
	bt.mTurnOff = systray.AddMenuItem(printer.Sprintf("Turn off"), "")
	bt.mTurnOff.Hide()
	systray.AddSeparator()

	if bt.conf.SelectGateway {
		bt.addGateways()
	}

	bt.mHelp = systray.AddMenuItem(printer.Sprintf("Help..."), "")
	bt.mDonate = systray.AddMenuItem(printer.Sprintf("Donate..."), "")
	bt.mAbout = systray.AddMenuItem(printer.Sprintf("About..."), "")
	systray.AddSeparator()

	bt.mQuit = systray.AddMenuItem(printer.Sprintf("Quit"), "")
}

func (bt *bmTray) loop(bm bitmask.Bitmask, notify *notificator, as bitmask.Autostart) {
	<-bt.waitCh
	bt.waitCh = nil

	bt.bm = bm
	bt.notify = notify
	bt.autostart = as
	bt.setUpSystray()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	ch := bt.bm.GetStatusCh()
	if status, err := bt.bm.GetStatus(); err != nil {
		log.Printf("Error getting status: %v", err)
	} else {
		bt.changeStatus(status)
	}

	for {
		select {
		case status := <-ch:
			log.Println("status: " + status)
			bt.changeStatus(status)

		case <-bt.mTurnOn.ClickedCh:
			log.Println("on")
			bt.changeStatus("starting")
			bt.bm.StartVPN(config.Provider)
			bt.conf.setUserStoppedVPN(false)
		case <-bt.mTurnOff.ClickedCh:
			log.Println("off")
			bt.changeStatus("stopping")
			bt.bm.StopVPN()
			bt.conf.setUserStoppedVPN(true)

		case <-bt.mHelp.ClickedCh:
			open.Run(config.HelpURL)
		case <-bt.mDonate.ClickedCh:
			bt.conf.setDonated()
			open.Run(config.DonateURL)
		case <-bt.mAbout.ClickedCh:
			bitmaskVersion, err := bt.bm.Version()
			versionStr := bt.conf.Version
			if err != nil {
				log.Printf("Error getting version: %v", err)
			} else if bitmaskVersion != "" {
				versionStr = fmt.Sprintf("%s (bitmaskd %s)", bt.conf.Version, bitmaskVersion)
			}
			go bt.notify.about(versionStr)

		case <-bt.mQuit.ClickedCh:
			err := bt.autostart.Disable()
			if err != nil {
				log.Printf("Error disabling autostart: %v", err)
			}
			/* we return and leave bt.quit() to the caller */
			return
		case <-signalCh:
			/* we return and leave bt.quit() to the caller */
			return

		case <-time.After(5 * time.Second):
			if status, err := bt.bm.GetStatus(); err != nil {
				log.Printf("Error getting status: %v", err)
			} else {
				bt.changeStatus(status)
			}
		}
	}
}

func (bt *bmTray) addGateways() {
	gatewayList, err := bt.bm.ListGateways(config.Provider)
	if err != nil {
		log.Printf("Gateway initialization error: %v", err)
		return
	}

	mGateway := systray.AddMenuItem(bt.conf.Printer.Sprintf("Route traffic through:"), "")
	mGateway.Disable()
	for i, city := range gatewayList {
		menuItem := systray.AddMenuItem(city, bt.conf.Printer.Sprintf("Use %s %v gateway", config.ApplicationName, city))
		gateway := gatewayTray{menuItem, city}

		if i == 0 {
			menuItem.Check()
			menuItem.SetTitle("*" + city)
			bt.activeGateway = &gateway
		} else {
			menuItem.Uncheck()
		}

		go func(gateway gatewayTray) {
			for {
				<-menuItem.ClickedCh
				gateway.menuItem.SetTitle("*" + gateway.name)
				gateway.menuItem.Check()

				bt.activeGateway.menuItem.Uncheck()
				bt.activeGateway.menuItem.SetTitle(bt.activeGateway.name)
				bt.activeGateway = &gateway

				bt.bm.UseGateway(gateway.name)
				log.Printf("Manual connection to %s gateway\n", gateway.name)
				bt.bm.StartVPN(config.Provider)
			}
		}(gateway)
	}

	systray.AddSeparator()
}

func (bt *bmTray) changeStatus(status string) {
	printer := bt.conf.Printer
	if bt.waitCh != nil {
		bt.waitCh <- true
		bt.waitCh = nil
	}

	var statusStr string
	switch status {
	case "on":
		systray.SetIcon(icon.On)
		bt.mTurnOff.SetTitle(printer.Sprintf("Turn off"))
		statusStr = printer.Sprintf("%s on", config.ApplicationName)
		bt.mTurnOn.Hide()
		bt.mTurnOff.Show()

	case "off":
		systray.SetIcon(icon.Off)
		bt.mTurnOn.SetTitle(printer.Sprintf("Turn on"))
		statusStr = printer.Sprintf("%s off", config.ApplicationName)
		bt.mTurnOn.Show()
		bt.mTurnOff.Hide()

	case "starting":
		bt.waitCh = make(chan bool)
		go bt.waitIcon()
		bt.mTurnOff.SetTitle(printer.Sprintf("Cancel"))
		statusStr = printer.Sprintf("Connecting to %s", config.ApplicationName)
		bt.mTurnOn.Hide()
		bt.mTurnOff.Show()

	case "stopping":
		bt.waitCh = make(chan bool)
		go bt.waitIcon()
		statusStr = printer.Sprintf("Stopping %s", config.ApplicationName)
		bt.mTurnOn.Hide()
		bt.mTurnOff.Hide()

	case "failed":
		systray.SetIcon(icon.Blocked)
		bt.mTurnOn.SetTitle(printer.Sprintf("Reconnect"))
		bt.mTurnOff.SetTitle(printer.Sprintf("Turn off"))
		statusStr = printer.Sprintf("%s blocking internet", config.ApplicationName)
		bt.mTurnOn.Show()
		bt.mTurnOff.Show()
	}

	systray.SetTooltip(statusStr)
	bt.mStatus.SetTitle(statusStr)
}

func (bt *bmTray) waitIcon() {
	icons := [][]byte{icon.Wait0, icon.Wait1, icon.Wait2, icon.Wait3}
	for i := 0; true; i = (i + 1) % 4 {
		systray.SetIcon(icons[i])

		select {
		case <-bt.waitCh:
			return
		case <-time.After(time.Millisecond * 500):
			continue
		}
	}
}

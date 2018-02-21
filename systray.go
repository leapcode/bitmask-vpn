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
	"log"
	"os"
	"time"

	"0xacab.org/leap/bitmask-systray/bitmask"
	"0xacab.org/leap/bitmask-systray/icon"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

type bmTray struct {
	bm            *bitmask.Bitmask
	conf          *systrayConfig
	waitCh        chan bool
	mStatus       *systray.MenuItem
	mTurnOn       *systray.MenuItem
	mTurnOff      *systray.MenuItem
	mDonate       *systray.MenuItem
	mHaveDonated  *systray.MenuItem
	mCancel       *systray.MenuItem
	activeGateway *gatewayTray
}

type gatewayTray struct {
	menuItem *systray.MenuItem
	name     string
}

func run(bm *bitmask.Bitmask, conf *systrayConfig) {
	bt := bmTray{bm: bm, conf: conf}
	systray.Run(bt.onReady, bt.onExit)
}

func (bt bmTray) onExit() {
	// TODO: this doesn't get executed :(
	log.Println("Finished onExit")
}

func (bt *bmTray) onReady() {
	systray.SetIcon(icon.Off)

	bt.mStatus = systray.AddMenuItem(printer.Sprintf("Checking status..."), "")
	bt.mStatus.Disable()
	bt.mTurnOn = systray.AddMenuItem(printer.Sprintf("Turn on"), printer.Sprintf("Turn RiseupVPN on"))
	go bt.mTurnOn.Hide()
	bt.mTurnOff = systray.AddMenuItem(printer.Sprintf("Turn off"), printer.Sprintf("Turn RiseupVPN off"))
	go bt.mTurnOff.Hide()
	bt.mCancel = systray.AddMenuItem(printer.Sprintf("Cancel"), printer.Sprintf("Cancel connection to RiseupVPN"))
	go bt.mCancel.Hide()
	systray.AddSeparator()

	if bt.conf.SelectWateway {
		bt.addGateways()
	}

	mHelp := systray.AddMenuItem(printer.Sprintf("Help ..."), "")
	bt.mDonate = systray.AddMenuItem(printer.Sprintf("Donate ..."), "")
	bt.mHaveDonated = systray.AddMenuItem(printer.Sprintf("... I have donated"), "")
	mAbout := systray.AddMenuItem(printer.Sprintf("About ..."), "")
	systray.AddSeparator()

	mQuit := systray.AddMenuItem(printer.Sprintf("Quit"), printer.Sprintf("Quit BitmaskVPN"))

	go func() {
		ch := bt.bm.GetStatusCh()
		if status, err := bt.bm.GetStatus(); err != nil {
			log.Printf("Error getting status: %v", err)
		} else {
			bt.changeStatus(status)
		}

		for {
			bt.updateDonateMenu()

			select {
			case status := <-ch:
				log.Println("status: " + status)
				bt.changeStatus(status)

			case <-bt.mTurnOn.ClickedCh:
				log.Println("on")
				bt.bm.StartVPN(provider)
			case <-bt.mTurnOff.ClickedCh:
				log.Println("off")
				bt.bm.StopVPN()
			case <-bt.mCancel.ClickedCh:
				log.Println("cancel")
				bt.bm.StopVPN()

			case <-mHelp.ClickedCh:
				open.Run("https://riseup.net/vpn")
			case <-bt.mDonate.ClickedCh:
				open.Run("https://riseup.net/donate-vpn")
			case <-bt.mHaveDonated.ClickedCh:
				bt.conf.setDonated()
			case <-mAbout.ClickedCh:
				open.Run("https://bitmask.net")

			case <-mQuit.ClickedCh:
				systray.Quit()
				bt.bm.Close()
				log.Println("Quit now...")
				os.Exit(0)

			case <-time.After(time.Minute * 30):
			}
		}
	}()
}

func (bt *bmTray) addGateways() {
	gatewayList, err := bt.bm.ListGateways(provider)
	if err != nil {
		log.Printf("Gateway initialization error: %v", err)
		return
	}

	mGateway := systray.AddMenuItem(printer.Sprintf("Route traffic through"), "")
	mGateway.Disable()
	for i, name := range gatewayList {
		menuItem := systray.AddMenuItem(name, printer.Sprintf("Use RiseupVPN %v gateway", name))
		gateway := gatewayTray{menuItem, name}

		if i == 0 {
			menuItem.Check()
			menuItem.SetTitle("*" + name)
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
			}
		}(gateway)
	}

	systray.AddSeparator()
}

func (bt *bmTray) changeStatus(status string) {
	// TODO: ugly hacks with 'go' to hide/show
	statusStr := status
	bt.mTurnOn.SetTitle(printer.Sprintf("Turn on"))
	if bt.waitCh != nil {
		bt.waitCh <- true
		bt.waitCh = nil
	}

	switch status {
	case "on":
		systray.SetIcon(icon.On)
		go bt.mTurnOn.Hide()
		go bt.mTurnOff.Show()
		go bt.mCancel.Hide()

	case "off":
		systray.SetIcon(icon.Off)
		go bt.mTurnOn.Show()
		go bt.mTurnOff.Hide()
		go bt.mCancel.Hide()

	case "starting":
		bt.waitCh = make(chan bool)
		go bt.waitIcon()
		go bt.mTurnOn.Hide()
		go bt.mTurnOff.Hide()
		go bt.mCancel.Show()

	case "stopping":
		bt.waitCh = make(chan bool)
		go bt.waitIcon()
		go bt.mTurnOn.Hide()
		go bt.mTurnOff.Hide()
		go bt.mCancel.Hide()

	case "failed":
		systray.SetIcon(icon.Blocked)
		bt.mTurnOn.SetTitle(printer.Sprintf("Retry"))
		go bt.mTurnOn.Show()
		go bt.mTurnOff.Show()
		go bt.mCancel.Hide()
		statusStr = printer.Sprintf("blocking internet")
	}

	systray.SetTooltip(printer.Sprintf("RiseupVPN is %v", statusStr))
	bt.mStatus.SetTitle(printer.Sprintf("VPN is %v", statusStr))
	bt.mStatus.SetTooltip(printer.Sprintf("RiseupVPN is %v", statusStr))
}

func (bt *bmTray) updateDonateMenu() {
	if bt.conf.hasDonated() {
		go bt.mHaveDonated.Hide()
		go bt.mDonate.Hide()
	} else {
		go bt.mHaveDonated.Show()
		go bt.mDonate.Show()
	}
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

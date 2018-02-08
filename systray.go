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
	waitCh        chan bool
	mStatus       *systray.MenuItem
	mTurnOn       *systray.MenuItem
	mTurnOff      *systray.MenuItem
	mCancel       *systray.MenuItem
	activeGateway *gatewayTray
}

type gatewayTray struct {
	menuItem *systray.MenuItem
	name     string
}

func run(bm *bitmask.Bitmask) {
	bt := bmTray{bm: bm}
	systray.Run(bt.onReady, bt.onExit)
}

func (bt bmTray) onExit() {
	// TODO: this doesn't get executed :(
	log.Println("Finished onExit")
}

func (bt *bmTray) onReady() {
	bt.mStatus = systray.AddMenuItem("Checking status...", "")
	bt.mStatus.Disable()
	bt.mTurnOn = systray.AddMenuItem("Turn on", "Turn RiseupVPN on")
	go bt.mTurnOn.Hide()
	bt.mTurnOff = systray.AddMenuItem("Turn off", "Turn RiseupVPN off")
	go bt.mTurnOff.Hide()
	bt.mCancel = systray.AddMenuItem("Cancel", "Cancel connection to RiseupVPN")
	go bt.mCancel.Hide()
	systray.AddSeparator()

	bt.addGateways()

	mHelp := systray.AddMenuItem("Help ...", "")
	mDonate := systray.AddMenuItem("Donate ...", "")
	mAbout := systray.AddMenuItem("About ...", "")
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit BitmaskVPN")

	go func() {
		ch := bt.bm.GetStatusCh()
		status, err := bt.bm.GetStatus()
		if err != nil {
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
				bt.bm.StartVPN(provider)
			case <-bt.mTurnOff.ClickedCh:
				log.Println("off")
				bt.bm.StopVPN()
			case <-bt.mCancel.ClickedCh:
				log.Println("cancel")
				bt.bm.StopVPN()

			case <-mHelp.ClickedCh:
				open.Run("https://riseup.net/en/vpn/vpn-black")
			case <-mDonate.ClickedCh:
				open.Run("https://riseup.net/en/donate")
			case <-mAbout.ClickedCh:
				open.Run("https://bitmask.net")

			case <-mQuit.ClickedCh:
				systray.Quit()
				bt.bm.Close()
				log.Println("Quit now...")
				os.Exit(0)
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

	mGateway := systray.AddMenuItem("Route traffic through", "")
	mGateway.Disable()
	for i, name := range gatewayList {
		menuItem := systray.AddMenuItem(name, "Use RiseupVPN "+name+" gateway")
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

	go bt.gatewaySelection()
}

func (bt *bmTray) gatewaySelection() {

}

func (bt *bmTray) changeStatus(status string) {
	statusStr := status
	bt.mTurnOn.SetTitle("Turn on")
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
		systray.SetIcon(icon.Error)
		bt.mTurnOn.SetTitle("Retry")
		go bt.mTurnOn.Show()
		go bt.mTurnOff.Show()
		go bt.mCancel.Hide()
		statusStr = "blocking internet"
	}

	systray.SetTooltip("RiseupVPN is " + statusStr)
	bt.mStatus.SetTitle("VPN is " + statusStr)
	bt.mStatus.SetTooltip("RiseupVPN is " + statusStr)
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

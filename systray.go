package main

import (
	"fmt"
	"os"

	"0xacab.org/meskio/bitmask-systray/icon"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

type bmTray struct {
	ch        chan string
	mStatus   *systray.MenuItem
	mTurnOn   *systray.MenuItem
	mTurnOff  *systray.MenuItem
	mCancel   *systray.MenuItem
	mGateways []*systray.MenuItem
}

func run(ch chan string) {
	bt := bmTray{ch: ch}
	systray.Run(bt.onReady, bt.onExit)
}

func (bt bmTray) onExit() {
	// TODO: this doesn't get executed :(
	fmt.Println("Finished onExit")
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

	mGateway := systray.AddMenuItem("Route traffic through", "")
	mGateway.Disable()
	bt.mGateways = append(bt.mGateways, systray.AddMenuItem("Seattle", "Use RiseupVPN Seattle gateway"))
	bt.mGateways = append(bt.mGateways, systray.AddMenuItem("Montreal", "Use RiseupVPN Montreal gateway"))
	bt.mGateways = append(bt.mGateways, systray.AddMenuItem("Amsterdam", "Use RiseupVPN Amsterdam gateway"))
	bt.mGateways[0].Check()
	bt.mGateways[1].Uncheck()
	bt.mGateways[2].Uncheck()
	systray.AddSeparator()

	mHelp := systray.AddMenuItem("Help", "")
	mDonate := systray.AddMenuItem("Donate", "")
	mAbout := systray.AddMenuItem("About", "")
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit BitmaskVPN")

	go func() {
		bt.changeStatus(getVPNStatus())

		for {
			select {
			case status := <-ch:
				fmt.Println("status: " + status)
				bt.changeStatus(status)

			case <-bt.mTurnOn.ClickedCh:
				fmt.Println("on")
				startVPN()
			case <-bt.mTurnOff.ClickedCh:
				fmt.Println("off")
				stopVPN()
			case <-bt.mCancel.ClickedCh:
				fmt.Println("cancel")
				cancelVPN()

			case <-mHelp.ClickedCh:
				open.Run("https://riseup.net/en/vpn/vpn-black")
			case <-mDonate.ClickedCh:
				open.Run("https://riseup.net/en/donate")
			case <-mAbout.ClickedCh:
				open.Run("https://bitmask.net")

			case <-mQuit.ClickedCh:
				systray.Quit()
				fmt.Println("Quit now...")
				os.Exit(0)
			}
		}
	}()
}

func (bt *bmTray) changeStatus(status string) {
	statusStr := status

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
		systray.SetIcon(icon.Wait)
		go bt.mTurnOn.Hide()
		go bt.mTurnOff.Hide()
		go bt.mCancel.Show()

	case "stopping":
		systray.SetIcon(icon.Wait)
		go bt.mTurnOn.Hide()
		go bt.mTurnOff.Hide()
		go bt.mCancel.Hide()

	case "failed":
		systray.SetIcon(icon.Error)
		go bt.mTurnOn.Show()
		go bt.mTurnOff.Hide()
		go bt.mCancel.Show()
		statusStr = "blocking internet"
	}

	systray.SetTooltip("RiseupVPN is " + statusStr)
	bt.mStatus.SetTitle("VPN is " + statusStr)
	bt.mStatus.SetTooltip("RiseupVPN is " + statusStr)
}

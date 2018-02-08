package main

import (
	"os"
	"path"
	"time"

	"github.com/0xAX/notificator"
)

const (
	notificationText = `The RiseupVPN service is expensive to run. Because we don't want to store personal information about you, there is no accounts or billing for this service. But if you want the service to continue, donate at least $5 each month at https://riseup.net/donate-vpn`
)

func notificate(conf *systrayConfig) {
	wd, _ := os.Getwd()
	notify := notificator.New(notificator.Options{
		DefaultIcon: path.Join(wd, "riseupvpn.svg"),
		AppName:     "RiseupVPN",
	})

	time.Sleep(time.Minute * 5)
	for {
		if conf.needsNotification() {
			notify.Push("Donate to RiseupVPN", notificationText, "", notificator.UR_NORMAL)
			conf.setNotification()
		}
		time.Sleep(time.Hour)
	}
}

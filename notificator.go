package main

import (
	"os"
	"path"
	"time"

	"github.com/0xAX/notificator"
)

func notificate() {
	// TODO: we need a proper icon
	wd, _ := os.Getwd()
	notify := notificator.New(notificator.Options{
		DefaultIcon: path.Join(wd, "mask.svg"),
		AppName:     "RiseupVPN",
	})

	for {
		notify.Push("Donate", "Have you already donated to RiseupVPN?", "", notificator.UR_NORMAL)
		time.Sleep(time.Minute * 5)
	}
}

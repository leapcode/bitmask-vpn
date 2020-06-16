package backend

import (
	"log"
	"time"
)

func needsDonationReminder() bool {
	return ctx.cfg.NeedsDonationReminder()
}

func donateAccepted() {
	stmut.Lock()
	defer stmut.Unlock()
	ctx.DonateDialog = false
	log.Println("marking as donated")
	ctx.cfg.SetDonated()
	go trigger(OnStatusChanged)
}

func donateRejected() {
	timer := time.NewTimer(time.Hour)
	go func() {
		<-timer.C
		showDonate()
	}()
}

func showDonate() {
	stmut.Lock()
	defer stmut.Unlock()
	ctx.DonateDialog = true
	ctx.cfg.SetLastReminded()
	go trigger(OnStatusChanged)
}

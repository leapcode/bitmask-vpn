package backend

import (
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

// runDonationReminder checks every hour if we need to show the reminder,
// and trigger the launching of the dialog if needed.
func runDonationReminder() {
	go func() {
		for {
			time.Sleep(time.Hour)
			if needsDonationReminder() {
				showDonate()
			}
		}
	}()
}

func wantDonations() bool {
	if config.AskForDonations == "true" {
		return true
	}
	return false
}

func needsDonationReminder() bool {
	return ctx.cfg.NeedsDonationReminder()
}

func donateAccepted() {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	ctx.DonateDialog = false
	ctx.cfg.SetDonated()
	go trigger(OnStatusChanged)
}

func showDonate() {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	ctx.DonateDialog = true
	ctx.cfg.SetLastReminded()
	go trigger(OnStatusChanged)
}

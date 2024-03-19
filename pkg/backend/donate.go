package backend

import (
	"time"
)

// runDonationReminder checks every six hours if we need to show the reminder,
// and trigger the launching of the dialog if needed.
func runDonationReminder() {
	go func() {
		for {
			time.Sleep(time.Hour * 6)
			if needsDonationReminder() {
				showDonate()
			}
		}
	}()
	// to test manually, uncomment this line.
	// time.AfterFunc(1*time.Minute, func() { showDonate() })
}

func needsDonationReminder() bool {
	return ctx.cfg.NeedsDonationReminder()
}

/*
to be called from the gui, the visibility toggle will be updated on the next

	status change
*/
func donateSeen() {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	ctx.DonateDialog = false
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

package backend

import (
	"log"
	"os"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
)

// initializeContext initializes an empty connStatus and assigns it to the
// global ctx holder. This is expected to be called only once, so the public
// api uses the sync.Once primitive to call this.
func initializeContext(opts *InitOpts) {
	var st status = off
	ctx = &connectionCtx{
		AppName:         opts.AppName,
		Provider:        opts.Provider,
		TosURL:          config.TosURL,
		HelpURL:         config.HelpURL,
		DonateURL:       config.DonateURL,
		AskForDonations: wantDonations(),
		DonateDialog:    false,
		Version:         version.VERSION,
		Status:          st,
	}
	errCh := make(chan string)
	go trigger(OnStatusChanged)
	go checkErrors(errCh)
	initializeBitmask(errCh, opts)
}

func checkErrors(errCh chan string) {
	for {
		err := <-errCh
		ctx.Errors = err
		go trigger(OnStatusChanged)
	}
}

func initializeBitmask(errCh chan string, opts *InitOpts) {
	if ctx == nil {
		log.Println("bug: cannot initialize bitmask, ctx is nil!")
		os.Exit(1)
	}
	bitmask.InitializeLogger()

	b, err := bitmask.InitializeBitmask(opts.SkipLaunch)
	if err != nil {
		log.Println("error: cannot initialize bitmask")
		errCh <- err.Error()
		return
	}

	helpers, privilege, err := b.VPNCheck()

	if err != nil {
		log.Println("error doing vpn check")
		errCh <- err.Error()
	}

	if helpers == false {
		log.Println("no helpers")
		errCh <- "nohelpers"
	}
	if privilege == false {
		log.Println("no polkit")
		errCh <- "nopolkit"
	}

	ctx.bm = b
	ctx.cfg = config.ParseConfig()
}

package backend

import (
	"log"
	"os"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
	"0xacab.org/leap/bitmask-vpn/pkg/pid"
)

// initializeContext initializes an empty connStatus and assigns it to the
// global ctx holder. This is expected to be called only once, so the public
// api uses the sync.Once primitive to call this.
func initializeContext(opts *InitOpts) {
	var st status = off

	// TODO - now there's really no need to dance between opts and config anymore
	// but this was the simplest transition. We should probably keep the multi-provider config in the backend too, and just
	// switch the "active" here in the ctx, after the user has selected one in the combobox.
	ctx = &connectionCtx{
		AppName:         opts.ProviderOptions.AppName,
		Provider:        opts.ProviderOptions.Provider,
		TosURL:          opts.ProviderOptions.TosURL,
		HelpURL:         opts.ProviderOptions.HelpURL,
		DonateURL:       opts.ProviderOptions.DonateURL,
		AskForDonations: opts.ProviderOptions.AskForDonations,
		DonateDialog:    false,
		Version:         version.VERSION,
		Status:          st,
		IsReady:         false,
	}
	errCh := make(chan string)
	go checkErrors(errCh)
	// isReady is set after Bitmask initialization
	initializeBitmask(errCh, opts)
	go trigger(OnStatusChanged)
	ctx.delayCheckForGateways()
}

func checkErrors(errCh chan string) {
	for {
		err := <-errCh
		// TODO consider a queue instead
		ctx.Errors = err
		go trigger(OnStatusChanged)
	}
}

func initializeBitmask(errCh chan string, opts *InitOpts) {
	if ctx == nil {
		log.Println("BUG: cannot initialize bitmask, ctx is nil!")
		os.Exit(1)
	}
	bitmask.InitializeLogger()
	ctx.cfg = config.ParseConfig()
	setConfigOpts(opts, ctx.cfg)

	err := pid.AcquirePID()
	if err != nil {
		log.Println("Error acquiring PID:", err)
		log.Fatal(err.Error())
	}

	b, err := bitmask.InitializeBitmask(ctx.cfg)
	if err != nil {
		log.Println("ERROR: cannot initialize bitmask")
		errCh <- err.Error()
		return
	}
	ctx.autostart = initializeAutostart(ctx.cfg)

	helpers, privilege, err := b.VPNCheck()

	if err != nil {
		log.Println("ERROR: vpn check")
		errCh <- err.Error()
	}

	if helpers == false {
		log.Println("ERROR: no helpers")
		errCh <- "nohelpers"
	}
	if privilege == false {
		log.Println("ERROR: no polkit")
		errCh <- "nopolkit"
	}
	ctx.bm = b
	ctx.IsReady = true
}

func setConfigOpts(opts *InitOpts, conf *config.Config) {
	conf.SkipLaunch = opts.SkipLaunch
	if opts.StartVPN != "" {
		if opts.StartVPN != "on" && opts.StartVPN != "off" {
			log.Println("-start-vpn should be 'on' or 'off'")
		} else {
			conf.StartVPN = opts.StartVPN == "on"
		}
	}
	if opts.Obfs4 {
		conf.Obfs4 = opts.Obfs4
	}
	if opts.DisableAutostart {
		conf.DisableAustostart = opts.DisableAutostart
	}
}

func initializeAutostart(conf *config.Config) bitmask.Autostart {
	autostart := bitmask.NewAutostart(config.ApplicationName, "")
	if conf.SkipLaunch || conf.DisableAustostart {
		autostart.Disable()
		autostart = &bitmask.DummyAutostart{}
	}

	err := autostart.Enable()
	if err != nil {
		log.Printf("Error enabling autostart: %v", err)
	}
	return autostart
}

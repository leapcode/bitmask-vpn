package backend

import (
	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	bitmaskAutostart "0xacab.org/leap/bitmask-vpn/pkg/bitmask/autostart"
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
	// but this was the simplest transition. We should probably keep the
	// multi-provider config in the backend too, and just
	// switch the "active" here in the ctx, after the user has selected one
	// in the combobox.
	ctx = &connectionCtx{
		AppName:         opts.ProviderOptions.AppName,
		Provider:        opts.ProviderOptions.Provider,
		TosURL:          opts.ProviderOptions.TosURL,
		HelpURL:         opts.ProviderOptions.HelpURL,
		DonateURL:       opts.ProviderOptions.DonateURL,
		AskForDonations: opts.ProviderOptions.AskForDonations,
		DonateDialog:    false,
		Version:         version.Version(),
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
		log.Fatal().
			Msg("Could not initialize bitmask, ctx is nil!")
	}
	bitmask.InitializeLogger()
	ctx.cfg = config.ParseConfig()
	setConfigOpts(opts, ctx.cfg)
	ctx.UseUDP = ctx.cfg.UDP

	err := pid.AcquirePID()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not acquire PID")
	}

	b, err := bitmask.InitializeBitmask(ctx.cfg)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Could not initialize bitmask")
		errCh <- err.Error()
		return
	}
	// right now we just get autostart from an init flag,
	// but we want to be able to persist that option from the preferences
	// pane
	ctx.autostart = initializeAutostart(ctx.cfg)

	helpers, privilege, err := b.VPNCheck()

	if err != nil {
		log.Error().
			Err(err).
			Msg("Could not check VPN (b.VPNCheck)")
		errCh <- err.Error()
	}

	if !helpers {
		log.Error().Msg("Could not find helpers")
		errCh <- "nohelpers"
	}
	if !privilege {
		log.Error().Msg("Could not find polkit")
		errCh <- "nopolkit"
	}
	ctx.bm = b
	ctx.IsReady = true
}

// transfer initialization options from the config json to the config object
func setConfigOpts(opts *InitOpts, conf *config.Config) {
	conf.SkipLaunch = opts.SkipLaunch
	if opts.StartVPN != "" {
		if opts.StartVPN != "on" && opts.StartVPN != "off" {
			log.Warn().
				Str("startVPN", opts.StartVPN).
				Msg("setConfigOpts: -start-vpn should be 'on' or 'off'")
		} else {
			conf.StartVPN = opts.StartVPN == "on"
		}
	}
	if opts.Obfs4 {
		conf.Obfs4 = opts.Obfs4
	}
	if opts.UDP {
		conf.UDP = opts.UDP
	}
	if opts.DisableAutostart {
		conf.DisableAutostart = opts.DisableAutostart
	}
}

func initializeAutostart(conf *config.Config) bitmaskAutostart.Autostart {
	autostart := bitmaskAutostart.NewAutostart(config.ApplicationName, "")
	if conf.SkipLaunch || conf.DisableAutostart {
		// Disable removes ~.config/autostart/RiseupVPN.desktop: (on Linux)
		// it's possible that the file does not exist, so no need to check err
		_ = autostart.Disable()
		autostart = &bitmaskAutostart.DummyAutostart{}
	} else {
		err := autostart.Enable()
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not enable autostart during initialization")
		}
	}
	return autostart
}

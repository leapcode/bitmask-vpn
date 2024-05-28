package backend

import (
	"github.com/rs/zerolog/log"
)

func startVPN() {
	setError("")
	err := ctx.bm.StartVPN(ctx.Provider)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not start VPN")
		setError(err.Error())
	}
}

func stopVPN() {
	err := ctx.bm.StopVPN()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not stop VPN")
	}
}

func getGateway() string {
	return ctx.bm.GetCurrentGateway()
}

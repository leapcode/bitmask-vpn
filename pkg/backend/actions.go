package backend

import (
	"log"
)

func startVPN() {
	setError("")
	err := ctx.bm.StartVPN(ctx.Provider)
	if err != nil {
		log.Println("ERROR: ", err)
		setError(err.Error())
	}
}

func stopVPN() {
	err := ctx.bm.StopVPN()
	if err != nil {
		log.Println(err)
	}
}

func getGateway() string {
	return ctx.bm.GetCurrentGateway()
}

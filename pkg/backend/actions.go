package backend

import (
	"log"
	"os"
)

func startVPN() {
	err := ctx.bm.StartVPN(ctx.Provider)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func stopVPN() {
	err := ctx.bm.StopVPN()
	if err != nil {
		log.Println(err)
	}
}

func closeVPN() {
	ctx.bm.Close()
}

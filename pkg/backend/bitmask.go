package backend

import (
	"log"
	"os"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
)

func initializeBitmask() {
	if ctx == nil {
		log.Println("error: cannot initialize bitmask, ctx is nil")
		os.Exit(1)
	}
	bitmask.InitializeLogger()

	b, err := bitmask.InitializeBitmask()
	if err != nil {
		log.Println("error: cannot initialize bitmask")
	}
	ctx.bm = b
}

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

// initializeContext initializes an empty connStatus and assigns it to the
// global ctx holder. This is expected to be called only once, so the public
// api uses the sync.Once primitive to call this.
func initializeContext(provider, appName string) {
	var st status = off
	ctx = &connectionCtx{
		AppName:  appName,
		Provider: provider,
		Donate:   false,
		Status:   st,
	}
	go trigger(OnStatusChanged)
	initializeBitmask()
}

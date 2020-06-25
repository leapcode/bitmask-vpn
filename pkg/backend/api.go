/* All the exported functions should be added here */

package backend

import (
	"C"
	"fmt"
	"log"
	"unsafe"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/pickle"
)

func SwitchOn() {
	go setStatus(starting)
	go startVPN()
}

func SwitchOff() {
	go setStatus(stopping)
	go stopVPN()
}

func Unblock() {
	//TODO -
	fmt.Println("unblock... [not implemented]")
}

func Quit() {
	if ctx.Status != off {
		go setStatus(stopping)
		ctx.cfg.SetUserStoppedVPN(false)
	} else {
		ctx.cfg.SetUserStoppedVPN(true)
	}
	ctx.bm.Close()
}

func DonateAccepted() {
	donateAccepted()
}

func SubscribeToEvent(event string, f unsafe.Pointer) {
	subscribe(event, f)
}

func InitializeBitmaskContext() {
	p := bitmask.GetConfiguredProvider()

	initOnce.Do(func() { initializeContext(p.Provider, p.AppName) })
	runDonationReminder()
	go ctx.updateStatus()
}

func RefreshContext() *C.char {
	c, _ := ctx.toJson()
	return C.CString(string(c))
}

func InstallHelpers() {
	pickle.InstallHelpers()
}

func MockUIInteraction() {
	log.Println("mocking ui interaction on port 8080. \nTry 'curl localhost:8080/{on|off|failed}' to toggle status.")
	go mockUI()
}

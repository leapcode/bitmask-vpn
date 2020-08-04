/* All the exported functions should be added here */

package backend

import (
	"C"
	"fmt"
	"log"
	"strconv"
	"unsafe"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
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

type InitOpts struct {
	Provider   string
	AppName    string
	SkipLaunch bool
}

func InitializeBitmaskContext(opts *InitOpts) {
	p := bitmask.GetConfiguredProvider()
	opts.Provider = p.Provider
	opts.AppName = p.AppName

	initOnce.Do(func() { initializeContext(opts) })
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

func EnableMockBackend() {
	log.Println("[+] Mocking ui interaction on port 8080. \nTry 'curl localhost:8080/{on|off|failed}' to toggle status.")
	go enableMockBackend()
}

func EnableWebAPI(port string) {
	intPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal("Cannot parse port", port)
	}
	go enableWebAPI(intPort)
}

/* these two are a bit redundant since we already add them to ctx. however, we
   want to have them available before everything else, to be able to parse cli
   arguments. In the long run, we probably want to move all vendoring to qt, so
   this probably should not live in the backend. */

func GetVersion() *C.char {
	return C.CString(version.VERSION)
}

func GetAppName() *C.char {
	p := bitmask.GetConfiguredProvider()
	return C.CString(p.AppName)
}

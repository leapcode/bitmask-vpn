/* All the exported functions live here */

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
	fmt.Println("unblock... [not implemented]")
}

func Quit() {
	if ctx.Status != off {
		go setStatus(stopping)
		stopVPN()
	}
}

func ToggleDonate() {
	toggleDonate()
}

func SubscribeToEvent(event string, f unsafe.Pointer) {
	subscribe(event, f)
}

func InitializeBitmaskContext() {
	pi := bitmask.GetConfiguredProvider()

	initOnce.Do(func() {
		initializeContext(pi.Provider, pi.AppName)
	})
	go ctx.updateStatus()

	/* DEBUG
	timer := time.NewTimer(time.Second * 3)
	go func() {
		<-timer.C
		fmt.Println("donate timer fired")
		toggleDonate()
	}()
	*/
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

package main

/* a wrapper around bitmask that exposes status to a QtQml gui.
   Have a look at the pkg/backend module for further enlightment. */

import (
	"C"
	"unsafe"

	"0xacab.org/leap/bitmask-vpn/pkg/backend"
)

//export SwitchOn
func SwitchOn() {
	backend.SwitchOn()
}

//export SwitchOff
func SwitchOff() {
	backend.SwitchOff()
}

//export Unblock
func Unblock() {
	backend.Unblock()
}

//export Quit
func Quit() {
	backend.Quit()

}

//export DonateAccepted
func DonateAccepted() {
	backend.DonateAccepted()
}

//export DonateRejected
func DonateRejected() {
	backend.DonateRejected()
}

//export SubscribeToEvent
func SubscribeToEvent(event string, f unsafe.Pointer) {
	backend.SubscribeToEvent(event, f)
}

//export InitializeBitmaskContext
func InitializeBitmaskContext() {
	backend.InitializeBitmaskContext()
}

//export RefreshContext
func RefreshContext() *C.char {
	return (*C.char)(backend.RefreshContext())
}

//export InstallHelpers
func InstallHelpers() {
	backend.InstallHelpers()
}

func main() {}

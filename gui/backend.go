package main

/* a wrapper around bitmask that exposes status to a QtQml gui.
   Have a look at the pkg/backend module for further enlightment. */

import (
	"C"
	"unsafe"

	"0xacab.org/leap/bitmask-vpn/pkg/backend"
)

//export GetBitmaskVersion
func GetBitmaskVersion() *C.char {
	return (*C.char)(backend.GetVersion())
}

//export Login
func Login(username, password string) {
	backend.Login(username, password)
}

//export SwitchOn
func SwitchOn() {
	backend.SwitchOn()
}

//export SwitchOff
func SwitchOff() {
	backend.SwitchOff()
}

//export UseGateway
func UseGateway(label string) {
	backend.UseGateway(label)
}

//export UseTransport
func UseTransport(transport string) {
	backend.UseTransport(transport)
}

//export Quit
func Quit() {
	backend.Quit()
}

//export DonateAccepted
func DonateAccepted() {
	backend.DonateAccepted()
}

//export DonateSeen
func DonateSeen() {
	backend.DonateSeen()
}

//export SubscribeToEvent
func SubscribeToEvent(event string, f unsafe.Pointer) {
	backend.SubscribeToEvent(event, f)
}

//export InitializeBitmaskContext
func InitializeBitmaskContext(provider string,
	jsonPtr unsafe.Pointer, jsonLen C.int,
	obfs4 bool, disableAutostart bool, startVPN string) {

	json := C.GoBytes(jsonPtr, jsonLen)
	opts := backend.InitOptsFromJSON(provider, string(json))
	opts.Obfs4 = obfs4
	opts.DisableAutostart = disableAutostart
	opts.StartVPN = startVPN
	backend.InitializeBitmaskContext(opts)
}

//export InitializeTestBitmaskContext
func InitializeTestBitmaskContext(provider string,
	jsonPtr unsafe.Pointer, jsonLen C.int) {
	json := C.GoBytes(jsonPtr, jsonLen)
	opts := backend.InitOptsFromJSON(provider, string(json))
	opts.DisableAutostart = true
	opts.SkipLaunch = true
	opts.StartVPN = "no"
	backend.InitializeBitmaskContext(opts)
	backend.EnableMockBackend()
}

//export EnableWebAPI
func EnableWebAPI(port string) {
	backend.EnableWebAPI(port)
}

//export RefreshContext
func RefreshContext() *C.char {
	return (*C.char)(backend.RefreshContext())
}

//export ResetError
func ResetError(errname string) {
	backend.ResetError(errname)
}

//export ResetNotification
func ResetNotification(label string) {
	backend.ResetNotification(label)
}

//export InstallHelpers
func InstallHelpers() {
	backend.InstallHelpers()
}

func main() {}

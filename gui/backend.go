package main

/* a wrapper around bitmask that exposes status to a QtQml gui.
   Have a look at the pkg/backend module for further enlightment. */

import "C"

import "unsafe"

import (
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

//export UseLocation
func UseLocation(label *C.char) {
	location := C.GoString(label)
	backend.UseLocation(location)
}

//export UseAutomaticGateway
func UseAutomaticGateway() {
	backend.UseAutomaticGateway()
}

//export SetTransport
func SetTransport(transport *C.char) {
	tp := C.GoString(transport)
	backend.SetTransport(tp)
}

//export GetTransport
func GetTransport() *C.char {
	return (*C.char)(backend.GetTransport())
}

//export SetUDP
func SetUDP(udp bool) {
	backend.SetUDP(udp)
}

//export SetSnowflake
func SetSnowflake(snowflake bool) {
	backend.SetSnowflake(snowflake)
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
func InitializeBitmaskContext(provider *C.char,
	jsonPtr unsafe.Pointer, jsonLen C.int,
	obfs4 bool, disableAutostart bool, startVPN string) {
	json := C.GoBytes(jsonPtr, jsonLen)
	providerName := C.GoString(provider)
	opts := backend.InitOptsFromJSON(providerName, string(json))
	opts.Obfs4 = obfs4
	opts.DisableAutostart = disableAutostart
	opts.StartVPN = startVPN
	opts.DisableAutostart = true
	opts.SkipLaunch = true
	go backend.InitializeBitmaskContext(opts)
}

//export SwitchProvider
func SwitchProvider(provider *C.char) {
	// provider could be provider URL or provider name
	providerNameOrURL := C.GoString(provider)
	var opts = &backend.InitOpts{}
	var providerName string

	/* TODO: read the following values from the on-disk config file */
	opts.Obfs4 = false
	opts.DisableAutostart = true
	opts.SkipLaunch = true

	if backend.IsProviderURI(providerNameOrURL) {
		providerName = backend.FetchProviderOptsFromRemote(providerNameOrURL)
	}

	if len(providerName) == 0 {
		providerName = providerNameOrURL
	}

	opts = backend.InitOptsFromJSON(providerName, "")
	go backend.InitializeBitmaskContext(opts)
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

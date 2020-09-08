/* All the exported functions should be added here */

package backend

import (
	"C"
	"encoding/json"
	"log"
	"strconv"
	"unsafe"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
	"0xacab.org/leap/bitmask-vpn/pkg/pickle"
)

func Login(username, password string) {
	success, err := ctx.bm.DoLogin(username, password)
	if err != nil {
		log.Printf("Error on login: %v", err)
		ctx.Errors = "bad_auth"
	} else if success {
		log.Printf("Logged in as %s", username)
		ctx.LoginOk = true
		ctx.LoginDialog = false
	} else {
		// TODO: display login again with an err
		log.Printf("Failed to login as %s", username)
		ctx.LoginDialog = true
		ctx.Errors = "bad_auth"
	}
	go ctx.updateStatus()
}

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
	if ctx.bm != nil {
		ctx.bm.Close()
	}
}

func DonateAccepted() {
	donateAccepted()
}

func SubscribeToEvent(event string, f unsafe.Pointer) {
	subscribe(event, f)
}

type Providers struct {
	Default string                 `json:"default"`
	Data    []bitmask.ProviderOpts `json:"providers"`
}

type InitOpts struct {
	ProviderOptions *bitmask.ProviderOpts
	SkipLaunch      bool
}

func InitOptsFromJSON(provider, providersJSON string) *InitOpts {
	providers := Providers{}
	err := json.Unmarshal([]byte(providersJSON), &providers)
	if err != nil {
		log.Println("ERROR while parsing json:", err)
	}
	if len(providers.Data) != 1 {
		panic("BUG: we do not support multi-provider yet")
	}
	providerOpts := &providers.Data[0]
	return &InitOpts{providerOpts, false}
}

func InitializeBitmaskContext(opts *InitOpts) {
	bitmask.ConfigureProvider(opts.ProviderOptions)

	initOnce.Do(func() { initializeContext(opts) })
	if ctx.bm != nil {
		ctx.LoginDialog = ctx.bm.NeedsCredentials()
		go ctx.updateStatus()
	}
	runDonationReminder()
}

func RefreshContext() *C.char {
	c, _ := ctx.toJson()
	return C.CString(string(c))
}

func ResetError(errname string) {
	if ctx.Errors == errname {
		ctx.Errors = ""
	}
}

func ResetNotification(label string) {
	switch label {
	case "login_ok":
		ctx.LoginOk = false
		break
	default:
		break
	}
	go trigger(OnStatusChanged)
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
   this probably should not live in the backend, see #326*/

func GetVersion() *C.char {
	return C.CString(version.VERSION)
}

func GetAppName() *C.char {
	p := bitmask.GetConfiguredProvider()
	return C.CString(p.AppName)
}

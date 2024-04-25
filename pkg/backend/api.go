/* All the exported functions should be added here */

package backend

import (
	"C"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
	"0xacab.org/leap/bitmask-vpn/pkg/pickle"
	"0xacab.org/leap/bitmask-vpn/pkg/pid"
	"github.com/rs/zerolog/log"
)

func Login(username, password string) {
	success, err := ctx.bm.DoLogin(username, password)
	if err != nil {
		log.Warn().
			Err(err).
			Str("username", username).
			Msg("Could not log in")
		if err.Error() == "TokenErrTimeout" {
			ctx.Errors = "bad_auth_timeout"
		} else if err.Error() == "TokenErrBadStatus 502" {
			ctx.Errors = "bad_auth_502"
		} else {
			ctx.Errors = "bad_auth"
		}
	} else if success {
		log.Info().
			Str("username", username).
			Msg("Sucessfully logged in")
		ctx.LoginOk = true
		ctx.LoginDialog = false
	} else {
		log.Warn().
			Str("username", username).
			Msg("Could not log in (else)")
		ctx.LoginDialog = true
		ctx.Errors = "bad_auth"
	}
	// XXX shouldn't this be statusChanged?
	go ctx.updateStatus()
}

func setError(err string) {
	ctx.Errors = err
	go setStatus(off)
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

func UseLocation(label string) {
	if ctx.ManualLocation && label == ctx.CurrentLocation {
		return
	}

	ctx.bm.UseGateway(label)
	go trigger(OnStatusChanged)
	if ctx.Status == on && label != strings.ToLower(ctx.CurrentLocation) {
		go ctx.bm.Reconnect()
	}
}

func UseAutomaticGateway() {
	if !ctx.ManualLocation {
		return
	}

	ctx.bm.UseAutomaticGateway()
	go trigger(OnStatusChanged)
	if ctx.Status == on {
		ctx.bm.Reconnect()
	}
}

func SetTransport(label string) {
	err := ctx.bm.SetTransport(label)
	if err != nil {
		log.Warn().
			Err(err).
			Str("transport", label).
			Msg("Could not set transport")
	}
	if label == "obfs4" {
		// XXX this is an expedite way of avoiding the corner case
		// in which user has selected a manual location that does not offer bridges.
		// In the future, we can be more delicate and 1. do the switch only if the manual location
		// is incompatible with obfs4; 2. notify the user of the change.
		// But tonight we're in problem-solving mode, and we can assume that user wants to use bridges,
		// no matter what. So let's assume that "use obfs4" supersedes everything else and be done.
		UseAutomaticGateway()
		ctx.cfg.SetUseObfs4(true)
	} else {
		ctx.cfg.SetUseObfs4(false)
	}
	go trigger(OnStatusChanged)
}

func SetUDP(udp bool) {
	log.Info().
		Bool("useUDP", udp).
		Msg("Configuring UDP")
	ctx.cfg.SetUseUDP(udp)
	ctx.bm.UseUDP(udp)
	go trigger(OnStatusChanged)
}

func SetSnowflake(snowflake bool) {
	log.Info().
		Bool("useSnowflake", snowflake).
		Msg("Configuring Snowflake")
	ctx.cfg.SetUseSnowflake(snowflake)
	ctx.bm.UseSnowflake(snowflake)
	go trigger(OnStatusChanged)
}

func GetTransport() *C.char {
	return C.CString(ctx.bm.GetTransport())
}

func Quit() {
	if ctx.autostart != nil {
		err := ctx.autostart.Disable()
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not disable autostart")
		}
	}
	if ctx.Status != off {
		go setStatus(stopping)
		err := ctx.cfg.SetUserStoppedVPN(false)
		if err != nil {
			log.Warn().
				Err(err).
				Bool("userStopped", false).
				Msg("Could not set UserStoppedVPN")
		}
	} else {
		err := ctx.cfg.SetUserStoppedVPN(true)
		if err != nil {
			log.Warn().
				Err(err).
				Bool("userStopped", false).
				Msg("Could not set UserStoppedVPN")
		}
	}
	if ctx.bm != nil {
		ctx.bm.Close()
	}
	err := pid.ReleasePID()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not release PID")
	}
}

func DonateAccepted() {
	donateAccepted()
}

func DonateSeen() {
	donateSeen()
}

func SubscribeToEvent(event string, f unsafe.Pointer) {
	subscribe(event, f)
}

type Providers struct {
	Default string                 `json:"default"`
	Data    []bitmask.ProviderOpts `json:"providers"`
}

type InitOpts struct {
	ProviderOptions  *bitmask.ProviderOpts
	SkipLaunch       bool
	Obfs4            bool
	UDP              bool
	DisableAutostart bool
	StartVPN         string
}

func InitOptsFromJSON(provider, providersJSON string) *InitOpts {
	providers := Providers{}
	err := json.Unmarshal([]byte(providersJSON), &providers)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("providersJson", providersJSON).
			Msg("Could not parse provider json")
	}
	var providerOpts *bitmask.ProviderOpts
	providerOpts = &providers.Data[0]
	if len(providers.Data) != 1 {
		chosenProvider := os.Getenv("LEAP_PROVIDER")
		if chosenProvider != "" {
			for _, p := range providers.Data {
				if p.Provider == chosenProvider {
					log.Info().
						Str("provider", chosenProvider).
						Msg("Selecting provider")
					return &InitOpts{ProviderOptions: &p}
				}
			}
			panic("BUG: unknown provider")
		}
	}
	return &InitOpts{ProviderOptions: providerOpts}
}

func InitializeBitmaskContext(opts *InitOpts) {
	bitmask.ConfigureProvider(opts.ProviderOptions)

	initOnce.Do(func() { initializeContext(opts) })
	if ctx.bm != nil {
		ctx.LoginDialog = ctx.bm.NeedsCredentials()
		go ctx.updateStatus()
	}
	if ctx.AskForDonations {
		runDonationReminder()
	}
}

func RefreshContext() *C.char {
	c, err := ctx.toJson()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not refresh context")
	}
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
	log.Info().Msg("[+] Mocking ui interaction on port 8080. \nTry 'curl localhost:8080/{on|off|failed}' to toggle status.")
	go enableMockBackend()
}

func EnableWebAPI(port string) {
	intPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("port", port).
			Msg("Could not parse WebAPI port")
	}
	go enableWebAPI(intPort)
}

func GetVersion() *C.char {
	return C.CString(version.VERSION)
}

// go:build cgo
// +go:build cgo

/* All the exported functions should be added here */

package backend

import (
	"C"
	"encoding/json"
	"net/url"
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

var providers *Providers

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
	switch label {
	case "obfs4":
		// XXX this is an expedite way of avoiding the corner case
		// in which user has selected a manual location that does not offer bridges.
		// In the future, we can be more delicate and 1. do the switch only if the manual location
		// is incompatible with obfs4; 2. notify the user of the change.
		// But tonight we're in problem-solving mode, and we can assume that user wants to use bridges,
		// no matter what. So let's assume that "use obfs4" supersedes everything else and be done.
		UseAutomaticGateway()
		ctx.cfg.SetUseObfs4(true)
		ctx.cfg.SetUseKCP(false)
	case "kcp":
		ctx.cfg.SetUseObfs4(true)
		ctx.cfg.SetUseKCP(true)
	default:
		ctx.cfg.SetUseObfs4(false)
		ctx.cfg.SetUseKCP(false)
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
	ProviderOptions    *bitmask.ProviderOpts
	SkipLaunch         bool
	Obfs4              bool
	UDP                bool
	DisableAutostart   bool
	StartVPN           string
	AvailableProviders []string
}

// InitOptsFromJSON initializes the provider configuration (InitOpts) struct. It is
// called by the c++ code during startup. There is hardcoded provider information from
// the binary (providersJSON). And there are dynamic provider configurations saved on
// disk (loaded by appendOnDiskProviders function).
func InitOptsFromJSON(providerName, providersJSON string) *InitOpts {
	log.Debug().Str("providerName", providerName).Msg("initializing for provider")

	// load all providers into providers (the one from the binary + all providers stored on disk)
	if providers == nil {
		providers = &Providers{}
		err := json.Unmarshal([]byte(providersJSON), providers)
		if err != nil {
			log.Fatal().
				Err(err).
				Str("providersJson", providersJSON).
				Msg("Could not parse provider json")
		}
		providers = appendOnDiskProviders(providers)
	}
	initOpts := &InitOpts{}

	if enforcedProviderEnv := os.Getenv("LEAP_PROVIDER"); enforcedProviderEnv != "" {
		providerName = enforcedProviderEnv
	}

	for _, p := range providers.Data {
		initOpts.AvailableProviders = append(initOpts.AvailableProviders, p.Provider)
	}

	// we do the following check as a protection for release builds providers.Data will always
	// be > 0
	if len(providers.Data) > 0 {
		for _, p := range providers.Data {
			if p.Provider == providerName {
				log.Info().
					Str("providerName", providerName).
					Msg("Selecting provider")
				initOpts.ProviderOptions = &p
				return initOpts
			}
		}
		log.Fatal().
			Str("providerName", providerName).
			Msg("Provider not found in providers.json")
	}
	return initOpts
}

func InitializeBitmaskContext(opts *InitOpts) {
	log.Info().Msg("Initializing bitmask context")
	bitmask.ConfigureProvider(opts.ProviderOptions)

	initializeContext(opts)
	if ctx.bm != nil {
		log.Info().Msg("Updating status in context")
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
	return C.CString(version.Version())
}

func IsProviderURI(provider string) bool {
	if _, err := url.ParseRequestURI(provider); err != nil {
		return false
	}
	return true
}

func FetchProviderOptsFromRemote(providerURL string) string {
	url, err := url.ParseRequestURI(providerURL)
	if err != nil {
		log.Debug().
			Err(err).
			Msg("Failed to parse provider URL")
		return ""
	}

	opts := &bitmask.ProviderOpts{}
	switch url.Scheme {
	case "obfsvpnintro":
		opts = fetchProviderOptsWithIntroducer(providerURL)
	case "https", "http":
		opts = fetchProviderOptsWitBootstrapper(providerURL)
	}

	if len(opts.Provider) > 0 {
		if !providerAlreadyExists(providers, opts) {
			log.Debug().
				Msg("Adding newly fetched provider to global providers var")
			if err := writeProviderOptsToFile(opts); err != nil {
				log.Debug().
					Err(err).
					Msg("Failed to write provider options to file")
			}
			providers.Data = append(providers.Data, *opts)
		}
		return opts.Provider
	}
	return ""
}

func writeProviderOptsToFile(opts *bitmask.ProviderOpts) error {
	return writeProviderJSONToFile(opts, getProviderJSONPath(opts.Provider))
}

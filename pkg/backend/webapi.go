package backend

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"net/http"
	"os"
	"strconv"
	"time"
)

func CheckAuth(handler http.HandlerFunc, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("X-Auth-Token")
		if t == token {
			handler(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 - Unauthorized"))
		}
	}
}

func webOn(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Web UI: on")
	SwitchOn()
}

func webOff(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Web UI: off")
	SwitchOff()
}

func webStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, ctx.Status.String())
}

func webGatewayGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, ctx.bm.GetCurrentGateway())
}

func webGatewaySet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		gwLabel := r.FormValue("transport")
		fmt.Fprintf(w, "selected gateway: %s\n", gwLabel)
		ctx.bm.UseGateway(gwLabel)
		// TODO make sure we don't tear the fw down on reconnect...
		SwitchOff()
		// a little sleep is needed, though, because iptables takes some time
		time.Sleep(500 * time.Millisecond)
		SwitchOn()
	default:
		fmt.Fprintf(w, "Only POST supported.")
	}
}

func webGatewayList(w http.ResponseWriter, r *http.Request) {
	transport := ctx.bm.GetTransport()
	locationJson, err := json.Marshal(ctx.bm.ListLocationFullness(transport))
	if err != nil {
		fmt.Fprintf(w, "Error converting json: %v", err)
	}
	fmt.Fprintf(w, string(locationJson))
}

func webTransportGet(w http.ResponseWriter, r *http.Request) {
	t, err := json.Marshal(ctx.bm.GetTransport())
	if err != nil {
		fmt.Fprintf(w, "Error converting json: %v", err)
	}
	fmt.Fprintf(w, string(t))

}

func webTransportSet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		t := r.FormValue("transport")
		if isValidTransport(t) {
			fmt.Fprintf(w, "Selected transport: %s\n", t)
			go ctx.bm.SetTransport(string(t))
		} else {
			fmt.Fprintf(w, "Unknown transport: %s\n", t)
		}
	default:
		fmt.Fprintf(w, "Only POST supported.")
	}
}

func webTransportList(w http.ResponseWriter, r *http.Request) {
	t, err := json.Marshal([]string{"openvpn", "obfs4"})
	if err != nil {
		fmt.Fprintf(w, "Error converting json: %v", err)
	}
	fmt.Fprintf(w, string(t))
}

func webQuit(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Web UI: quit")
	Quit()
	os.Exit(0)
}

func enableWebAPI(port int) {
	log.Debug().
		Int("port", port).
		Msg("Starting WebAPI")
	generateAuthToken()
	token := readAuthToken()
	http.Handle("/vpn/start", CheckAuth(http.HandlerFunc(webOn), token))
	http.Handle("/vpn/stop", CheckAuth(http.HandlerFunc(webOff), token))
	http.Handle("/vpn/gw/get", CheckAuth(http.HandlerFunc(webGatewayGet), token))
	http.Handle("/vpn/gw/set", CheckAuth(http.HandlerFunc(webGatewaySet), token))
	http.Handle("/vpn/gw/list", CheckAuth(http.HandlerFunc(webGatewayList), token))
	http.Handle("/vpn/transport/get", CheckAuth(http.HandlerFunc(webTransportGet), token))
	http.Handle("/vpn/transport/set", CheckAuth(http.HandlerFunc(webTransportSet), token))
	http.Handle("/vpn/transport/list", CheckAuth(http.HandlerFunc(webTransportList), token))
	http.Handle("/vpn/status", CheckAuth(http.HandlerFunc(webStatus), token))
	http.Handle("/vpn/quit", CheckAuth(http.HandlerFunc(webQuit), token))
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Warn().
			Err(err).
			Str("bindAddr", ":"+strconv.Itoa(port)).
			Msg("Could not listen on port for WebAPI")

	}
}

func isValidTransport(t string) bool {
	for _, b := range []string{"openvpn", "obfs4"} {
		if b == t {
			return true
		}
	}
	return false
}

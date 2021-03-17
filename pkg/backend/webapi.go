package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
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
	log.Println("Web UI: on")
	SwitchOn()
}

func webOff(w http.ResponseWriter, r *http.Request) {
	log.Println("Web UI: off")
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
		gwLabel := r.FormValue("gw")
		fmt.Fprintf(w, "selected gateway: %s\n", gwLabel)
		// FIXME catch error here, return it (error code)
		useGateway(gwLabel)
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
	gws, err := ctx.bm.ListGatewaysByCity(ctx.Provider)
	if err != nil {
		fmt.Fprintf(w, "ListGatewaysByCity() err: %v", err)
	}
	gwJson, _ := json.Marshal(gws)
	fmt.Fprintf(w, string(gwJson))
}

// TODO
func webTransportGet(w http.ResponseWriter, r *http.Request) {
}

// TODO
func webTransportSet(w http.ResponseWriter, r *http.Request) {
}

// TODO
func webTransportList(w http.ResponseWriter, r *http.Request) {
}

func webQuit(w http.ResponseWriter, r *http.Request) {
	log.Println("Web UI: quit")
	Quit()
	os.Exit(0)
}

func enableWebAPI(port int) {
	log.Println("Starting WebAPI in port", port)
	bitmask.GenerateAuthToken()
	token := bitmask.ReadAuthToken()
	http.Handle("/vpn/start", CheckAuth(http.HandlerFunc(webOn), token))
	http.Handle("/vpn/stop", CheckAuth(http.HandlerFunc(webOff), token))
	http.Handle("/vpn/gw/get", CheckAuth(http.HandlerFunc(webGatewayGet), token))
	http.Handle("/vpn/gw/set", CheckAuth(http.HandlerFunc(webGatewaySet), token))
	http.Handle("/vpn/gw/list", CheckAuth(http.HandlerFunc(webGatewayList), token))
	//http.Handle("/vpn/transport/get", CheckAuth(http.HandlerFunc(webTransportGet), token))
	//http.Handle("/vpn/transport/set", CheckAuth(http.HandlerFunc(webTransportSet), token))
	//http.Handle("/vpn/transport/list", CheckAuth(http.HandlerFunc(webTransportList), token))
	http.Handle("/vpn/status", CheckAuth(http.HandlerFunc(webStatus), token))
	http.Handle("/vpn/quit", CheckAuth(http.HandlerFunc(webQuit), token))
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

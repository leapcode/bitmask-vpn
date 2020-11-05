package backend

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
	http.Handle("/vpn/status", CheckAuth(http.HandlerFunc(webStatus), token))
	http.Handle("/vpn/quit", CheckAuth(http.HandlerFunc(webQuit), token))
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

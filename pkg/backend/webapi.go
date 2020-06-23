package backend

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
)

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

type Adapter func(http.Handler) http.Handler

func CheckAuth(token string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t := r.Header.Get("X-Auth-Token")
			if t == token {
				h.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("401 - Unauthorized"))
			}

		})
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
	log.Println("Web UI: status")
	fmt.Fprintf(w, ctx.Status.String())
}

func webQuit(w http.ResponseWriter, r *http.Request) {
	log.Println("Web UI: quit")
	Quit()
	os.Exit(0)
}

func enableWebAPI() {
	bitmask.GenerateAuthToken()
	auth := CheckAuth(bitmask.ReadAuthToken())
	http.Handle("/vpn/start", Adapt(http.HandlerFunc(webOn), auth))
	http.Handle("/vpn/stop", Adapt(http.HandlerFunc(webOff), auth))
	http.Handle("/vpn/status", Adapt(http.HandlerFunc(webStatus), auth))
	http.Handle("/vpn/quit", Adapt(http.HandlerFunc(webQuit), auth))
	http.ListenAndServe(":8080", nil)
}

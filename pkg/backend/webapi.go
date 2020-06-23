package backend

import (
	"fmt"
	"log"
	"net/http"
)

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
}

func enableWebAPI() {
	http.HandleFunc("/vpn/start", webOn)
	http.HandleFunc("/vpn/stop", webOff)
	http.HandleFunc("/vpn/status", webStatus)
	http.HandleFunc("/vpn/quit", webQuit)
	http.ListenAndServe(":8080", nil)
}

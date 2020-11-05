package backend

import (
	"log"
	"net/http"
)

/* mock http server: easy way to mocking vpn behavior on ui interaction. This
* should also show a good way of writing functionality tests just for the Qml
* layer */

func enableMockBackend() {
	log.Println("[+] You should not use this in production!")
	http.HandleFunc("/on", mockUIOn)
	http.HandleFunc("/off", mockUIOff)
	http.HandleFunc("/failed", mockUIFailed)
	http.ListenAndServe(":8080", nil)
}

func mockUIOn(w http.ResponseWriter, r *http.Request) {
	log.Println("changing status: on")
	setStatus(on)
}

func mockUIOff(w http.ResponseWriter, r *http.Request) {
	log.Println("changing status: off")
	setStatus(off)
}

func mockUIFailed(w http.ResponseWriter, r *http.Request) {
	log.Println("changing status: failed")
	setStatus(failed)
}

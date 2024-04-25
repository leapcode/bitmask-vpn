package backend

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

/* mock http server: easy way to mocking vpn behavior on ui interaction. This
* should also show a good way of writing functionality tests just for the Qml
* layer */

func enableMockBackend() {
	log.Warn().Msg("[+] You should not use this in production!")
	http.HandleFunc("/on", mockUIOn)
	http.HandleFunc("/off", mockUIOff)
	http.HandleFunc("/failed", mockUIFailed)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not run mock backend")
	}
}

func mockUIOn(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("changing status: on")
	setStatus(on)
}

func mockUIOff(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("changing status: off")
	setStatus(off)
}

func mockUIFailed(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("changing status: failed")
	setStatus(failed)
}

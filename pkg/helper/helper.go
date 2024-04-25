// Copyright (C) 2018 LEAP
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// This helper is intended to be long-lived, and run with administrator privileges.
// It will launch a http server and expose a REST API to control OpenVPN and the firewall.
// At the moment, it is only used in Darwin and Windows - although it could also be used in GNU/Linux systems (but we use the one-shot bitmask-root wrapper in GNU/Linux instead).
// In Windows, this helper will run on the first available port after the standard one (7171).
// In other systems, the 7171 port is hardcoded.

package helper

import (
	"encoding/json"
	"net"
	"net/http"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	AppName    = "DemoLibVPN"
	BinaryName = "bitmask"
	Version    = "git"
)

type openvpnT struct {
	cmd *exec.Cmd
}

// startHelper is the main entrypoint. It can react to cli args (used to install or manage the service in windows), and
// eventually will start the http server.
func StartHelper(port int) {
	initializeService(port)
	parseCliArgs()
	daemonize()
	runServer(port)
}

// serveHTTP will start the HTTP server that exposes the firewall and openvpn api.
// this can be called at different times by the different implementations of the helper.
func serveHTTP(bindAddr string) {
	log.Info().
		Str("bindAddr", bindAddr).
		Msg("Starting HTTP server")
	openvpn := openvpnT{nil}
	http.HandleFunc("/openvpn/start", openvpn.start)
	http.HandleFunc("/openvpn/stop", openvpn.stop)
	http.HandleFunc("/firewall/start", firewallStartHandler)
	http.HandleFunc("/firewall/stop", firewallStopHandler)
	http.HandleFunc("/firewall/isup", firewallIsUpHandler)
	http.HandleFunc("/version", versionHandler)

	err := http.ListenAndServe(bindAddr, nil)
	log.Fatal().
		Err(err).
		Msg("Could not start HTTP Server")
}

func (openvpn *openvpnT) start(w http.ResponseWriter, r *http.Request) {
	args, err := getArgs(r)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not process OpenVPN arguments")
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Warn().
					Err(err).
					Msg("Could not write http reponse")
			}
		}
		return
	}

	args = parseOpenvpnArgs(args)
	log.Info().
		Str("args", strings.Join(args, " ")).
		Msg("Starting OpenVPN")
	err = openvpn.run(args)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not start OpenVPN")
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not write http reponse")
		}
	}
}

func (openvpn *openvpnT) run(args []string) error {
	if openvpn.cmd != nil {
		log.Info().Msg("OpenVPN is running, stopping it")
		err := openvpn.kill()
		if err != nil {
			return err
		}
	}
	log.Debug().
		Str("path", getOpenvpnPath()).
		Msg("OpenVPN path")

	// TODO: if it dies we should restart it
	openvpn.cmd = exec.Command(getOpenvpnPath(), args...)
	return openvpn.cmd.Start()
}

func (openvpn *openvpnT) stop(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Stopping OpenVPN")
	if openvpn.cmd == nil || openvpn.cmd.ProcessState != nil {
		openvpn.cmd = nil
		return
	}

	err := openvpn.kill()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not stop OpenVPN process")
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not write http reponse")
		}
	}
}

func (openvpn *openvpnT) kill() error {
	err := kill(openvpn.cmd)
	if err == nil {
		err = openvpn.cmd.Wait()
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not wait for process")
		}
	} else {
		log.Warn().
			Err(err).
			Msg("Could not kill process")
	}

	openvpn.cmd = nil
	return nil
}

func firewallStartHandler(w http.ResponseWriter, r *http.Request) {
	mode := "tcp"
	query := r.URL.Query()
	udp, udpParam := query["udp"]
	if udpParam && len(udp) == 1 && udp[0] == "1" {
		mode = "udp"
	}
	gateways, err := getArgs(r)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not process gateways")
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not write http reponse")
		}
		return
	}

	for _, gw := range gateways {
		if !validAddress(gw) {
			_, err = w.Write([]byte("bad argument"))
			if err != nil {
				log.Warn().
					Err(err).
					Msg("Could not write http reponse")
			}
		}
	}

	err = firewallStart(gateways, mode)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not start firewall")
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not write http reponse")
		}
		return
	}
	log.Info().Msg("Successfully started firewall")
}

func firewallStopHandler(w http.ResponseWriter, r *http.Request) {
	err := firewallStop()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not stop firewall")
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not write http reponse")
		}
	}
	log.Info().Msg("Successfully stopped firewalll")
}

func firewallIsUpHandler(w http.ResponseWriter, r *http.Request) {
	if firewallIsUp() {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("true"))
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not write http reponse")
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
		_, err := w.Write([]byte("false"))
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not write http reponse")
		}
	}
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(AppName + "/" + Version + "\n"))
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not write http reponse")
	}
}

func getArgs(r *http.Request) ([]string, error) {
	args := []string{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&args)
	return args, err
}

func validAddress(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	} else {
		return true
	}
}

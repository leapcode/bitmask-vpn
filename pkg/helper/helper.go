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
	"log"
	"net"
	"net/http"
	"os/exec"
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
	log.Println("Starting HTTP server at", bindAddr)
	openvpn := openvpnT{nil}
	http.HandleFunc("/openvpn/start", openvpn.start)
	http.HandleFunc("/openvpn/stop", openvpn.stop)
	http.HandleFunc("/firewall/start", firewallStartHandler)
	http.HandleFunc("/firewall/stop", firewallStopHandler)
	http.HandleFunc("/firewall/isup", firewallIsUpHandler)
	http.HandleFunc("/version", versionHandler)

	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

func (openvpn *openvpnT) start(w http.ResponseWriter, r *http.Request) {
	args, err := getArgs(r)
	if err != nil {
		log.Printf("An error has occurred processing flags: %v", err)
		w.Write([]byte(err.Error()))
		return
	}

	args = parseOpenvpnArgs(args)
	log.Printf("start openvpn: %v", args)
	err = openvpn.run(args)
	if err != nil {
		log.Printf("Error starting openvpn: %v", err)
		w.Write([]byte(err.Error()))
	}
}

func (openvpn *openvpnT) run(args []string) error {
	if openvpn.cmd != nil {
		log.Printf("openvpn was running, stop it first")
		err := openvpn.kill()
		if err != nil {
			return err
		}
	}
	log.Println("OPENVPN PATH:", getOpenvpnPath())

	// TODO: if it dies we should restart it
	openvpn.cmd = exec.Command(getOpenvpnPath(), args...)
	return openvpn.cmd.Start()
}

func (openvpn *openvpnT) stop(w http.ResponseWriter, r *http.Request) {
	log.Println("stop openvpn")
	if openvpn.cmd == nil || openvpn.cmd.ProcessState != nil {
		openvpn.cmd = nil
		return
	}

	err := openvpn.kill()
	if err != nil {
		log.Printf("Error stoping openvpn: %v", err)
		w.Write([]byte(err.Error()))
	}
}

func (openvpn *openvpnT) kill() error {
	err := kill(openvpn.cmd)
	if err == nil {
		openvpn.cmd.Wait()
	} else {
		log.Printf("Error killing the process: %v", err)
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
		log.Printf("An error has occurred processing gateways: %v", err)
		w.Write([]byte(err.Error()))
		return
	}

	for _, gw := range gateways {
		if !validAddress(gw) {
			w.Write([]byte("bad argument"))
		}
	}

	err = firewallStart(gateways, mode)
	if err != nil {
		log.Printf("Error starting firewall: %v", err)
		w.Write([]byte(err.Error()))
		return
	}
	log.Println("Start firewall: firewall started")
}

func firewallStopHandler(w http.ResponseWriter, r *http.Request) {
	err := firewallStop()
	if err != nil {
		log.Printf("Error stoping firewall: %v", err)
		w.Write([]byte(err.Error()))
	}
	log.Println("Stop firewall: firewall stopped")
}

func firewallIsUpHandler(w http.ResponseWriter, r *http.Request) {
	if firewallIsUp() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	} else {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("false"))
	}
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(AppName + "/" + Version + "\n"))
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

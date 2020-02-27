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

package helper

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
)

type openvpnT struct {
	cmd *exec.Cmd
}

func runCommandServer(bindAddr string) {
	openvpn := openvpnT{nil}
	http.HandleFunc("/openvpn/start", openvpn.start)
	http.HandleFunc("/openvpn/stop", openvpn.stop)
	http.HandleFunc("/firewall/start", firewallStartHandler)
	http.HandleFunc("/firewall/stop", firewallStopHandler)
	http.HandleFunc("/firewall/isup", firewallIsUpHandler)

	log.Fatal(http.ListenAndServe(bindAddr, nil))
}

func ServeHTTP(port int) {
	parseCliArgs()
	daemonize()
	doHandleCommands(port)
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
	gateways, err := getArgs(r)
	if err != nil {
		log.Printf("An error has occurred processing gateways: %v", err)
		w.Write([]byte(err.Error()))
		return
	}

	err = firewallStart(gateways)
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
		w.Write([]byte("true"))
		w.WriteHeader(http.StatusOK)
	} else {
		w.Write([]byte("false"))
		w.WriteHeader(http.StatusNoContent)
	}
}

func getArgs(r *http.Request) ([]string, error) {
	args := []string{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&args)
	return args, err
}

// +build !linux
// Copyright (C) 2018-2020 LEAP
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

package vpn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
)

type launcher struct {
	helperAddr string
}

const initialHelperPort = 7171

func probeHelperPort(port int) int {
	// this should be enough for a local reply
	timeout := time.Duration(500 * time.Millisecond)
	c := http.Client{Timeout: timeout}
	for {
		if smellsLikeOurHelperSpirit(port, &c) {
			return port
		}
		port++
		/* we could go until 65k, but there's really no need */
		if port > 10000 {
			break
		}
	}
	log.Println("WARN: Cannot find any working helper")
	return 0
}

func smellsLikeOurHelperSpirit(port int, c *http.Client) bool {
	uri := "http://localhost:" + strconv.Itoa(port) + "/version"
	resp, err := c.Get(uri)
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		ver, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return false
		}
		if strings.Contains(string(ver), config.ApplicationName) {
			log.Println("DEBUG: Successfully probed for matching helper at", uri)
			return true
		} else {
			log.Println("DEBUG: Another helper seems to be running:", string(ver))
			log.Println("DEBUG: But we were hoping to find:", config.ApplicationName)
		}
	}
	return false
}

func newLauncher() (*launcher, error) {
	helperPort := probeHelperPort(initialHelperPort)
	helperAddr := "http://localhost:" + strconv.Itoa(helperPort)
	return &launcher{helperAddr}, nil
}

func (l *launcher) close() error {
	return nil
}

func (l *launcher) check() (helpers bool, priviledge bool, err error) {
	return true, true, nil
}

func (l *launcher) openvpnStart(flags ...string) error {
	byteFlags, err := json.Marshal(flags)
	if err != nil {
		return err
	}
	return l.send("/openvpn/start", byteFlags)
}

func (l *launcher) openvpnStop() error {
	return l.send("/openvpn/stop", nil)
}

func (l *launcher) firewallStart(gateways []bonafide.Gateway) error {
	ipList := make([]string, len(gateways))
	for i, gw := range gateways {
		ipList[i] = gw.IPAddress
	}
	byteIPs, err := json.Marshal(ipList)
	if err != nil {
		return err
	}
	return l.send("/firewall/start", byteIPs)
}

func (l *launcher) firewallStop() error {
	return l.send("/firewall/stop", nil)
}

func (l *launcher) firewallIsUp() bool {
	var isup bool = false
	res, err := http.Post(l.helperAddr+"/firewall/isup", "", nil)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Got an error status code for firewall/isup: %v\n", res.StatusCode)
		isup = false
	} else {
		upStr, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Errorf("Error getting body for firewall/isup: %q", err)
			return false
		}
		isup, err = strconv.ParseBool(string(upStr))
		if err != nil {
			fmt.Errorf("Error parsing body for firewall/isup: %q", err)
			return false
		}
	}
	return isup
}

func (l *launcher) send(path string, body []byte) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	res, err := http.Post(l.helperAddr+path, "", reader)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resErr, err := ioutil.ReadAll(res.Body)
	if len(resErr) > 0 {
		/* FIXME why do we trigger a fatal with this error? */
		return fmt.Errorf("FATAL: Helper returned an error: %q", resErr)
	}
	return err
}

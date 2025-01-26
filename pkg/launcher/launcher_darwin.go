// Copyright (C) 2018-2021 LEAP
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

package launcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
)

type Launcher struct {
	helperAddr string
	client     *http.Client
	Failed     bool
	MngPass    string
}

const helperSocket = "bitmask-helper.sock"

func getHelperSocketPath() string {
	return filepath.Join("/tmp", helperSocket)
}

func probeHelper() (*http.Client, error) {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", getHelperSocketPath())
			},
		},
		Timeout: 500 * time.Millisecond, // this should be enough for a local reply
	}

	if smellsLikeOurHelperSpirit(&client) {
		return &client, nil
	}
	return nil, errors.New("Could not find port of helper backend")
}

func smellsLikeOurHelperSpirit(c *http.Client) bool {
	uri := "http://bitmask/version"
	resp, err := c.Get(uri)
	if err != nil {
		log.Warn().
			Err(err).
			Str("url", uri).
			Msg("Could not get")
		return false
	}
	if resp.StatusCode == 200 {
		ver, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not read web response")
			return false
		}
		if strings.Contains(string(ver), config.ProviderConfig.ApplicationName) {
			log.Debug().
				Str("url", uri).
				Msg("Successfully probed for matching helper")
			return true
		} else {
			log.Debug().
				Str("anotherHelper", string(ver)).
				Str("expectedHelper", config.ProviderConfig.ApplicationName).
				Msg("Found invalid helper already running")
		}
	}
	return false
}

func NewLauncher() (*Launcher, error) {
	client, err := probeHelper()
	if err != nil {
		return nil, err
	}
	return &Launcher{
		helperAddr: "http://bitmask",
		client:     client,
		Failed:     false,
	}, nil
}

func (l *Launcher) Close() error {
	return nil
}

func (l *Launcher) Check() (helpers bool, priviledge bool, err error) {
	return true, true, nil
}

func (l *Launcher) OpenvpnStart(flags ...string) error {
	byteFlags, err := json.Marshal(flags)
	if err != nil {
		return err
	}
	return l.send("/openvpn/start", byteFlags)
}

func (l *Launcher) OpenvpnStop() error {
	return l.send("/openvpn/stop", nil)
}

func (l *Launcher) FirewallStart(gateways []bonafide.Gateway) error {
	ipList := make([]string, len(gateways))
	for i, gw := range gateways {
		ipList[i] = gw.IPAddress
	}
	byteIPs, err := json.Marshal(ipList)
	if err != nil {
		return err
	}
	uri := "/firewall/start"
	if os.Getenv("UDP") == "1" {
		uri = uri + "?udp=1"
	}
	return l.send(uri, byteIPs)
}

func (l *Launcher) FirewallStop() error {
	return l.send("/firewall/stop", nil)
}

func (l *Launcher) FirewallIsUp() bool {
	var isup bool = false
	res, err := l.client.Post(l.helperAddr+"/firewall/isup", "", nil)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Warn().
			Int("statusCode", res.StatusCode).
			Msg("Got an error status code for firewall/isup")
		isup = false
	} else {
		upStr, err := io.ReadAll(res.Body)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not read body for firewall/isup")
			return false
		}
		isup, err = strconv.ParseBool(string(upStr))
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not parse body for firewall/isup")
			return false
		}
	}
	return isup
}

func (l *Launcher) send(path string, body []byte) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	res, err := l.client.Post(l.helperAddr+path, "", reader)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// read the error sent back by helper
	// in case of no error helper doesn't
	// write anything as reponse
	resErr, err := io.ReadAll(res.Body)
	if len(resErr) > 0 {
		return fmt.Errorf("FATAL: Helper returned an error: %q", resErr)
	}
	return err
}

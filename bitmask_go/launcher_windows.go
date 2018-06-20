// +build windows
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

package bitmask

import (
	"encoding/json"
	"net/textproto"
)

const (
	helperAddr = "localhost:7171"
)

type launcher struct {
	conn *textproto.Conn
}

func newLauncher() (*launcher, error) {
	conn, err := textproto.Dial("tcp", helperAddr)
	return &launcher{conn}, err
}

func (l *launcher) close() error {
	return l.conn.Close()
}

func (l *launcher) openvpnStart(flags ...string) error {
	return l.send("openvpn_start", flags...)
}

func (l *launcher) openvpnStop() error {
	return l.send("openvpn_stop")
}

func (l *launcher) firewallStart(gateways []gateway) error {
	return nil
}

func (l *launcher) firewallStop() error {
	return nil
}

func (l *launcher) send(cmd string, args ...string) error {
	if args == nil {
		args = []string{}
	}
	command := struct {
		Cmd  string   `json:"cmd"`
		Args []string `json:"args"`
	}{cmd, args}
	bytesCmd, err := json.Marshal(command)
	if err != nil {
		return err
	}

	_, err = l.conn.Cmd(string(bytesCmd))
	return err
}

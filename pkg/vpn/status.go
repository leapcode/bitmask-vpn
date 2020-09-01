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
	"fmt"
	"log"
	"strings"

	"github.com/apparentlymart/go-openvpn-mgmt/openvpn"
)

const (
	On       = "on"
	Off      = "off"
	Starting = "starting"
	Stopping = "stopping"
	Failed   = "failed"
)

var statusNames = map[string]string{
	"CONNECTING":   Starting,
	"WAIT":         Starting,
	"AUTH":         Starting,
	"GET_CONFIG":   Starting,
	"ASSIGN_IP":    Starting,
	"ADD_ROUTES":   Starting,
	"CONNECTED":    On,
	"RECONNECTING": Starting,
	"EXITING":      Stopping,
	"OFF":          Off,
	"FAILED":       Off,
}

func (b *Bitmask) openvpnManagement() {
	// TODO: we should warn the user on ListenAndServe errors
	newConnection := func(conn openvpn.IncomingConn) {
		eventCh := make(chan openvpn.Event, 10)
		log.Println("New connection into the management")
		b.managementClient = conn.Open(eventCh)
		b.managementClient.SetStateEvents(true)
		b.eventHandler(eventCh)
	}
	log.Fatal(openvpn.ListenAndServe(
		fmt.Sprintf("%s:%s", openvpnManagementAddr, openvpnManagementPort),
		openvpn.IncomingConnHandlerFunc(newConnection),
	))
}

func (b *Bitmask) eventHandler(eventCh <-chan openvpn.Event) {
	for event := range eventCh {
		log.Printf("Event: %v", event)
		stateEvent, ok := event.(*openvpn.StateEvent)
		if !ok {
			continue
		}
		statusName := stateEvent.NewState()
		status, ok := statusNames[statusName]
		if ok {
			b.statusCh <- status
		}
		if statusName == "CONNECTED" {
			b.onGateway = strings.Split(stateEvent.String(), ": ")[1]
			log.Println(">>> CONNECTED TO", b.onGateway)
		}
	}
	b.statusCh <- Off
}

func (b *Bitmask) getOpenvpnState() (string, error) {
	if b.managementClient == nil {
		return "", fmt.Errorf("No management connected")
	}
	stateEvent, err := b.managementClient.LatestState()
	if err != nil {
		return "", err
	}
	status, ok := statusNames[stateEvent.NewState()]
	if !ok {
		return "", fmt.Errorf("Unkonw status")
	}
	return status, nil
}

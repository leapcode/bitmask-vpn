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

package legacy

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/vpn/management"
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
	newConnection := func(conn management.IncomingConn) {
		eventCh := make(chan management.Event, 10)
		log.Info().Msg("New connection into the management")
		b.managementClient = conn.Open(eventCh)
		b.managementClient.SendPassword(b.launch.MngPass)
		b.managementClient.SetStateEvents(true)
		b.eventHandler(eventCh)
	}
	err := management.ListenAndServe(
		fmt.Sprintf("%s:%s", openvpnManagementAddr, openvpnManagementPort),
		management.IncomingConnHandlerFunc(newConnection),
	)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not run management backend")

	}
}

func (b *Bitmask) eventHandler(eventCh <-chan management.Event) {
	for event := range eventCh {
		log.Debug().
			Str("event", event.String()).
			Msg("Got event from OpenVPN process")
		stateEvent, ok := event.(*management.StateEvent)
		if !ok {
			continue
		}
		statusName := stateEvent.NewState()
		status, ok := statusNames[statusName]
		if ok {
			b.statusCh <- status
		}
		if statusName == "CONNECTED" {
			ip := strings.Split(stateEvent.String(), ": ")[1]
			if ip == "127.0.0.1" {
				// we're using pluggable transports
				b.onGateway = b.ptGateway
			} else {
				gw, err := b.bonafide.GetGatewayByIP(ip)
				if err == nil {
					b.onGateway = gw
					log.Info().
						Str("host", b.onGateway.Host).
						Msg("Connected to gateway")
				} else {
					log.Warn().
						Str("ip", ip).
						Msg("Connected to unknown gateway")
				}
			}
		}
	}
	b.statusCh <- Off
}

func (b *Bitmask) GetCurrentGateway() string {
	return b.onGateway.Host
}

func (b *Bitmask) GetCurrentLocation() string {
	return b.onGateway.LocationName
}

func (b *Bitmask) GetCurrentCountry() string {
	return b.onGateway.CountryCode
}

func (b *Bitmask) GetBestLocation(transport string) string {
	return b.bonafide.GetBestLocation(transport)
}

func (b *Bitmask) IsManualLocation() bool {
	return b.bonafide.IsManualLocation()
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

func (b *Bitmask) isFailed() bool {
	return b.launch.Failed
}

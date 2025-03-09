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
	"net"
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

const (
	openvpnManagementAddr = "127.0.0.1"
	openvpnManagementPort = "6061"
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
	"TCP_CONNECT":  Starting,
	"EXITING":      Stopping,
	"OFF":          Off,
	"FAILED":       Off,
}

var managementListener *management.MgmtListener

// Start listener on 127.0.0.1:6061 and run backend handler, the OpenVPN process
// connects to this port and tells us things like the state or the connected gateway.
// Our management backend is implemented in the management package (pkg/vpn/management)
// OpenVPN gets invoked with --management-client and --management 127.0.0.1 6061 <secret>
// Reference: https://openvpn.net/community-resources/management-interface/
func (b *Bitmask) initOpenVPNManagementHandler() {
	if managementListener != nil {
		if err := managementListener.Close(); err != nil {
			log.Warn().
				Err(err).
				Msg("failed to close openvpn management listener")
		}
	}
	listenAddr := net.JoinHostPort(openvpnManagementAddr, openvpnManagementPort)

	newConnection := func(conn management.IncomingConn) {
		eventCh := make(chan management.Event, 10)
		log.Debug().
			Str("endpoint", listenAddr).
			Msg("OpenVPN process connected to our management backend")
		b.managementClient = conn.Open(eventCh)
		b.managementClient.SendPassword(b.launch.MngPass)
		b.managementClient.SetStateEvents(true)
		b.eventHandler(eventCh)
	}

	ml, err := management.Listen(listenAddr)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to create management listener")
		return
	}
	// idealy this variable should be protected with a lock
	// since this method is run as a goroutine
	managementListener = ml

	if err := ml.Serve(management.IncomingConnHandlerFunc(newConnection)); err != nil {
		log.Warn().
			Err(err).
			Str("listen", listenAddr).
			Msgf("Could not start OpenVPN management backend")

	}
}

// Handle events sent by OpenVPN management. For more information please read
// the docstring of the initOpenVPNManagementHandler function
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
			state := strings.Split(stateEvent.String(), ":")
			ip := strings.TrimSpace(state[1])
			port := state[2]

			if ip == "127.0.0.1" {
				// we're using pluggable transports
				b.onGateway = b.ptGateway
			} else {
				gw, err := b.api.GetGatewayByIP(ip)
				if err == nil {
					b.onGateway = gw
					log.Info().
						Str("gateway", b.onGateway.Host).
						Str("port", port).
						Msg("Sucessfully connected to gateway")
				} else {
					log.Warn().
						Str("ip", ip).
						Str("port", port).
						Msg("Connected to unknown gateway")
				}
			}
		}
	}
	b.statusCh <- Off
}

// About the Getter functions here:
// In pkg/backend/status.go there is a function toJson which is called regularly by the cpp
// part (toJson is called in RefreshContext which is defined in pkg/backend/api.go and exported
// in libgoshim). It is used get the current state of the application. In toJson, all the Getter
// functions like GetCurrentGateway and IsManualLocation are called

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
	location, err := b.api.GetBestLocation(transport)
	if err != nil {
		log.Warn().
			Str("err", err.Error()).
			Str("transport", transport).
			Msg("Could not get best location")
	}
	// TODO: we return here an empty string in case of an error
	return location
}

func (b *Bitmask) IsManualLocation() bool {
	return b.api.IsManualLocation()
}

// The OpenVPN process connects to our management backend.
// If the OpenVPN state changes, it sends a line of text, e.g. "RECONNECTING: connection-reset"
// We use the map statusNames to set an internal state (e.g RECONNECTING results in state Starting)
// In getOpenvpnState we only use the last state that was set. It gets updated in eventHandler function
// The state is used by the GUI to handle the UI parts
func (b *Bitmask) getOpenvpnState() (string, error) {
	if b.managementClient == nil {
		log.Trace().
			Str("state", Off).
			Str("reason", "OpenVPN process has not (yet) connected to our management backend").
			Msg("Returning OpenVPN state")
		return Off, nil
	}
	stateEvent, err := b.managementClient.LatestState()
	if err != nil {
		log.Debug().
			Err(err).
			Msg("error fetching latest state from management interface")
		return "", err
	}
	status, ok := statusNames[stateEvent.NewState()]
	if !ok {
		return "", fmt.Errorf("Unknown status: %s", stateEvent.NewState())
	}
	return status, nil
}

func (b *Bitmask) isFailed() bool {
	return b.launch.Failed
}

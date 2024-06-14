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

package bitmask

import "0xacab.org/leap/bitmask-vpn/pkg/snowflake"

// This interface is used to implement different versions (v3 + v5)

type Bitmask interface {
	GetStatusCh() <-chan string
	GetSnowflakeCh() <-chan *snowflake.StatusEvent
	Close()
	Version() (string, error)
	StartVPN(provider string) error
	CanStartVPN() bool
	StopVPN() error
	Reconnect() error
	ReloadFirewall() error
	GetStatus() (string, error)
	VPNCheck() (helpers bool, priviledge bool, err error)
	GetLocationQualityMap(protocol string) map[string]float64
	GetLocationLabels(protocol string) map[string][]string
	GetBestLocation(protocol string) string
	UseGateway(name string)
	UseAutomaticGateway()
	SetProvider(string)
	GetTransport() string
	SetTransport(string) error
	UseUDP(bool)
	UseSnowflake(bool) error
	OffersUDP() bool
	GetCurrentGateway() string
	GetCurrentLocation() string
	GetCurrentCountry() string
	IsManualLocation() bool
	NeedsCredentials() bool
	DoLogin(username, password string) (bool, error)
	CanUpgrade() bool
	GetMotd() string
}

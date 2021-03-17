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

import (
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
)

type Bitmask interface {
	GetStatusCh() <-chan string
	Close()
	Version() (string, error)
	StartVPN(provider string) error
	CanStartVPN() bool
	StopVPN() error
	ReloadFirewall() error
	GetStatus() (string, error)
	InstallHelpers() error
	VPNCheck() (helpers bool, priviledge bool, err error)
	/* this is kind of breaking the abstract interface, maybe we don't need this anymore */
	ListGatewaysByCity(protocol string) (map[string]bonafide.Gateway, error)
	UseGateway(name string) error
	GetCurrentGateway() string
	GetCurrentLocation() string
	GetGatewayDetails(label string) (interface{}, error)
	UseTransport(transport string) error
	NeedsCredentials() bool
	DoLogin(username, password string) (bool, error)
}

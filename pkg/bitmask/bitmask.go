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
	ListLocationFullness(protocol string) map[string]float64
	UseGateway(name string)
	UseAutomaticGateway()
	GetCurrentGateway() string
	GetCurrentLocation() string
	GetCurrentCountry() string
	IsManualLocation() bool
	UseTransport(transport string) error
	NeedsCredentials() bool
	DoLogin(username, password string) (bool, error)
}

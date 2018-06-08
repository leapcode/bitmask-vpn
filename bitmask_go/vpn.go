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
	"path"
)

const (
	openvpnManagementAddr = "127.0.0.1"
	openvpnManagementPort = "6061"
)

var gateways = []string{
	"5.79.86.180",
	"199.58.81.145",
	"198.252.153.28",
}

// StartVPN for provider
func (b *Bitmask) StartVPN(provider string) error {
	// TODO: openvpn args are hardcoded
	err := firewallStart(gateways)
	if err != nil {
		return err
	}

	arg := []string{"--nobind", "--verb", "1"}
	for _, gw := range gateways {
		arg = append(arg, "--remote", gw, "443", "tcp4")
	}
	certPemPath := b.getCertPemPath()
	arg = append(arg, "--client", "--tls-client", "--remote-cert-tls", "server", "--tls-cipher", "DHE-RSA-AES128-SHA", "--cipher", "AES-128-CBC", "--tun-ipv6", "--auth", "SHA1", "--keepalive", "10 30", "--management-client", "--management", openvpnManagementAddr+" "+openvpnManagementPort, "--ca", b.getCaCertPath(), "--cert", certPemPath, "--key", certPemPath)
	return openvpnStart(arg...)
}

// StopVPN or cancel
func (b *Bitmask) StopVPN() error {
	err := firewallStop()
	if err != nil {
		return err
	}
	return openvpnStop()
}

// GetStatus returns the VPN status
func (b *Bitmask) GetStatus() (string, error) {
	status, err := b.getOpenvpnState()
	if err != nil {
		status = Off
	}
	return status, nil
}

// InstallHelpers into the system
func (b *Bitmask) InstallHelpers() error {
	// TODO
	return nil
}

// VPNCheck returns if the helpers are installed and up to date and if polkit is running
func (b *Bitmask) VPNCheck() (helpers bool, priviledge bool, err error) {
	// TODO
	return true, true, nil
}

// ListGateways return the names of the gateways
func (b *Bitmask) ListGateways(provider string) ([]string, error) {
	// TODO
	return []string{}, nil
}

// UseGateway selects name as the default gateway
func (b *Bitmask) UseGateway(name string) error {
	// TODO
	return nil
}

func (b *Bitmask) getCertPemPath() string {
	return path.Join(b.tempdir, "openvpn.pem")
}

func (b *Bitmask) getCaCertPath() string {
	return path.Join(b.tempdir, "cacert.pem")
}

// SPDX-FileCopyrightText: 2018 LEAP
// SPDX-License-Identifier: GPL-3.0-or-later
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

package standalone

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"0xacab.org/leap/shapeshifter"
)

const (
	openvpnManagementAddr = "127.0.0.1"
	openvpnManagementPort = "6061"
)

// StartVPN for provider
func (b *Bitmask) StartVPN(provider string) error {
	var proxy string
	if b.transport != "" {
		var err error
		proxy, err = b.startTransport()
		if err != nil {
			return err
		}
	}

	return b.startOpenVPN(proxy)
}

func (b *Bitmask) startTransport() (proxy string, err error) {
	proxy = "127.0.0.1:4430"
	if b.shapes != nil {
		return proxy, nil
	}

	gateways, err := b.bonafide.GetGateways(b.transport)
	if err != nil {
		return "", err
	}
	if len(gateways) == 0 {
		log.Printf("No gateway for transport %s in provider", b.transport)
		return "", nil
	}

	for _, gw := range gateways {
		if _, ok := gw.Options["cert"]; !ok {
			continue
		}
		b.shapes = &shapeshifter.ShapeShifter{
			Cert:      gw.Options["cert"],
			Target:    gw.IPAddress + ":" + gw.Ports[0],
			SocksAddr: proxy,
		}
		go b.listenShapeErr()
		if iatMode, ok := gw.Options["iat-mode"]; ok {
			b.shapes.IatMode, err = strconv.Atoi(iatMode)
			if err != nil {
				b.shapes.IatMode = 0
			}
		}
		err = b.shapes.Open()
		if err != nil {
			log.Printf("Can't connect to transport %s: %v", b.transport, err)
			continue
		}
		return proxy, nil
	}
	return "", fmt.Errorf("No working gateway for transport %s: %v", b.transport, err)
}

func (b *Bitmask) listenShapeErr() {
	ch := b.shapes.GetErrorChannel()
	for {
		err, more := <-ch
		if !more {
			return
		}
		log.Printf("Error from shappeshifter: %v", err)
	}
}

func (b *Bitmask) startOpenVPN(proxy string) error {
	certPemPath, err := b.getCert()
	if err != nil {
		return err
	}
	arg, err := b.bonafide.GetOpenvpnArgs()
	if err != nil {
		return err
	}

	if proxy == "" {
		gateways, err := b.bonafide.GetGateways("openvpn")
		if err != nil {
			return err
		}
		err = b.launch.firewallStart(gateways)
		if err != nil {
			return err
		}

		for _, gw := range gateways {
			for _, port := range gw.Ports {
				arg = append(arg, "--remote", gw.IPAddress, port, "tcp4")
			}
		}
	} else {
		gateways, err := b.bonafide.GetGateways(b.transport)
		if err != nil {
			return err
		}
		err = b.launch.firewallStart(gateways)
		if err != nil {
			return err
		}

		proxyArgs := strings.Split(proxy, ":")
		arg = append(arg, "--remote", proxyArgs[0], proxyArgs[1], "tcp4")
	}
	arg = append(arg,
		"--verb", "1",
		"--management-client",
		"--management", openvpnManagementAddr, openvpnManagementPort,
		"--ca", b.getCaCertPath(),
		"--cert", certPemPath,
		"--key", certPemPath)
	return b.launch.openvpnStart(arg...)
}

func (b *Bitmask) getCert() (certPath string, err error) {
	certPath = b.getCertPemPath()

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		cert, err := b.bonafide.GetCertPem()
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(certPath, cert, 0600)
	}

	return certPath, err
}

// StopVPN or cancel
func (b *Bitmask) StopVPN() error {
	err := b.launch.firewallStop()
	if err != nil {
		return err
	}
	if b.shapes != nil {
		b.shapes.Close()
		b.shapes = nil
	}
	return b.launch.openvpnStop()
}

// ReloadFirewall restarts the firewall
func (b *Bitmask) ReloadFirewall() error {
	err := b.launch.firewallStop()
	if err != nil {
		return err
	}

	status, err := b.GetStatus()
	if err != nil {
		return err
	}

	if status != Off {
		gateways, err := b.bonafide.GetGateways("openvpn")
		if err != nil {
			return err
		}
		return b.launch.firewallStart(gateways)
	}
	return nil
}

// GetStatus returns the VPN status
func (b *Bitmask) GetStatus() (string, error) {
	status, err := b.getOpenvpnState()
	if err != nil {
		status = Off
	}
	if status == Off && b.launch.firewallIsUp() {
		return Failed, nil
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
	return b.launch.check()
}

// ListGateways return the names of the gateways
func (b *Bitmask) ListGateways(provider string) ([]string, error) {
	gateways, err := b.bonafide.GetGateways("openvpn")
	if err != nil {
		return nil, err
	}
	gatewayNames := make([]string, len(gateways))
	for i, gw := range gateways {
		gatewayNames[i] = gw.Location
	}
	return gatewayNames, nil
}

// UseGateway selects name as the default gateway
func (b *Bitmask) UseGateway(name string) error {
	b.bonafide.SetDefaultGateway(name)
	return nil
}

// UseTransport selects an obfuscation transport to use
func (b *Bitmask) UseTransport(transport string) error {
	if transport != "obfs4" {
		return fmt.Errorf("Transport %s not implemented", transport)
	}
	b.transport = transport
	return nil
}

func (b *Bitmask) getCertPemPath() string {
	return path.Join(b.tempdir, "openvpn.pem")
}

func (b *Bitmask) getCaCertPath() string {
	return path.Join(b.tempdir, "cacert.pem")
}

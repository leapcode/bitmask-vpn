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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/shapeshifter"
)

const (
	openvpnManagementAddr = "127.0.0.1"
	openvpnManagementPort = "6061"
)

// StartVPN for provider
func (b *Bitmask) StartVPN(provider string) error {
	if !b.CanStartVPN() {
		return errors.New("BUG: cannot start vpn")
	}

	var err error
	b.certPemPath, err = b.getCert()
	if err != nil {
		return err
	}
	b.openvpnArgs, err = b.bonafide.GetOpenvpnArgs()
	if err != nil {
		return err
	}

	return b.startOpenVPN()
}

func (b *Bitmask) CanStartVPN() bool {
	/* FIXME this is not enough. We should check, if provider needs
	* credentials, if we have a valid token, otherwise remove it and
	make sure that we're asking for the credentials input */
	return !b.bonafide.NeedsCredentials()
}

func (b *Bitmask) startTransport(host string) (proxy string, err error) {
	// TODO configure port if not available
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
		if gw.Host != host {
			continue
		}
		if _, ok := gw.Options["cert"]; !ok {
			continue
		}
		log.Println("Selected Gateway:", gw.Host, gw.IPAddress)
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
		log.Println("Connected via obfs4 to", gw.IPAddress, "(", gw.Host, ")")
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

func (b *Bitmask) startOpenVPN() error {
	arg := []string{}
	// Empty transport means we get only the openvpn gateways
	if b.transport == "" {
		arg = b.openvpnArgs
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
		// For now, obf4 is the only supported Pluggable Transport
		gateways, err := b.bonafide.GetGateways(b.transport)
		if err != nil {
			return err
		}
		if len(gateways) == 0 {
			log.Printf("ERROR No gateway for transport %s in provider", b.transport)
			return errors.New("ERROR: cannot find any gateway for selected transport")
		}

		gw := gateways[0]
		proxy, err := b.startTransport(gw.Host)
		if err != nil {
			return err
		}
		b.ptGateway = gw

		err = b.launch.firewallStart(gateways)
		if err != nil {
			return err
		}

		proxyArgs := strings.Split(proxy, ":")
		arg = append(arg, "--remote", proxyArgs[0], proxyArgs[1], "tcp4")
		arg = append(arg, "--route", gw.IPAddress, "255.255.255.255", "net_gateway")
	}
	arg = append(arg,
		"--verb", "3",
		"--management-client",
		"--management", openvpnManagementAddr, openvpnManagementPort,
		"--ca", b.getTempCaCertPath(),
		"--cert", b.certPemPath,
		"--key", b.certPemPath,
		"--persist-tun")
	return b.launch.openvpnStart(arg...)
}

func (b *Bitmask) getCert() (certPath string, err error) {
	persistentCertFile := filepath.Join(config.Path, strings.ToLower(config.Provider)+".pem")
	if _, err := os.Stat(persistentCertFile); !os.IsNotExist(err) && isValidCert(persistentCertFile) {
		// reuse cert. for the moment we're not writing one there, this is
		// only to allow users to get certs off-band and place them there
		// as a last-resort fallback for circumvention.
		certPath = persistentCertFile
		err = nil
	} else {
		// download one fresh
		certPath = b.getTempCertPemPath()
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			log.Println("Fetching certificate to", certPath)
			cert, err := b.bonafide.GetPemCertificate()
			if err != nil {
				return "", err
			}
			err = ioutil.WriteFile(certPath, cert, 0600)
		}
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

// Reconnect to the VPN
func (b *Bitmask) Reconnect() error {
	if !b.CanStartVPN() {
		return errors.New("BUG: cannot start vpn")
	}

	status, err := b.GetStatus()
	if err != nil {
		return err
	}
	log.Println("reconnect")
	if status != Off {
		if b.shapes != nil {
			b.shapes.Close()
			b.shapes = nil
		}
		err = b.launch.openvpnStop()
		if err != nil {
			return err
		}
	}

	err = b.launch.firewallStop()
	if err != nil {
		return err
	}
	return b.startOpenVPN()
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
		gateways, err := b.bonafide.GetAllGateways("any")
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

func (b *Bitmask) InstallHelpers() error {
	// TODO use pickle module from here
	return nil
}

// VPNCheck returns if the helpers are installed and up to date and if polkit is running
func (b *Bitmask) VPNCheck() (helpers bool, privilege bool, err error) {
	return b.launch.check()
}

func (b *Bitmask) ListLocationFullness(transport string) map[string]float64 {
	return b.bonafide.ListLocationFullness(transport)
}

// UseGateway selects a gateway, by label, as the default gateway
func (b *Bitmask) UseGateway(label string) {
	b.bonafide.SetManualGateway(label)
}

// UseAutomaticGateway sets the gateway to be selected automatically
// best gateway will be used
func (b *Bitmask) UseAutomaticGateway() {
	b.bonafide.SetAutomaticGateway()
}

// UseTransport selects an obfuscation transport to use
func (b *Bitmask) UseTransport(transport string) error {
	if transport != "obfs4" {
		return fmt.Errorf("Transport %s not implemented", transport)
	}
	b.transport = transport
	return nil
}

func (b *Bitmask) getTempCertPemPath() string {
	return path.Join(b.tempdir, "openvpn.pem")
}

func (b *Bitmask) getTempCaCertPath() string {
	return path.Join(b.tempdir, "cacert.pem")
}

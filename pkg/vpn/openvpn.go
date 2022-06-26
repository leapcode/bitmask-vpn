// Copyright (C) 2018-2021 LEAP
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
	obfsvpn "0xacab.org/leap/obfsvpn/client"
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
	// TODO configure socks port if not available
	// TODO get port from UI/config file
	proxyAddr := "127.0.0.1:8080"

	if b.obfsvpnProxy != nil {
		return proxyAddr, nil
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

		kcpMode := false
		if os.Getenv("LEAP_KCP") == "1" {
			kcpMode = true
		}

		log.Println("connecting with cert:", gw.Options["cert"])

		b.obfsvpnProxy = obfsvpn.NewClient(kcpMode, proxyAddr, gw.Options["cert"])
		go func() {
			_, err = b.obfsvpnProxy.Start()
			if err != nil {
				log.Printf("Can't connect to transport %s: %v", b.transport, err)
			}
			log.Println("Connected via obfs4 to", gw.IPAddress, "(", gw.Host, ")")
		}()

		return proxyAddr, nil
	}
	return "", fmt.Errorf("No working gateway for transport %s: %v", b.transport, err)
}

func (b *Bitmask) startOpenVPN() error {
	arg := b.openvpnArgs
	b.statusCh <- Starting
	if b.GetTransport() == "obfs4" {
		gateways, err := b.bonafide.GetGateways("obfs4")
		if err != nil {
			return err
		}
		if len(gateways) == 0 {
			log.Printf("ERROR No gateway for transport %s in provider", b.transport)
			return errors.New("ERROR: cannot find any gateway for selected transport")
		}

		gw := gateways[0]
		b.ptGateway = gw

		proxy, err := b.startTransport(gw.Host)
		if err != nil {
			// TODO this is not going to return the error since it blocks
			// we need to get an error channel from obfsvpn.
			return err
		}

		err = b.launch.firewallStart(gateways)
		if err != nil {
			return err
		}

		proxyArgs := strings.Split(proxy, ":")
		arg = append(arg, "--socks-proxy", proxyArgs[0], proxyArgs[1])
		arg = append(arg, "--remote", gw.IPAddress, gw.Ports[0], "tcp4")
		arg = append(arg, "--route", gw.IPAddress, "255.255.255.255", "net_gateway")
	} else {
		log.Println("args passed to bitmask-root:", arg)
		gateways, err := b.bonafide.GetGateways("openvpn")
		if err != nil {
			return err
		}
		if b.udp {
			os.Setenv("UDP", "1")
		} else {
			os.Setenv("UDP", "0")
		}
		err = b.launch.firewallStart(gateways)
		if err != nil {
			return err
		}

		for _, gw := range gateways {
			for _, port := range gw.Ports {
				if port != "53" {
					if b.udp {
						arg = append(arg, "--remote", gw.IPAddress, port, "udp4")
					} else {
						arg = append(arg, "--remote", gw.IPAddress, port, "tcp4")
					}
				}
			}
		}
	}
	openvpnVerb := os.Getenv("OPENVPN_VERBOSITY")
	verb, err := strconv.Atoi(openvpnVerb)
	if err != nil || verb > 6 || verb < 3 {
		openvpnVerb = "3"
	}
	// TODO we need to check if the openvpn options pushed by server are
	// not overriding (or duplicating) some of the options we're adding here.
	log.Println("VERB", verb)

	arg = append(arg,
		"--verb", openvpnVerb,
		"--management-client",
		"--management", openvpnManagementAddr, openvpnManagementPort,
		"--ca", b.getTempCaCertPath(),
		"--cert", b.certPemPath,
		"--key", b.certPemPath,
		"--persist-tun") // needed for reconnects
	//		"--float")
	if verb > 3 {
		arg = append(
			arg,
			"--log", "/tmp/leap-vpn.log")
	}
	if os.Getenv("LEAP_DRYRUN") == "1" {
		arg = append(
			arg,
			"--pull-filter", "ignore", "route")
	}
	return b.launch.openvpnStart(arg...)
}

func (b *Bitmask) getCert() (certPath string, err error) {
	log.Println("Getting certificate...")
	failed := false
	persistentCertFile := filepath.Join(config.Path, strings.ToLower(config.Provider)+".pem")
	if _, err := os.Stat(persistentCertFile); !os.IsNotExist(err) && isValidCert(persistentCertFile) {
		// TODO snowflake might have written a cert here
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
				log.Println(err)
				failed = true
			}
			err = ioutil.WriteFile(certPath, cert, 0600)
			if err != nil {
				failed = true
			}
		}
	}
	if failed || !isValidCert(certPath) {
		d := config.APIURL[8 : len(config.APIURL)-1]
		logDnsLookup(d)
		cert, err := b.bonafide.GetPemCertificateNoDNS()
		if cert != nil {
			log.Println("Successfully did certificate bypass")
			err = nil
		} else {
			err = errors.New("Cannot get vpn certificate")
		}
		err = ioutil.WriteFile(certPath, cert, 0600)
		if err != nil {
			failed = true
		}
	}

	return certPath, err
}

// Explicit call to GetGateways, to be able to fetch them all before starting the vpn
func (b *Bitmask) fetchGateways() {
	log.Println("Fetching gateways...")
	_, err := b.bonafide.GetAllGateways(b.transport)
	if err != nil {
		log.Println("ERROR Cannot fetch gateways")
	}
}

// StopVPN or cancel
func (b *Bitmask) StopVPN() error {
	err := b.launch.firewallStop()
	if err != nil {
		return err
	}
	if b.obfsvpnProxy != nil {
		b.obfsvpnProxy.Stop()
		b.obfsvpnProxy = nil
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
	log.Println("DEBUG Reconnecting")
	if status != Off {
		b.statusCh <- Stopping
		if b.obfsvpnProxy != nil {
			b.obfsvpnProxy.Stop()
			b.obfsvpnProxy = nil
		}
		err = b.launch.openvpnStop()
		if err != nil {
			return err
		}
	}

	err = b.launch.firewallStop()
	// FIXME - there's a window in which we might leak traffic here!
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
	status := Off
	if b.isFailed() {
		status = Failed
	} else {
		status, err := b.getOpenvpnState()
		if err != nil {
			status = Off
		}
		if status == Off && b.launch.firewallIsUp() {
			return Failed, nil
		}
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

func (b *Bitmask) ListLocationLabels(transport string) map[string][]string {
	return b.bonafide.ListLocationLabels(transport)
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

// SetTransport selects an obfuscation transport to use
func (b *Bitmask) SetTransport(t string) error {
	if t != "openvpn" && t != "obfs4" {
		return fmt.Errorf("Transport %s not implemented", t)
	}
	log.Println("Setting transport to", t)
	// compare and set string looks strange, but if assigning directly
	// we're getting some kind of corruption with the transport string.
	// I suspect something's
	// not quite right with the c<->go char pointers handling.
	if t == "obfs4" {
		b.transport = "obfs4"
	} else if t == "openvpn" {
		b.transport = "openvpn"
	}
	return nil
}

// GetTransport gets the obfuscation transport to use. Only obfs4 available for now.
func (b *Bitmask) GetTransport() string {
	if b.transport == "obfs4" {
		return "obfs4"
	} else {
		return "openvpn"
	}
}

func (b *Bitmask) getTempCertPemPath() string {
	return path.Join(b.tempdir, "openvpn.pem")
}

func (b *Bitmask) getTempCaCertPath() string {
	return path.Join(b.tempdir, "cacert.pem")
}

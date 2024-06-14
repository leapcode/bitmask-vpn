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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	obfsvpn "0xacab.org/leap/obfsvpn/client"
)

const (
	openvpnManagementAddr = "127.0.0.1"
	openvpnManagementPort = "6061"
)

// StartVPN for provider
func (b *Bitmask) StartVPN(provider string) error {
	if !b.CanStartVPN() {
		log.Warn().Msg("BUG cannot start")
		return errors.New("BUG: cannot start vpn")
	}

	var err error
	err = b.getCert()
	if err != nil {
		return err
	}
	b.openvpnArgs, err = b.api.GetOpenvpnArgs()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return b.startOpenVPN(ctx)
}

func (b *Bitmask) CanStartVPN() bool {
	/* FIXME this is not enough. We should check, if provider needs
	* credentials, if we have a valid token, otherwise remove it and
	make sure that we're asking for the credentials input */
	return !b.api.NeedsCredentials()
}

func (b *Bitmask) startTransportForPrivateBridge(ctx context.Context, gw bonafide.Gateway) (proxy string, err error) {
	proxyAddr := "127.0.0.1:8080"
	kcpMode := false
	if os.Getenv("LEAP_KCP") == "1" {
		kcpMode = true
	}
	b.obfsvpnProxy = obfsvpn.NewClient(ctx, kcpMode, proxyAddr, gw.Options["cert"]).(*obfsvpn.Client)
	go func() {
		_, err = b.obfsvpnProxy.Start()
		if err != nil {
			log.Warn().
				Err(err).
				Str("transport", b.transport).
				Msg("Could not connect to transport")
		}
		log.Info().
			Str("ip", gw.IPAddress).
			Str("host", gw.Host).
			Msg("Connected via obfs4")
	}()

	return proxyAddr, nil
}

func (b *Bitmask) startTransport(ctx context.Context, host string) (proxy string, err error) {
	log.Debug().
		Str("host", host).
		Msg("Starting transport")

	// TODO configure socks port if not available
	// TODO get port from UI/config file
	proxyAddr := "127.0.0.1:8080"

	if b.obfsvpnProxy != nil {
		return proxyAddr, nil
	}

	gateways, err := b.api.GetBestGateways(b.transport)
	if err != nil {
		return "", err
	}
	if len(gateways) == 0 {
		log.Warn().
			Str("transport", b.transport).
			Msg("No gateway for transport in provider")
		return "", nil
	}

	for _, gw := range gateways {
		if gw.Host != host {
			continue
		}
		if _, ok := gw.Options["cert"]; !ok {
			continue
		}
		log.Info().
			Str("host", gw.Host).
			Str("ip", gw.IPAddress).
			Msg("Selected Gateway")

		kcpMode := false
		if os.Getenv("LEAP_KCP") == "1" {
			kcpMode = true
		}

		log.Debug().
			Str("host", gw.Host).
			Str("ip", gw.IPAddress).
			Bool("kcp", kcpMode).
			Str("cert", gw.Options["cert"]).
			Str("proxyAddr", proxyAddr).
			Str("transport", b.transport).
			Msg("Using gateway")

		b.obfsvpnProxy = obfsvpn.NewClient(ctx, kcpMode, proxyAddr, gw.Options["cert"]).(*obfsvpn.Client)
		go func() {
			_, err = b.obfsvpnProxy.Start()
			if err != nil {
				log.Warn().
					Err(err).
					Str("transport", b.transport).
					Msg("Could not connect to transport")
			}
			log.Info().
				Str("ip", gw.IPAddress).
				Str("host", gw.Host).
				Msg("Connected via obfs4")
		}()

		return proxyAddr, nil
	}
	return "", fmt.Errorf("No working gateway for transport %s: %v", b.transport, err)
}

func maybeGetPrivateGateway() (bonafide.Gateway, bool) {
	gw := bonafide.Gateway{}
	privateBridge := os.Getenv("LEAP_PRIVATE_BRIDGE")
	if privateBridge == "" {
		return gw, false
	}
	obfs4Cert := os.Getenv("LEAP_PRIVATE_BRIDGE_CERT")
	if privateBridge == "" {
		return gw, false
	}
	bridgeArgs := strings.Split(privateBridge, ":")
	gw.Host = bridgeArgs[0]
	gw.Ports = []string{bridgeArgs[1]}
	opt := make(map[string]string)
	opt["cert"] = obfs4Cert
	gw.Options = opt
	return gw, true
}

// generates a password and returns the path for a temporary file where this password is written
func (b *Bitmask) generateManagementPassword() string {
	pass := getRandomPass(12)
	tmpFile, err := ioutil.TempFile(b.tempdir, "leap-vpn-")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not create temporary file")
	}
	tmpFile.Write([]byte(pass))
	b.launch.MngPass = pass
	return tmpFile.Name()
}

func (b *Bitmask) startOpenVPN(ctx context.Context) error {
	arg := b.openvpnArgs
	/*
		XXX has this changed??
		 arg, err := b.bonafide.GetOpenvpnArgs()
		 if err != nil {
		 	return err
		 }
	*/
	/*
		XXX and this??
		 certPemPath, err := b.getCert()
		 if err != nil {
		 	return err
		 }
	*/
	b.statusCh <- Starting
	if b.GetTransport() == "obfs4" {

		if config.ApiVersion == 5 {
			// if I return an error, the GUI state does not get updated properly to Failed/Stopped and
			// continues to stay in state Connecting (also clicking Cancel doesnot work)
			log.Fatal().Msg("Could not start OpenVPN with obfs4. This is currently not supported via v5")
			// menshen/v5 has different api endpoints: gateways and bridges
			// gw.Options is always empty right now
		}

		var gw bonafide.Gateway
		var gateways []bonafide.Gateway
		var proxy string

		gw, gotPrivate := maybeGetPrivateGateway()
		if gotPrivate {
			var err error
			log.Info().
				Str("host", gw.Host).
				Msgf("Got a private bridge with options: %v", gw.Options)
			gateways = []bonafide.Gateway{gw}
			proxy, err = b.startTransportForPrivateBridge(ctx, gw)
			if err != nil {
				// TODO this is not going to return the error since it blocks
				// we need to get an error channel from obfsvpn.
				return err
			}
		} else {
			// get a gateway from bonafide looking at the services announced in eip-service

			log.Debug().Msg("Getting a gateway with obfs4 transport...")

			gateways, err := b.api.GetBestGateways("obfs4")
			if err != nil {
				return err
			}
			if len(gateways) == 0 {
				log.Warn().
					Str("transport", b.transport).
					Msg("No gateway for transport in provider")
				return errors.New("ERROR: cannot find any gateway for selected transport")
			}

			gw = gateways[0]
			b.ptGateway = gw

			proxy, err = b.startTransport(ctx, gw.Host)
			if err != nil {
				// TODO this is not going to return the error since it blocks
				// we need to get an error channel from obfsvpn.
				return err
			}
		}

		err := b.launch.FirewallStart(gateways)
		if err != nil {
			return err
		}

		proxyArgs := strings.Split(proxy, ":")
		arg = append(arg, "--socks-proxy", proxyArgs[0], proxyArgs[1])
		arg = append(arg, "--remote", gw.IPAddress, gw.Ports[0], "tcp4")
		arg = append(arg, "--route", gw.IPAddress, "255.255.255.255", "net_gateway")
	} else {

		gateways, err := b.api.GetBestGateways("openvpn")
		if err != nil {
			return err
		}
		log.Info().Msgf("Got best gateway %v", gateways)

		// env UDP is used by bitmask-root helper
		if b.useUDP {
			os.Setenv("UDP", "1")
		} else {
			os.Setenv("UDP", "0")
		}
		err = b.launch.FirewallStart(gateways)
		if err != nil {
			return err
		}

		var proto string
		for _, gw := range gateways {
			for _, port := range gw.Ports {
				// issue about udp/53: https://0xacab.org/leap/bitmask-vpn/-/issues/796
				if port != "53" {
					if b.useUDP {
						proto = "udp4"
					} else {
						proto = "tcp4"
					}
					arg = append(arg, "--remote", gw.IPAddress, port, proto)
					log.Debug().
						Str("gateway", gw.Host).
						Str("ip4", gw.IPAddress).
						Str("port", port).
						Str("proto", proto).
						Msg("Adding gateway to command line via --remote")
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
	log.Debug().
		Str("verb", openvpnVerb).
		Msg("Setting OpenVPN verbosity")

	passFile := b.generateManagementPassword()

	arg = append(arg,
		"--verb", openvpnVerb,
		"--management-client",
		"--management", openvpnManagementAddr, openvpnManagementPort, passFile,
		"--ca", b.getTempCaCertPath(),
		"--cert", b.certPemPath,
		"--key", b.certPemPath,
		"--persist-tun") // needed for reconnects
	//		"--float")

	if os.Getenv("OPENVPN_LOG_TO_FILE") != "" {
		openVpnLogFile := filepath.Join(os.TempDir(), "leap-vpn.log")
		log.Debug().
			Str("logFile", openVpnLogFile).
			Msg("Telling OpenVPN to log to a file")
		arg = append(arg, "--log", openVpnLogFile)
	}

	if os.Getenv("LEAP_DRYRUN") == "1" {
		log.Debug().Msg("Not routing traffic over OpenVPN (LEAP_DRYRUN=1)")
		arg = append(arg, "--pull-filter", "ignore", "route")
	}
	return b.launch.OpenvpnStart(arg...)
}

// Get valid client credentials (key + cert) from menshen. Currently, there is no caching implemented
func (b *Bitmask) getCert() error {
	log.Info().Msg("Getting OpenVPN client certificate")

	persistentCertFile := filepath.Join(config.Path, strings.ToLower(config.Provider)+".pem")
	// snowflake might have written a cert here
	// reuse cert. for the moment we're not writing one there, this is
	// only to allow users to get certs off-band and place them there
	// as a last-resort fallback for circumvention.
	if _, err := os.Stat(persistentCertFile); !os.IsNotExist(err) && isValidCert(persistentCertFile) {
		log.Trace().
			Str("persistentCertFile", persistentCertFile).
			Msg("Found local client credentials")
		b.certPemPath = persistentCertFile
		return nil
	}

	b.certPemPath = b.getTempCertPemPath()
	// If we start OpenVPN, openvpn.pem does not exist and isValidCert returns false
	// If we start OpenVPN later again (not restarting the  client), there
	// should be a valid openvpn.pem
	// If there is no valid openvpn.pem, fetch a new one from menshen
	// Note: b.tempdir is unique for every run of the desktop client
	if !isValidCert(b.certPemPath) {
		cert, err := b.api.GetPemCertificate()
		if err != nil {
			// if we can't speak with API => resolve DNS and log
			url, err := url.Parse(config.APIURL)
			if err != nil {
				log.Warn().
					Err(err).
					Str("apiUrl", config.APIURL).
					Msg("Could not parse domain out of API URL")
			}
			logDnsLookup(url.Host)
			return err
		}
		err = ioutil.WriteFile(b.certPemPath, cert, 0600)
		if err != nil {
			return err
		}
		if !isValidCert(b.certPemPath) {
			return fmt.Errorf("The certificate given by API is invalid")
		}
	}
	return nil
}

// Explicit call to GetGateways, to be able to fetch them all before starting the vpn
func (b *Bitmask) fetchGateways() {
	log.Info().Msg("Fetching gateways...")
	_, err := b.api.GetAllGateways(b.transport)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not fetch gateways")
	}
}

// StopVPN or cancel
func (b *Bitmask) StopVPN() error {
	err := b.launch.FirewallStop()
	if err != nil {
		return err
	}
	if b.obfsvpnProxy != nil {
		b.obfsvpnProxy.Stop()
		b.obfsvpnProxy = nil
	}
	b.tryStopFromManagement()
	b.launch.OpenvpnStop()
	return nil
}

func (b *Bitmask) tryStopFromManagement() {
	if b.managementClient != nil {
		b.managementClient.SendSignal("SIGTERM")
	}
}

// Reconnect to the VPN
func (b *Bitmask) Reconnect() error {
	log.Info().Msgf("Restarting OpenVPN")

	if !b.CanStartVPN() {
		return errors.New("BUG: cannot start vpn (CanStartVPN)")
	}

	status, err := b.GetStatus()
	if err != nil {
		return err
	}

	if status != Off {
		b.statusCh <- Stopping
		if b.obfsvpnProxy != nil {
			b.obfsvpnProxy.Stop()
			b.obfsvpnProxy = nil
		}
		err = b.launch.OpenvpnStop()
		if err != nil {
			return err
		}
	}

	err = b.launch.FirewallStop()
	// FIXME - there's a window in which we might leak traffic here!
	if err != nil {
		return err
	}
	ctx := context.Background()
	return b.startOpenVPN(ctx)
}

// ReloadFirewall restarts the firewall
func (b *Bitmask) ReloadFirewall() error {
	err := b.launch.FirewallStop()
	if err != nil {
		return err
	}

	status, err := b.GetStatus()
	if err != nil {
		return err
	}

	if status != Off {
		gateways, err := b.api.GetAllGateways("any")
		if err != nil {
			return err
		}
		return b.launch.FirewallStart(gateways)
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
		if status == Off && b.launch.FirewallIsUp() {
			return Failed, nil
		}
	}
	return status, nil
}

// VPNCheck returns if the helpers are installed and up to date and if polkit is running
func (b *Bitmask) VPNCheck() (helpers bool, privilege bool, err error) {
	return b.launch.Check()
}

func (b *Bitmask) GetLocationQualityMap(transport string) map[string]float64 {
	return b.api.GetLocationQualityMap(transport)
}

func (b *Bitmask) GetLocationLabels(transport string) map[string][]string {
	return b.api.GetLocationLabels(transport)
}

// UseGateway selects a gateway, by label, as the default gateway
func (b *Bitmask) UseGateway(label string) {
	b.api.SetManualGateway(label)
}

// UseAutomaticGateway sets the gateway to be selected automatically
// best gateway will be used
func (b *Bitmask) UseAutomaticGateway() {
	b.api.SetAutomaticGateway()
}

// SetTransport selects an obfuscation transport to use
func (b *Bitmask) SetTransport(t string) error {
	if t != "openvpn" && t != "obfs4" {
		return fmt.Errorf("Transport %s not implemented", t)
	}
	log.Info().
		Str("transport", t).
		Msg("Setting transport")
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
	return filepath.Join(b.tempdir, "openvpn.pem")
}

func (b *Bitmask) getTempCaCertPath() string {
	return filepath.Join(b.tempdir, "cacert.pem")
}

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
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/launcher"
	"0xacab.org/leap/bitmask-vpn/pkg/motd"
	"0xacab.org/leap/bitmask-vpn/pkg/snowflake"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/management"
	obfsvpn "0xacab.org/leap/obfsvpn/client"
)

type Bitmask struct {
	api              apiInterface           // handles backend API communication, implemented in bonafide (v3) or menshen (v5)
	onGateway        bonafide.Gateway       // gateway we are connected
	ptGateway        bonafide.Gateway       // public transport gateway we are connected with
	launch           *launcher.Launcher     // launcher manages the firewall and starts/stops openvpn
	canUpgrade       bool                   // is there an update available?
	motd             []motd.Message         // cached message of the day (ony fetched once during startup)
	statusCh         chan string            // channel used to get current OpenVPN state (remote ip, connection state, ...)
	managementClient *management.MgmtClient // used to speak with our own management backend (OpenVPN process connects to it)
	transport        string                 // used transport, e.g. OpenVPN (plain) or obfuscated protocols (obfs4)
	openvpnArgs      []string               // arguments used for invoking the OpenVPN process
	useUDP           bool                   // should we use UDP?
	obfsvpnProxy     *obfsvpn.Client        // handles OpenVPN obfuscation, e.g. starts/stops obfs4 bridge
	useSnowflake     bool                   // should we use Snowflake?
	provider         string                 // currently not used, get fixed if we get to the provider agnostic client
	tempdir          string                 // random base temp dir. Used for OpenVPN CA (cacert.pem),
	// client certificate (openvpn.pem, holds key and certificate) and management communication
	// authentication (key file prefixed with leap-vpn-...). Directory gets deleted during teardown
	certPemPath string // path of OpenVPN client certificate. Normally this is $tempdir/openvpn.pem,
	// but it also can be $config/$provider.pem (if snowflake is used or supplied out-of-band in a censored network)
}

// Init the connection to bitmask
func Init() (*Bitmask, error) {
	tempdir, err := ioutil.TempDir("", "leap-")
	if err != nil {
		return nil, err
	}

	api := bonafide.New()
	launch, err := launcher.NewLauncher()
	if err != nil {
		return nil, err
	}

	b := Bitmask{
		tempdir:          tempdir,
		onGateway:        bonafide.Gateway{},
		ptGateway:        bonafide.Gateway{},
		statusCh:         make(chan string, 10),
		managementClient: nil,
		api:              api,
		launch:           launch,
		transport:        "",
		obfsvpnProxy:     nil,
		certPemPath:      "",
		openvpnArgs:      []string{},
		useUDP:           false,
		useSnowflake:     false,
		canUpgrade:       isUpgradeAvailable(),
		motd:             motd.FetchLatest(),
		provider:         "",
	}

	// FIXME multiprovider: need to pass provider name early on
	// XXX we want to block on these, but they can timeout if we're blocked.
	b.checkForMOTD()
	err = b.launch.FirewallStop()
	if err != nil {
		log.Printf("Could not stop firewall: %v", err)
	}
	/*
		TODO -- we still want to do this, since it resets the fw/vpn if running
		from a previous one, but first we need to complete all the
		system/helper checks that we can do. otherwise this times out with an
		error that's captured badly as of today.

			err = b.StopVPN()
			if err != nil {
				return nil, err
			}
	*/

	err = ioutil.WriteFile(b.getTempCaCertPath(), config.CaCert, 0600)
	go b.fetchGateways()
	go b.openvpnManagement()

	return &b, err
}

func (b *Bitmask) SetProvider(p string) {
	b.provider = p
}

func (b *Bitmask) checkForMOTD() {
	b.motd = motd.FetchLatest()
}

// GetStatusCh returns a channel that will recieve VPN status changes
func (b *Bitmask) GetStatusCh() <-chan string {
	return b.statusCh
}

func (b *Bitmask) GetSnowflakeCh() <-chan *snowflake.StatusEvent {
	return b.api.GetSnowflakeCh()
}

// Close the connection to bitmask, and does cleanup of temporal files
func (b *Bitmask) Close() {
	log.Info().Msg("Close: cleanup and vpn shutdown...")
	err := b.StopVPN()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("There was an error stopping the vpn")
	}
	time.Sleep(500 * time.Millisecond)
	err = b.launch.Close()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("There was an error closing the launcher")
	}
	time.Sleep(1 * time.Second)
	err = os.RemoveAll(b.tempdir)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("There was an error removing temp dir")
	}
}

// Version gets the bitmask version string
func (b *Bitmask) Version() (string, error) {
	return "", nil
}

func (b *Bitmask) NeedsCredentials() bool {
	return b.api.NeedsCredentials()
}

func (b *Bitmask) DoLogin(username, password string) (bool, error) {
	return b.api.DoLogin(username, password)
}

func (b *Bitmask) UseUDP(udp bool) {
	b.useUDP = udp
}

func (b *Bitmask) UseSnowflake(s bool) error {
	b.useSnowflake = s
	return nil
}

func (b *Bitmask) OffersUDP() bool {
	return b.api.IsUDPAvailable()
}

func (b *Bitmask) GetMotd() string {
	bytes, err := json.Marshal(b.motd)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("error marshalling motd")
	}
	return string(bytes)
}

func (b *Bitmask) CanUpgrade() bool {
	return b.canUpgrade
}

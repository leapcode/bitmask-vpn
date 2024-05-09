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

package legacy

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
	"0xacab.org/leap/bitmask-vpn/pkg/launcher"
	"0xacab.org/leap/bitmask-vpn/pkg/motd"
	"0xacab.org/leap/bitmask-vpn/pkg/snowflake"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/management"
	obfsvpn "0xacab.org/leap/obfsvpn/client"
)

// Bitmask3 holds the bitmask client data
// Bitmask3 implements Bitmask interface (api version v3)
type Bitmask3 struct {
	tempdir          string
	onGateway        bonafide.Gateway
	ptGateway        bonafide.Gateway
	statusCh         chan string
	managementClient *management.MgmtClient
	bonafide         *bonafide.Bonafide
	launch           *launcher.Launcher
	transport        string
	obfsvpnProxy     *obfsvpn.Client
	certPemPath      string
	openvpnArgs      []string
	udp              bool
	snowflake        bool
	offersUdp        bool
	failed           bool
	canUpgrade       bool
	motd             []motd.Message
	provider         string
}

// Init the connection to bitmask
func Init() (*Bitmask3, error) {
	statusCh := make(chan string, 10)
	tempdir, err := ioutil.TempDir("", "leap-")
	if err != nil {
		return nil, err
	}
	snowCh := make(chan *snowflake.StatusEvent, 20)
	bf := bonafide.New()
	bf.SnowflakeCh = snowCh
	launch, err := launcher.NewLauncher()
	if err != nil {
		return nil, err
	}

	b := Bitmask3{
		tempdir,
		bonafide.Gateway{},
		bonafide.Gateway{}, statusCh, nil, bf, launch,
		"", nil, "", []string{},
		false, false, false, false, false,
		[]motd.Message{}, ""}
	// FIXME multiprovider: need to pass provider name early on
	// XXX we want to block on these, but they can timeout if we're blocked.
	b.checkForUpgrades()
	b.checkForMOTD()
	b.launch.FirewallStop()
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

func (b *Bitmask3) SetProvider(p string) {
	b.provider = p
}

func (b *Bitmask3) checkForUpgrades() {

	// SNAPS have their own way of upgrading. We probably should also try to detect
	// if we've been installed via another package manager.
	// For now, it's maybe a good idea to disable the UI check in linux, and be
	// way more strict in windows/osx.
	if os.Getenv("SNAP") != "" {
		return
	}
	b.canUpgrade = version.CanUpgrade()
}

func (b *Bitmask3) checkForMOTD() {
	b.motd = motd.FetchLatest()
}

// GetStatusCh returns a channel that will recieve VPN status changes
func (b *Bitmask3) GetStatusCh() <-chan string {
	return b.statusCh
}

func (b *Bitmask3) GetSnowflakeCh() <-chan *snowflake.StatusEvent {
	return b.bonafide.SnowflakeCh
}

// Close the connection to bitmask, and does cleanup of temporal files
func (b *Bitmask3) Close() {
	log.Info().Msg("Close: cleanup and vpn shutdown...")
	b.StopVPN()
	time.Sleep(500 * time.Millisecond)
	err := b.launch.Close()
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
func (b *Bitmask3) Version() (string, error) {
	return "", nil
}

func (b *Bitmask3) NeedsCredentials() bool {
	return b.bonafide.NeedsCredentials()
}

func (b *Bitmask3) DoLogin(username, password string) (bool, error) {
	return b.bonafide.DoLogin(username, password)
}

func (b *Bitmask3) UseUDP(udp bool) {
	b.udp = udp
}

func (b *Bitmask3) UseSnowflake(s bool) error {
	b.snowflake = s
	return nil
}

func (b *Bitmask3) OffersUDP() bool {
	return b.bonafide.IsUDPAvailable()
}

func (b *Bitmask3) GetMotd() string {
	bytes, err := json.Marshal(b.motd)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("error marshalling motd")
	}
	return string(bytes)
}

func (b *Bitmask3) CanUpgrade() bool {
	return b.canUpgrade
}

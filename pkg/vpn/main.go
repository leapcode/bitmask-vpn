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
	"log"
	"os"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/config/version"
	"0xacab.org/leap/bitmask-vpn/pkg/motd"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"0xacab.org/leap/shapeshifter"
	"github.com/apparentlymart/go-openvpn-mgmt/openvpn"
)

// Bitmask holds the bitmask client data
type Bitmask struct {
	tempdir          string
	onGateway        bonafide.Gateway
	ptGateway        bonafide.Gateway
	statusCh         chan string
	managementClient *openvpn.MgmtClient
	bonafide         *bonafide.Bonafide
	launch           *launcher
	transport        string
	shapes           *shapeshifter.ShapeShifter
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
func Init() (*Bitmask, error) {
	statusCh := make(chan string, 10)
	tempdir, err := ioutil.TempDir("", "leap-")
	if err != nil {
		return nil, err
	}
	bf := bonafide.New()
	launch, err := newLauncher()
	if err != nil {
		return nil, err
	}

	b := Bitmask{
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
	b.launch.firewallStop()
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

func (b *Bitmask) checkForUpgrades() {

	// SNAPS have their own way of upgrading. We probably should also try to detect
	// if we've been installed via another package manager.
	// For now, it's maybe a good idea to disable the UI check in linux, and be
	// way more strict in windows/osx.
	if os.Getenv("SNAP") != "" {
		return
	}
	b.canUpgrade = version.CanUpgrade()
}

func (b *Bitmask) checkForMOTD() {
	b.motd = motd.FetchLatest()
}

// GetStatusCh returns a channel that will recieve VPN status changes
func (b *Bitmask) GetStatusCh() <-chan string {
	return b.statusCh
}

// Close the connection to bitmask, and does cleanup of temporal files
func (b *Bitmask) Close() {
	log.Printf("Close: cleanup and vpn shutdown...")
	b.StopVPN()
	time.Sleep(500 * time.Millisecond)
	err := b.launch.close()
	if err != nil {
		log.Printf("There was an error closing the launcher: %v", err)
	}
	err = os.RemoveAll(b.tempdir)
	if err != nil {
		log.Printf("There was an error removing temp dir: %v", err)
	}
}

// Version gets the bitmask version string
func (b *Bitmask) Version() (string, error) {
	return "", nil
}

func (b *Bitmask) NeedsCredentials() bool {
	return b.bonafide.NeedsCredentials()
}

func (b *Bitmask) DoLogin(username, password string) (bool, error) {
	return b.bonafide.DoLogin(username, password)
}

func (b *Bitmask) UseUDP(udp bool) error {
	b.udp = udp
	return nil
}

func (b *Bitmask) UseSnowflake(s bool) error {
	b.snowflake = s
	return nil
}

func (b *Bitmask) OffersUDP() bool {
	return b.bonafide.IsUDPAvailable()
}

func (b *Bitmask) GetMotd() string {
	bytes, err := json.Marshal(b.motd)
	if err != nil {
		log.Println("WARN error marshalling motd")
	}
	return string(bytes)
}

func (b *Bitmask) CanUpgrade() bool {
	return b.canUpgrade
}

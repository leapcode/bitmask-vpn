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
	"io/ioutil"
	"log"
	"os"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
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
	b := Bitmask{tempdir, bonafide.Gateway{}, bonafide.Gateway{}, statusCh, nil, bf, launch, "", nil, "", []string{}}

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

	err = ioutil.WriteFile(b.getCaCertPath(), config.CaCert, 0600)

	go b.openvpnManagement()
	return &b, err
}

// GetStatusCh returns a channel that will recieve VPN status changes
func (b *Bitmask) GetStatusCh() <-chan string {
	return b.statusCh
}

// Close the connection to bitmask, and does cleanup of temporal files
func (b *Bitmask) Close() {
	log.Printf("Close: cleanup and vpn shutdown...")
	b.StopVPN()
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

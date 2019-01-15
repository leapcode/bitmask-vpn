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

package standalone

import (
	"io/ioutil"
	"log"
	"os"

	"0xacab.org/leap/bitmask-systray/pkg/config"
	"github.com/apparentlymart/go-openvpn-mgmt/openvpn"
)

// Bitmask holds the bitmask client data
type Bitmask struct {
	tempdir          string
	statusCh         chan string
	managementClient *openvpn.MgmtClient
	bonafide         *bonafide
	launch           *launcher
}

// Init the connection to bitmask
func Init() (*Bitmask, error) {
	statusCh := make(chan string, 10)
	tempdir, err := ioutil.TempDir("", "leap-")
	if err != nil {
		return nil, err
	}
	bonafide := newBonafide()
	launch, err := newLauncher()
	if err != nil {
		return nil, err
	}
	b := Bitmask{tempdir, statusCh, nil, bonafide, launch}

	err = b.StopVPN()
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(b.getCaCertPath(), config.CaCert, 0600)

	go b.openvpnManagement()
	return &b, err
}

// GetStatusCh returns a channel that will recieve VPN status changes
func (b *Bitmask) GetStatusCh() <-chan string {
	return b.statusCh
}

// Close the connection to bitmask
func (b *Bitmask) Close() {
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

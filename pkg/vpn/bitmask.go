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
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-core/pkg/storage"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/launcher"
	"0xacab.org/leap/bitmask-vpn/pkg/motd"
	"0xacab.org/leap/bitmask-vpn/pkg/snowflake"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/management"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/menshen"
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
	statusCloseCh    chan int               // chnanel used to close the status fetch loop to update GUI
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
	tempdir, err := os.MkdirTemp("", "leap-")
	if err != nil {
		return nil, err
	}

	var api apiInterface
	if os.Getenv("API_VERSION") == "5" {
		config.ProviderConfig.ApiVersion = 5
		log.Debug().Msg("Enforcing API v5 by env variable")
	}
	log.Debug().
		Int("apiVersion", config.ProviderConfig.ApiVersion).
		Msg("Using specific API backend version")

	// here check instead in the api_versions array (if exists)
	// and then decide if api v5 should be used or v3
	if config.ProviderConfig.ApiVersion == 5 {
		api, err = menshen.New()
		if err != nil {
			return nil, err
		}
	} else if config.ProviderConfig.ApiVersion == 3 {
		api = bonafide.New()
	} else {
		log.Warn().
			Int("apiVersion", config.ProviderConfig.ApiVersion).
			Msg("ApiVersion of provider was not set correctly. Version 3 and 5 is supported. Using v3 for backwards compatiblity")
		api = bonafide.New()
	}

	launch, err := launcher.NewLauncher()
	if err != nil {
		return nil, err
	}

	b := Bitmask{
		tempdir:          tempdir,
		onGateway:        bonafide.Gateway{},
		ptGateway:        bonafide.Gateway{},
		statusCh:         make(chan string, 10),
		statusCloseCh:    make(chan int),
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

	err = os.WriteFile(b.getTempCaCertPath(), config.ProviderConfig.CaCert, 0600)
	if err != nil {
		return nil, err
	}

	log.Debug().
		Str("caCertPath", b.getTempCaCertPath()).
		Msg("Sucessfully wrote OpenVPN CA certificate (hardcoded in the binary, not coming from API)")

	if err := b.launch.FirewallStop(); err != nil {
		log.Warn().
			Err(err).
			Msg("Could not stop firewall")
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

	if config.ProviderConfig.ApiVersion == 5 && len(config.ProviderConfig.STUNServers) != 0 {
		/*
			Geolocation lookup should be done only once during startup. Changing the country
			code during runtime is not supported. The VPN must be turn off for the lookup.
			If the lookup succeeds, we save it in the config file and use it as fallback
			the next time.
		*/
		err := b.api.DoGeolocationLookup()
		if err != nil {
			log.Warn().
				Str("err", err.Error()).
				Msgf("Could not do geolocation lookup")
		}
	}

	go b.fetchGateways()
	go b.initOpenVPNManagementHandler()

	return &b, err
}

func (b *Bitmask) SetProvider(p string) {
	b.provider = p
}

// GetStatusCh returns a channel that will recieve VPN status changes
func (b *Bitmask) GetStatusCh() <-chan string {
	return b.statusCh
}

func (b *Bitmask) GetStatusCloseCh() chan int {
	return b.statusCloseCh
}

func (b *Bitmask) GetSnowflakeCh() <-chan *snowflake.StatusEvent {
	return b.api.GetSnowflakeCh()
}

// Close the connection to bitmask, and does cleanup of temporal files
func (b *Bitmask) Close() {

	log.Info().Msg("Close: cleanup and vpn shutdown...")
	defer config.CloseLogger()

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

	store, err := storage.GetStorage()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not get bitmask-core storage to close it")
		return
	}
	store.Close()
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

func (b *Bitmask) OffersObfs4() bool {
	return b.api.SupportsObfs4()
}

func (b *Bitmask) OffersQUIC() bool {
	return b.api.SupportsQUIC()
}

func (b *Bitmask) OffersKCP() bool {
	return b.api.SupportsKCP()
}

func (b *Bitmask) OffersHopping() bool {
	return b.api.SupportsHopping()
}

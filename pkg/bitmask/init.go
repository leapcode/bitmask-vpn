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

package bitmask

import (
	"os"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn"
	"github.com/rs/zerolog/log"
)

type ProviderInfo struct {
	Provider string
	AppName  string
}

type ProviderOpts struct {
	Provider             string   `json:"name"`
	AppName              string   `json:"applicationName"`
	BinaryName           string   `json:"binaryName"`
	Auth                 string   `json:"auth"`
	AuthEmptyPass        bool     `json:"authEmptyPass"`
	ProviderURL          string   `json:"providerURL"`
	DonateURL            string   `json:"donateURL"`
	ApiURL               string   `json:"apiURL"`
	TosURL               string   `json:"tosURL"`
	HelpURL              string   `json:"helpURL"`
	GeolocationURL       string   `json:"geolocationAPI"`
	AskForDonations      bool     `json:"askForDonations"`
	CaCert               string   `json:"caCertString"`
	ApiVersion           int      `json:"apiVersion"`
	STUNServers          []string `json:"STUNServers"`
	CountryCodeLookupURL string   `json:"countryCodeLookupURL"`
}

func ConfigureProvider(opts *ProviderOpts) {
	log.Info().Msg("configuring provider")
	config.ProviderConfig.Provider = opts.Provider
	config.ProviderConfig.ApplicationName = opts.AppName
	config.ProviderConfig.BinaryName = opts.BinaryName
	config.ProviderConfig.Auth = opts.Auth
	config.ProviderConfig.GeolocationAPI = opts.GeolocationURL
	config.ProviderConfig.APIURL = opts.ApiURL
	config.ProviderConfig.CaCert = []byte(opts.CaCert)
	config.ProviderConfig.ApiVersion = opts.ApiVersion
	config.ProviderConfig.STUNServers = opts.STUNServers
	config.ProviderConfig.CountryCodeLookupURL = opts.CountryCodeLookupURL
}

func InitializeBitmask(conf *config.Config) (Bitmask, error) {
	log.Trace().Msg("Initializing bitmask")
	if conf.SkipLaunch {
		log.Info().Msg("Not autostarting OpenVPN")
	}
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		err = os.MkdirAll(config.Path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	b, err := vpn.Init()
	if err != nil {
		return nil, err
	}
	b.SetProvider(config.ProviderConfig.Provider)

	err = setTransport(b, conf)
	if err != nil {
		return nil, err
	}

	setUDP(b, conf)

	if !conf.SkipLaunch {
		err := maybeStartVPN(b, conf)
		if err != nil {
			// we don't want this error to avoid initialization of
			// the bitmask object. If we cannot autostart it's not
			// so terrible.
			log.Warn().
				Err(err).
				Msg("Could not start OpenVPN (maybeStartVPN)")
		}
	}
	return b, nil
}

func setTransport(b Bitmask, conf *config.Config) error {
	if config.ProviderConfig.Provider == "riseup" {
		log.Info().Msg("Using transport openvpn")
		err := b.SetTransport("openvpn")
		if err != nil {
			return err
		}
		return nil
	}
	if conf.Obfs4 {
		log.Info().Msg("Using transport obfs4")
		err := b.SetTransport("obfs4")
		if err != nil {
			return err
		}
		return nil
	}
	if conf.KCP {
		log.Info().Msg("Using transport kcp")
		err := b.SetTransport("kcp")
		if err != nil {
			return err
		}
	}
	return nil
}

func setUDP(b Bitmask, conf *config.Config) {
	if conf.UDP {
		log.Info().Msg("Enabled UDP")
		b.UseUDP(true)
	} else {
		log.Warn().Msg("UDP not enabled in config")
	}
}

func maybeStartVPN(b Bitmask, conf *config.Config) error {
	if !conf.StartVPN {
		return nil
	}

	if b.CanStartVPN() {
		log.Info().Msg("Starting OpenVPN")
		err := b.StartVPN(config.ProviderConfig.Provider)
		if err != nil {
			return err
		}
		err = conf.SetUserStoppedVPN(false)
		return err
	}
	return nil
}

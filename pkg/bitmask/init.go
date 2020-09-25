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
	"log"
	"os"
	"path"
	"strconv"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn"
)

type ProviderInfo struct {
	Provider string
	AppName  string
}

type ProviderOpts struct {
	Provider        string `json:"name"`
	AppName         string `json:"applicationName"`
	BinaryName      string `json:"binaryName"`
	Auth            string `json:"auth"`
	AuthEmptyPass   string `json:"authEmptyPass"`
	ProviderURL     string `json:"providerURL"`
	DonateURL       string `json:"donateURL"`
	ApiURL          string `json:"apiURL"`
	TosURL          string `json:"tosURL"`
	HelpURL         string `json:"helpURL"`
	GeolocationURL  string `json:"geolocationAPI"`
	AskForDonations string `json:"askForDonations"`
	CaCert          string `json:"caCertString"`
}

func GetConfiguredProvider() *ProviderInfo {
	provider := config.Provider
	appName := config.ApplicationName
	return &ProviderInfo{provider, appName}
}

func ConfigureProvider(opts *ProviderOpts) {
	config.Provider = opts.ProviderURL
	config.ProviderName = opts.Provider
	config.ApplicationName = opts.AppName
	config.BinaryName = opts.BinaryName
	config.Auth = opts.Auth

	config.DonateURL = opts.DonateURL
	config.HelpURL = opts.HelpURL
	config.TosURL = opts.TosURL
	config.APIURL = opts.ApiURL
	config.GeolocationAPI = opts.GeolocationURL
	config.CaCert = []byte(opts.CaCert)

	wantsDonations, err := strconv.ParseBool(opts.AskForDonations)
	if err == nil {
		config.AskForDonations = wantsDonations
	}

	emptyPass, err := strconv.ParseBool(opts.AuthEmptyPass)
	if err == nil {
		config.AuthEmptyPass = emptyPass
		log.Println("DEBUG: provider allows empty pass:", emptyPass)
	}
}

func InitializeLogger() {
	_, err := config.ConfigureLogger(path.Join(config.LogPath))
	if err != nil {
		log.Println("Can't configure logger: ", err)
	}
}

func initBitmask() (Bitmask, error) {
	b, err := vpn.Init()
	if err != nil {
		log.Printf("An error ocurred starting bitmask: %v", err)
	}
	return b, err
}

func InitializeBitmask(conf *config.Config) (Bitmask, error) {
	if conf.SkipLaunch {
		log.Println("Initializing bitmask, but not launching it...")
	}
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		os.MkdirAll(config.Path, os.ModePerm)
	}

	b, err := initBitmask()
	if err != nil {
		return nil, err
	}

	err = setTransport(b, conf)
	if err != nil {
		return nil, err
	}

	if !conf.SkipLaunch {
		err := maybeStartVPN(b, conf)
		if err != nil {
			log.Println("Error starting VPN: ", err)
			return nil, err
		}
	}
	return b, nil
}

func setTransport(b Bitmask, conf *config.Config) error {
	if conf.Obfs4 {
		log.Printf("Use transport Obfs4")
		err := b.UseTransport("obfs4")
		if err != nil {
			log.Printf("Error setting transport: %v", err)
			return err
		}
	}
	return nil
}

func maybeStartVPN(b Bitmask, conf *config.Config) error {
	if !conf.StartVPN {
		return nil
	}

	if b.CanStartVPN() {
		err := b.StartVPN(config.Provider)
		conf.SetUserStoppedVPN(false)
		return err
	}
	return nil
}

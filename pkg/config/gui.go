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

package config

import (
	"encoding/json"
	"os"
	"path"
	"time"
)

const (
	oneDay   = time.Hour * 24
	oneWeek  = oneDay * 7
	oneMonth = oneDay * 30
)

var (
	configPath = path.Join(Path, "systray.json")
	LogPath    = path.Join(Path, "systray.log")
)

// Config holds the configuration of the systray
type Config struct {
	file struct {
		LastReminded      time.Time
		Donated           time.Time
		Obfs4             bool
		UserStoppedVPN    bool
		DisableAustostart bool
	}
	Obfs4             bool
	DisableAustostart bool
	StartVPN          bool
	SkipLaunch        bool
}

// ParseConfig reads the configuration from the configuration file
func ParseConfig() *Config {
	var conf Config

	f, err := os.Open(configPath)
	if err != nil {
		conf.save()
	} else {
		defer f.Close()
		dec := json.NewDecoder(f)
		err = dec.Decode(&conf.file)
	}

	conf.Obfs4 = conf.file.Obfs4
	conf.DisableAustostart = conf.file.DisableAustostart
	conf.StartVPN = !conf.file.UserStoppedVPN
	return &conf
}

func (c *Config) SetUserStoppedVPN(vpnStopped bool) error {
	c.file.UserStoppedVPN = vpnStopped
	return c.save()
}

func (c *Config) NeedsDonationReminder() bool {
	return !c.hasDonated() && c.file.LastReminded.Add(oneWeek).Before(time.Now())
}

func (c *Config) hasDonated() bool {
	return c.file.Donated.Add(oneMonth).After(time.Now())
}

func (c *Config) SetLastReminded() error {
	c.file.LastReminded = time.Now()
	return c.save()
}

func (c *Config) SetDonated() error {
	c.file.Donated = time.Now()
	return c.save()
}

func (c *Config) save() error {
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c.file)
}

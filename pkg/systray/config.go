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

package systray

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"0xacab.org/leap/bitmask-systray/pkg/config"
	"golang.org/x/text/message"
)

const (
	oneDay   = time.Hour * 24
	oneMonth = oneDay * 30
)

var (
	configPath = path.Join(config.Path, "systray.json")
)

// SystrayConfig holds the configuration of the systray
type SystrayConfig struct {
	LastNotification time.Time
	Donated          time.Time
	SelectGateway    bool
	UserStoppedVPN   bool
	Provider         string           `json:"-"`
	ApplicationName  string           `json:"-"`
	Version          string           `json:"-"`
	Printer          *message.Printer `json:"-"`
}

// ParseConfig reads the configuration from the configuration file
func ParseConfig() *SystrayConfig {
	var conf SystrayConfig

	f, err := os.Open(configPath)
	if err != nil {
		conf.save()
		return &conf
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&conf)
	return &conf
}

func (c *SystrayConfig) setUserStoppedVPN(vpnStopped bool) error {
	c.UserStoppedVPN = vpnStopped
	return c.save()
}

func (c *SystrayConfig) hasDonated() bool {
	return c.Donated.Add(oneMonth).After(time.Now())
}

func (c *SystrayConfig) needsNotification() bool {
	return !c.hasDonated() && c.LastNotification.Add(oneDay).Before(time.Now())
}

func (c *SystrayConfig) setNotification() error {
	c.LastNotification = time.Now()
	return c.save()
}

func (c *SystrayConfig) setDonated() error {
	c.Donated = time.Now()
	return c.save()
}

func (c *SystrayConfig) save() error {
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(c)
}

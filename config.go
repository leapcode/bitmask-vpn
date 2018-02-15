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

package main

import (
	"encoding/json"
	"flag"
	"os"
	"path"
	"time"

	"0xacab.org/leap/bitmask-systray/bitmask"
)

const (
	oneDay   = time.Hour * 24
	oneMonth = oneDay * 30
)

var (
	configPath = path.Join(bitmask.ConfigPath, "systray.json")
)

type systrayConfig struct {
	LastNotification time.Time
	Donated          time.Time
	SelectWateway    bool
}

func parseConfig() (*systrayConfig, error) {
	var conf systrayConfig

	f, err := os.Open(configPath)
	if os.IsNotExist(err) {
		return &conf, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&conf)
	conf.parseFlags()
	return &conf, err
}

func (c *systrayConfig) parseFlags() {
	flag.BoolVar(&c.SelectWateway, "select-gateway", false, "Enable gateway selection")
	flag.Parse()
}

func (c *systrayConfig) hasDonated() bool {
	return c.Donated.Add(oneMonth).After(time.Now())
}

func (c *systrayConfig) needsNotification() bool {
	return !c.hasDonated() && c.LastNotification.Add(oneDay).Before(time.Now())
}

func (c *systrayConfig) setNotification() error {
	c.LastNotification = time.Now()
	return c.save()
}

func (c *systrayConfig) setDonated() error {
	c.Donated = time.Now()
	return c.save()
}

func (c *systrayConfig) save() error {
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(c)
}

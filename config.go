package main

import (
	"encoding/json"
	"log"
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
	return &conf, err
}

func (c *systrayConfig) hasDonated() bool {
	log.Println("has donated ", c.Donated.Add(oneMonth))
	return c.Donated.Add(oneMonth).After(time.Now())
}

func (c *systrayConfig) needsNotification() bool {
	log.Println("needs ", c.LastNotification.Add(oneDay))
	log.Println(!c.hasDonated() && c.LastNotification.Add(oneDay).Before(time.Now()))
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

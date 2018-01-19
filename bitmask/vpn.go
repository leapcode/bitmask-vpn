package bitmask

import (
	"log"
	"time"
)

// StartVPN for provider
func (b *Bitmask) StartVPN(provider string) error {
	_, err := b.send("vpn", "start", provider)
	return err
}

// StopVPN or cancel
func (b *Bitmask) StopVPN() error {
	_, err := b.send("vpn", "stop")
	return err
}

// GetStatus returns the VPN status
func (b *Bitmask) GetStatus() (string, error) {
	res, err := b.send("vpn", "status")
	if err != nil {
		return "", err
	}
	return res["status"].(string), nil
}

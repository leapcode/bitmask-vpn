package bitmask

import (
	"errors"
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

// ListGateways return the names of the gateways
func (b *Bitmask) ListGateways(provider string) ([]string, error) {
	res, err := b.send("vpn", "list")
	if err != nil {
		return nil, err
	}

	names := []string{}
	locations, ok := res[provider].([]interface{})
	if !ok {
		return nil, errors.New("Can't read the locations for provider " + provider)
	}
	for i := range locations {
		loc := locations[i].(map[string]interface{})
		names = append(names, loc["name"].(string))
	}
	return names, nil
}

// UseGateway selects name as the default gateway
func (b *Bitmask) UseGateway(name string) error {
	_, err := b.send("vpn", "locations", name)
	return err
}

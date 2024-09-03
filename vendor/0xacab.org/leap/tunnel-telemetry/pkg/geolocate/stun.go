package geolocate

import (
	"encoding/json"

	"github.com/pion/stun"
)

var (
	// useful: https://github.com/pradt2/always-online-stun
	stunServers = []string{
		"stun.ekiga.net:3478",
		"stun.syncthing.net:3478",
		"stun.nextcloud.com:443",
		"relay.webwormhole.io:3478",
		"stun4.l.google.com:3478",
		"stun1.l.google.com:3478",
		"stun2.l.google.com:3478",
		"stun3.l.google.com:3478",
		"stun.cloudflare.com:3478",
		// cn
		"stun.xten.com:3478",
		"stun.miwifi.com:3478",
		"stun.chat.bilibili.com:3478",
		// hardcoded ips
		"209.105.241.31:3478",
		"51.68.45.75:3478",
		"37.139.120.14:3478",
		"3.132.228.249:3478",
		"198.27.70.99:3478",
		"5.9.87.18:3478",
		"193.22.17.97:3478",
	}
)

// FetchIPFromSTUNCall tries to get our public IP using the passed
// stun uri. It returns an error of the operation does not succeed.
func FetchIPFromSTUNCall(uri string) (string, error) {
	u, err := stun.ParseURI("stun:" + uri)
	if err != nil {
		return "", err
	}

	// Create a "connection" to STUN server.
	c, err := stun.DialURI(u, &stun.DialConfig{})
	if err != nil {
		return "", err
	}
	// Build binding request with random transaction id.
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	errch, ipch := make(chan error, 1), make(chan string, 1)

	// Send request to STUN server, waiting for response message.
	if err := c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			errch <- res.Error
			return
		}
		// Decode XOR-MAPPED-ADDRESS attribute from message.
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			errch <- err
			return
		}
		ipch <- xorAddr.IP.String()
	}); err != nil {
		return "", err
	}
	// TODO(ain): add timeout/ctx
	select {
	case err := <-errch:
		return "", err
	case ip := <-ipch:
		return ip, nil
	}
}

func getJSON(url string, target interface{}) error {
	r, err := defaultHTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

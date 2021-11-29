package snowflake

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/cretz/bine/tor"
)

const torrc = `UseBridges 1
DataDirectory datadir

ClientTransportPlugin snowflake exec /usr/local/bin/snowflake-client \
-url https://snowflake-broker.torproject.net.global.prod.fastly.net/ -front cdn.sstatic.net \
-ice stun:stun.voip.blackberry.com:3478,stun:stun.altar.com.pl:3478,stun:stun.antisip.com:3478,stun:stun.bluesip.net:3478,stun:stun.dus.net:3478,stun:stun.epygi.com:3478,stun:stun.sonetel.com:3478,stun:stun.sonetel.net:3478,stun:stun.stunprotocol.org:3478,stun:stun.uls.co.za:3478,stun:stun.voipgate.com:3478,stun:stun.voys.nl:3478 \
-max 3

Bridge snowflake 0.0.3.0:1`

func writeTorrc() string {
	f, err := ioutil.TempFile("", "torrc-snowflake-")
	if err != nil {
		log.Println(err)
	}
	f.Write([]byte(torrc))
	return f.Name()
}

func BootstrapWithSnowflakeProxies() error {
	rcfile := writeTorrc()
	conf := &tor.StartConf{DebugWriter: os.Stdout, TorrcFile: rcfile}

	fmt.Println("Starting Tor and fetching files to bootstrap VPN tunnel...")
	fmt.Println("")

	t, err := tor.Start(nil, conf)
	if err != nil {
		return err
	}
	defer t.Close()

	// Wait at most 5 minutes
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer dialCancel()
	dialer, err := t.Dialer(dialCtx, nil)
	if err != nil {
		return err
	}

	/*
		regClient := &http.Client{
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
			},
			Timeout: time.Minute * 5,
		}
	*/
	//fetchFile(regClient, "https://wtfismyip.com/json")

	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM(config.CaCert)

	apiClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
			DialContext: dialer.DialContext,
		},
		Timeout: time.Minute * 5,
	}

	// XXX parametrize these urls
	fetchFile(apiClient, "https://api.black.riseup.net/3/config/eip-service.json")
	fetchFile(apiClient, "https://api.black.riseup.net/3/cert")

	return nil
}

func fetchFile(client *http.Client, uri string) error {
	resp, err := client.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(c))
	return nil
}

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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/cretz/bine/tor"
)

// TODO
// [ ] fix snowflake-client binary
// [ ] find tor path

const torrc = `UseBridges 1
DataDirectory datadir

ClientTransportPlugin snowflake exec /usr/local/bin/snowflake-client -log /tmp/snowflake.log -url https://snowflake-broker.torproject.net.global.prod.fastly.net/ \
-front cdn.sstatic.net -ice stun:stun.voip.blackberry.com:3478,stun:stun.altar.com.pl:3478,stun:stun.antisip.com:3478,stun:stun.bluesip.net:3478,stun:stun.dus.net:3478,stun:stun.epygi.com:3478,stun:stun.sonetel.com:3478,stun:stun.sonetel.net:3478,stun:stun.stunprotocol.org:3478,stun:stun.uls.co.za:3478,stun:stun.voipgate.com:3478,stun:stun.voys.nl:3478 \
-max 5

Bridge snowflake 192.0.2.3:1

SocksPort auto`

type StatusEvent struct {
	Progress int
	Tag      string
}

type StatusLogger struct {
	ch chan *StatusEvent
}

func (e *StatusLogger) Write(p []byte) (n int, err error) {
	raw := strings.Split(string(p), ":")
	if len(raw) > 1 {
		l := raw[1]
		parts := strings.Split(string(l), " ")
		if len(parts) > 2 && parts[2] == "STATUS_CLIENT" {
			if parts[4] == "BOOTSTRAP" {
				if len(parts) > 6 {
					pr, _ := strconv.Atoi(parts[5][9:])
					event := &StatusEvent{Progress: pr, Tag: parts[6][4:]}
					go func() { e.ch <- event }()
				}
				fmt.Println()
			}
		}
	}
	return len(p), nil
}

func writeTorrc() string {
	f, err := ioutil.TempFile("", "torrc-snowflake-")
	if err != nil {
		log.Println(err)
	}
	f.Write([]byte(torrc))
	return f.Name()
}

// TODO pass provider api
func BootstrapWithSnowflakeProxies(provider string, api string, ch chan *StatusEvent) error {
	rcfile := writeTorrc()
	logger := &StatusLogger{ch}
	conf := &tor.StartConf{
		DebugWriter: logger,
		TorrcFile:   rcfile,
	}

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

	eipUri := "https://" + api + "/3/config/eip-service.json"
	eipFile := filepath.Join(config.Path, provider+"-eip.json")
	fetchFile(apiClient, eipUri, eipFile)

	certUri := "https://" + api + "/3/cert"
	certFile := filepath.Join(config.Path, provider+".pem")
	fetchFile(apiClient, certUri, certFile)

	return nil
}

func fetchFile(client *http.Client, uri string, file string) error {
	resp, err := client.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	if os.Getenv("DEBUG") == "1" {
		fmt.Println(string(c))
	}
	return ioutil.WriteFile(file, c, 0600)
}

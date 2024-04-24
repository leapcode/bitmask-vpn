package snowflake

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/cretz/bine/tor"
)

// TODO
// [ ] fix snowflake-client binary
// [ ] find tor path

const torrcOrig = `UseBridges 1
DataDirectory datadir

ClientTransportPlugin snowflake exec /usr/local/bin/snowflake-client -log /tmp/snowflake.log -url https://snowflake-broker.torproject.net.global.prod.fastly.net/ \
-front cdn.sstatic.net -ice stun:stun.voip.blackberry.com:3478,stun:stun.altar.com.pl:3478,stun:stun.antisip.com:3478,stun:stun.bluesip.net:3478,stun:stun.dus.net:3478,stun:stun.epygi.com:3478,stun:stun.sonetel.com:3478,stun:stun.sonetel.net:3478,stun:stun.stunprotocol.org:3478,stun:stun.uls.co.za:3478,stun:stun.voipgate.com:3478,stun:stun.voys.nl:3478 \
-max 5

Bridge snowflake 192.0.2.3:1

SocksPort auto`

const torrc = `UseBridges 1
DataDirectory datadir

ClientTransportPlugin snowflake exec /usr/local/bin/snowflake-client -log /tmp/snowflake.log -url https://snowflake-broker.azureedge.net/ \
-front ajax.aspnetcdn.com -ice stun:stun.l.google.com:19302 \
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
		log.Fatal().
			Err(err).
			Msg("Could not write tor config file")
	}
	f.Write([]byte(torrc))
	return f.Name()
}

// TODO pass provider api
func BootstrapWithSnowflakeProxies(provider string, ch chan *StatusEvent) error {
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

	apiClient := &http.Client{
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
		},
		Timeout: time.Minute * 5,
	}

	eipUri := config.APIURL + "/3/config/eip-service.json"
	eipFile := filepath.Join(config.Path, provider+"-eip.json")
	err = fetchFile(apiClient, eipUri, eipFile)
	if err != nil {
		return err
	}

	certUri := config.APIURL + "/3/cert"
	certFile := filepath.Join(config.Path, provider+".pem")
	err = fetchFile(apiClient, certUri, certFile)
	if err != nil {
		return err
	}
	return nil
}

func fetchFile(client *http.Client, uri string, file string) error {
	log.Debug().
		Str("uri", uri).
		Str("outFile", file).
		Msg("Fetching file over snowflake and saving to disk")
	resp, err := client.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Data received: %s", c)
	return ioutil.WriteFile(file, c, 0600)
}

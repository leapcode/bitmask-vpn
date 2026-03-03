package snowflake

import (
	"context"
	"io"
	"net/http"
	"os"
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
			}
		}
	}
	return len(p), nil
}

func writeTorrc() string {
	f, err := os.CreateTemp("", "torrc-snowflake-")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not write tor config file")
	}
	_, _ = f.Write([]byte(torrc))
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

	log.Info().Msg("Starting Tor and fetching files to bootstrap VPN tunnel...")

	t, err := tor.Start(context.TODO(), conf)
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

	EIPURI := config.ProviderConfig.APIURL + "/3/config/eip-service.json"
	EIPFile := filepath.Join(config.Path, provider+"-eip.json")
	err = fetchFile(apiClient, EIPURI, EIPFile)
	if err != nil {
		return err
	}

	certURI := config.ProviderConfig.APIURL + "/3/cert"
	certFile := filepath.Join(config.Path, provider+".pem")
	err = fetchFile(apiClient, certURI, certFile)
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
	return os.WriteFile(file, c, 0600)
}

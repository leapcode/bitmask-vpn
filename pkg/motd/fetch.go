package motd

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/rs/zerolog/log"
)

const riseupMOTD = "https://static.riseup.net/vpn/motd.json"

func FetchLatest() []Message {
	empty := []Message{}
	if os.Getenv("SKIP_MOTD") == "1" {
		log.Info().Msg("Skipping MOTD fetch")
		return empty
	}
	url := os.Getenv("MOTD_URL")
	if url == "" {
		switch config.Provider {
		case "riseup.net":
			url = riseupMOTD
		default:
			return empty
		}
	}

	log.Debug().
		Str("url", url).
		Msg("Fetching MOTD")

	b, err := fetchURL(url)
	if err != nil {
		log.Warn().Err(err).
			Str("url", "url").
			Msg("Could not fetch MOTD json")
		return empty
	}

	allMsg, err := getFromJSON(b)
	if err != nil {
		log.Warn().Err(err).
			Str("msg", string(b)).
			Msg("Could not json decode MOTD")
		return empty
	}
	valid := empty[:]
	if allMsg.Length() != 0 {
		log.Debug().
			Int("pendingMessages", allMsg.Length()).
			Msg("There are pending messages")
	}
	for _, msg := range allMsg.Messages {
		if msg.IsValid() {
			valid = append(valid, msg)
		}
	}
	return valid
}

func fetchURL(url string) ([]byte, error) {
	c := http.Client{Timeout: 30 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

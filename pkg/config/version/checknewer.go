package version

import (
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const verURI = "https://downloads.leap.se/RiseupVPN/"

// CanUpgrade returns true if there's a newer version string published on the server
// this needs to manually bump latest version for every platform in the
// downloads server.
// at the moment, we hardcode RiseupVPN in the path, assuming that all clients
// stay in sync.
func CanUpgrade() bool {
	if os.Getenv("SKIP_VERSION_CHECK") == "1" {
		log.Info().Msg("Not checking for upgrades")
		return false
	}
	uri := verURI
	switch runtime.GOOS {
	case "windows":
		uri += "windows"
	case "darwin":
		uri += "osx"
	case "linux":
		fallthrough
	default:
		uri += "linux"
	}
	uri += "/lastver"
	c := http.Client{Timeout: 30 * time.Second}
	resp, err := c.Get(uri)
	log.Info().
		Str("url", uri).
		Msg("Checking for updates")

	if err != nil {
		log.Warn().
			Err(err).
			Str("url", uri).
			Msg("Could not check if there are updates available")
		return false
	}
	defer resp.Body.Close()
	verStr, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not read http response")
		return false
	}
	r := strings.TrimSpace(string(verStr))
	if strings.Count(r, "\n") > 1 {
		log.Warn().
			Str("versionString", string(verStr)).
			Msg("Could not parse version string")
		return false
	}
	canUpgrade := versionOrdinal(r) > versionOrdinal(VERSION)
	log.Debug().
		Str("version", r).
		Msg("Remote version")
	log.Debug().
		Str("version", VERSION).
		Msg("Installed version")
	log.Info().
		Bool("updateAvailable", canUpgrade).
		Msg("Sucessfully checked if there is an update")
	return canUpgrade
}

// https://stackoverflow.com/a/18411978
func versionOrdinal(version string) string {
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			log.Warn().Msg("VersionOrdinal: invalid version")
			return string(vo)
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}

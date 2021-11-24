package version

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const verURI = "https://downloads.leap.se/RiseupVPN/"

// returns true if there's a newer version string published on the server
// this needs to manually bump latest version for every platform in the
// downloads server.
// at the moment, we hardcode RiseupVPN in the path, assuming that all clients
// stay in sync.
func CanUpgrade() bool {
	log.Println("Checking for updates...")
	uri := verURI
	switch runtime.GOOS {
	case "windows":
		uri += "windows"
	case "linux":
		uri += "linux"
	case "osx":
		uri += "osx"
	}
	uri += "/lastver"
	resp, err := http.Get(uri)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	verStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	r := strings.TrimSpace(string(verStr))
	if strings.Count(r, "\n") > 1 {
		log.Println("No remote version found at " + uri)
		return false
	}
	canUpgrade := versionOrdinal(r) > versionOrdinal(VERSION)
	if os.Getenv("DEBUG") == "1" {
		log.Println(">>> Remote version:  " + r)
		log.Println(">>> Current version: " + VERSION)
	}
	if canUpgrade {
		log.Println("There's a newer version available:", r)
	}
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
			log.Println("VersionOrdinal: invalid version")
			return string(vo)
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}

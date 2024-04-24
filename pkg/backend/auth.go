package backend

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

/* functions for local authentication of control endpoints */

const bitmaskToken = "bitmask-token"

func generateAuthToken() {
	if runtime.GOOS != "linux" {
		log.Warn().Msg("Authentication token only implemented in linux at the moment.")
		return
	}
	t := getRandomString()
	tokenPath := filepath.Join(os.TempDir(), bitmaskToken)
	err := ioutil.WriteFile(tokenPath, []byte(t), os.FileMode(int(0600)))
	if err != nil {
		log.Fatal().
			Err(err).
			Str("file", tokenPath).
			Msg("Could not write authentication token")
	}
}

func readAuthToken() string {
	if runtime.GOOS != "linux" {
		log.Warn().Msg("Authentication token only implemented in linux at the moment.")
		return ""
	}
	tokenPath := filepath.Join(os.TempDir(), bitmaskToken)
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("file", tokenPath).
			Msg("Could not read auth token from disk")
	}
	return string(token)
}

func getRandomString() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 40
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

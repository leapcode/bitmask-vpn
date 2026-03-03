//go:build darwin || linux
// +build darwin linux

package helper

import (
	"net"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func runServer(socketUID, socketGID int) {
	socketPath := filepath.Join("/tmp", helperSocket)
	if err := os.Remove(socketPath); err != nil {
		log.Warn().
			Err(err).
			Msg("unable to remove socket file or it doesn't exist")
	}
	unixListener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to create unix listener")
	}
	log.Info().
		Str("socketPath", socketPath).
		Msg("created listener")
	log.Info().
		Int("socket uid", socketUID).
		Int("socket gid", socketGID).
		Msg("changing socket ownership")

	if err = os.Chown(socketPath, socketUID, socketGID); err != nil {
		log.Fatal().
			Err(err).
			Msg("unable to change owner of socket file")
	}
	serveHTTP(unixListener)
}

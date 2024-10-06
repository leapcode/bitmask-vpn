//go:build darwin || linux
// +build darwin linux

package helper

import (
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"path/filepath"
)

func runServer(socketUid, socketGid int) {
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
		Int("socket uid", socketUid).
		Int("socket gid", socketGid).
		Msg("changing socket ownership")

	if err = os.Chown(socketPath, socketUid, socketGid); err != nil {
		log.Fatal().
			Err(err).
			Msg("unable to change owner of socket file")
	}
	serveHTTP(unixListener)
}

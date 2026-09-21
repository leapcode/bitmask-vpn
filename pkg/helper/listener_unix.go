//go:build darwin || linux
// +build darwin linux

package helper

import (
	"errors"
	"io/fs"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"
)

// socketDir returns the path to the sockets dir, inside the home of the user
// the helper serves (the helper itself runs as root).
func socketDir(socketUID int) (string, error) {
	u, err := user.LookupId(strconv.Itoa(socketUID))
	if err != nil {
		return "", err
	}
	return filepath.Join(u.HomeDir, ".config", "leap", "sockets"), nil
}

// cleanupListener wraps a unix listener, removing the socket file on Close.
type cleanupListener struct {
	net.Listener
	socketPath string
}

func (l cleanupListener) Close() error {
	if err := l.Listener.Close(); err != nil {
		return err
	}
	return os.Remove(l.socketPath)
}

func runServer(socketUID, socketGID int) {
	dir, err := socketDir(socketUID)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("unable to get socket dir")
	}
	if err := os.MkdirAll(dir, 0750); err != nil {
		log.Fatal().
			Err(err).
			Str("dir", dir).
			Msg("unable to create socket dir")
	}
	if err := os.Chown(dir, socketUID, socketGID); err != nil {
		log.Warn().
			Err(err).
			Msg("unable to change owner of socket dir")
	}

	socketPath := filepath.Join(dir, helperSocket)
	if err := os.Remove(socketPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Warn().
			Err(err).
			Msg("unable to remove stale socket file")
	}
	unixListener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal().
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
	serveHTTP(cleanupListener{unixListener, socketPath})
}

package snowflake

import (
	"errors"
	"os"
)

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func HasTor() bool {
	return exists("/usr/sbin/tor")
}

package storage

import (
	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-core/pkg/models"
)

func MaybeGetBridgeByName(name string) (models.Bridge, error) {
	var b models.Bridge
	var err error

	db, err := NewStorageWithDefaultDir()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot open storage")
	}
	defer db.Close()

	b, err = db.GetBridgeByName(name)
	if err != nil {
		return b, err
	}
	return b, nil
}

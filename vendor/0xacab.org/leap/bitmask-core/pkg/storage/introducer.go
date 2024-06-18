package storage

import (
	"strings"

	"github.com/rs/zerolog/log"
)

func MaybeGetIntroducerURLByName(introducerName string) string {
	if introducerName == "" {
		return ""
	}
	switch strings.HasPrefix(introducerName, "obfsvpnintro://") {
	// does not have prefix schema, let's treat it as a name in the internal storage
	case false:
		db, err := NewStorageWithDefaultDir()
		if err != nil {
			log.Fatal().Err(err).Msg("cannot open storage")
		}
		defer db.Close()

		_i, err := db.GetIntroducerByName(introducerName)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot get introducer by name")
		}
		// We got a valid (and unique) introducer from storage, so that
		// we can retrieve the URL value. This is assumed to have been validated
		// when introducing it, but I should be more paranoid here.
		// TODO(atanarjuat): be more paranoid and validate the URL/signature etc.
		return _i.URL
	default:
		return ""
	}
}

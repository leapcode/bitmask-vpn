// Copyright (C) 2018 LEAP
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ConfigureLogger() {
	os.MkdirAll(Path, 0750)

	runLogFile, _ := os.OpenFile(
		LogPath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02T15:04:05.999Z07:00",
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, runLogFile)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	envLogLevel := os.Getenv("LOG_LEVEL")
	if envLogLevel == "TRACE" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if envLogLevel == "DEBUG" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().
		Str("logFile", LogPath).
		Str("hint", "you can change the log level with env LOG_LEVEL=INFO|DEBUG|TRACE").
		Msg("Enabling logging")
}

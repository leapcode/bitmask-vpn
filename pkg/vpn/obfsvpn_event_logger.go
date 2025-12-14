package vpn

import "github.com/rs/zerolog/log"

type EventLogger struct {
	Events chan string
}

func (e *EventLogger) Log(state, message string) {
	log.Info().Msg(message)
	e.Events <- state
}

func (e *EventLogger) Error(message string) {
	log.Error().Msg(message)
	e.Events <- "ERROR"
}

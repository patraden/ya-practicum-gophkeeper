package main

import (
	"github.com/patraden/ya-practicum-gophkeeper/client/cmd"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
)

func main() {
	log := logger.StdoutConsole(zerolog.DebugLevel).GetZeroLog()
	if err := cmd.NewRootCmd().Execute(); err != nil {
		log.Error().Err(err).Msg("command error")
	}
}

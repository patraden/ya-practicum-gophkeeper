package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	log zerolog.Logger
}

func Stdout(level zerolog.Level) *Logger {
	return &Logger{
		log: zerolog.New(os.Stdout).
			With().
			Timestamp().
			Logger().
			Level(level),
	}
}

func (l *Logger) GetZeroLog() *zerolog.Logger {
	return &l.log
}

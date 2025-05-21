package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

func shortCallerMarshalFunc(_ uintptr, file string, line int) string {
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

type Logger struct {
	log zerolog.Logger
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log.Fatal().Msgf(format, v...)
}

func (l *Logger) Printf(format string, v ...any) {
	l.log.Info().Msgf(format, v...)
}

func Stdout(level zerolog.Level) *Logger {
	zerolog.CallerMarshalFunc = shortCallerMarshalFunc
	zerolog.CallerSkipFrameCount = 2

	return &Logger{
		log: zerolog.New(os.Stdout).
			With().
			Timestamp().
			Caller().
			Logger().
			Level(level),
	}
}

func (l *Logger) GetZeroLog() *zerolog.Logger {
	return &l.log
}

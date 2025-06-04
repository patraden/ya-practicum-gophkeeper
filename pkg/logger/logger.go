package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ipfans/fxlogger"
	"github.com/rs/zerolog"
	"go.uber.org/fx/fxevent"
)

func shortCallerMarshalFunc(_ uintptr, file string, line int) string {
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Logger is a wrapper around zerolog.Logger to provide structured logging.
type Logger struct {
	log zerolog.Logger
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log.Fatal().Msgf(format, v...)
}

func (l *Logger) Printf(format string, v ...any) {
	l.log.Info().Msgf(format, v...)
}

// Stdout initializes and returns a new Logger.
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

// GetLogger returns the zerolog.Logger instance for custom log messages.
func (l *Logger) GetZeroLog() *zerolog.Logger {
	return &l.log
}

// GetFxLogger returns uber fx compatible zerolog.
func (l *Logger) GetFxLogger() func() fxevent.Logger {
	return fxlogger.WithZerolog(l.log)
}

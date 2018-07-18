// Package log provides a global logger with zerolog.
package log

import (
	natsserver "github.com/nats-io/gnatsd/server"
	"github.com/rs/zerolog"
)

// Logger is nats server.Logger compatible logging wrapper for logrus
type natsLogger struct {
	zerolog.Logger
}

// NewNATSLogger creates a new NATS logger that delegates to the provided logger
func NewNATSLogger() natsserver.Logger {
	return natsLogger{Logger}
}

func (a natsLogger) Noticef(format string, v ...interface{}) {
	a.Info().Msgf(format, v...)
}

// Log a fatal error
func (a natsLogger) Fatalf(format string, v ...interface{}) {
	a.Fatal().Msgf(format, v...)
}

// Log an error
func (a natsLogger) Errorf(format string, v ...interface{}) {
	a.Error().Msgf(format, v...)
}

// Log a debug statement
func (a natsLogger) Debugf(format string, v ...interface{}) {
	a.Debug().Msgf(format, v...)
}

// Log a trace statement
func (a natsLogger) Tracef(format string, v ...interface{}) {
	a.Warn().Msgf(format, v...)
}

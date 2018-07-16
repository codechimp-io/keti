// Package log provides a global logger with zerolog.
package log

import (
	stdlog "log"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger is the global logger.
var Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

func init() {
	// log with nanosecond precision time
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"
	zerolog.CallerFieldName = "producer"

	// redirects go's std log to zerolog
	stdlog.SetFlags(0)
	stdlog.SetOutput(Logger)
}

func WithCaller(name string) {
	Logger = Logger.With().Str(zerolog.CallerFieldName, name).Logger()
}

func Infof(format string, v ...interface{}) {
	Logger.Info().Msgf(format, v...)
}

func Debugf(format string, v ...interface{}) {
	Logger.Debug().Msgf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	Logger.Error().Msgf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	Logger.Fatal().Msgf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	Logger.Warn().Msgf(format, v...)
}

func Info(entry string) {
	Logger.Info().Msg(entry)
}

func Debug(entry string) {
	Logger.Debug().Msg(entry)
}

func Error(entry string) {
	Logger.Error().Msg(entry)
}

func Fatal(entry string) {
	Logger.Fatal().Msg(entry)
}

func Warn(entry string) {
	Logger.Warn().Msgf(entry)
}

// Printf sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	Logger.Printf(format, v...)
}

// Print sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) {
	Logger.Print(v...)
}

package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the logger with the specified log level
func Init(logLevel string) {
	zerolog.TimeFieldFormat = time.RFC3339

	// Set log level from config
	level := getLogLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		// Pretty logging for console output
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// JSON logging for file output
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

// getLogLevel converts string level to zerolog.Level
func getLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel // Default to INFO level if unspecified or invalid
	}
}

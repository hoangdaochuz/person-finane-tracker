package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dev/personal-finance-tracker/backend/internal/config"
)

// Init initializes the global logger based on configuration
func Init(cfg *config.Config) {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.App.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Configure time format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set output format based on config
	if cfg.App.LogFormat == "json" {
		// JSON format for production
		log.Logger = zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
	} else {
		// Console format for development
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).Level(level).With().Timestamp().Logger()
	}
}

// Get returns the global logger instance
func Get() zerolog.Logger {
	return log.Logger
}

// With creates a child logger with additional context
func With() zerolog.Context {
	return log.With()
}

// Info logs an info message
func Info() *zerolog.Event {
	return log.Info()
}

// Error logs an error message
func Error() *zerolog.Event {
	return log.Error()
}

// Debug logs a debug message
func Debug() *zerolog.Event {
	return log.Debug()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return log.Warn()
}

// Fatal logs a fatal message and exits
func Fatal() *zerolog.Event {
	return log.Fatal()
}

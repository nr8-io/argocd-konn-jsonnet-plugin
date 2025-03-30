package zino

import (
	"io"

	"github.com/rs/zerolog"
)

type LoggerOption func(*Logger)

// WithLevel sets the log level
func WithLevel(level string) LoggerOption {
	return func(options *Logger) {
		options.Level = level
	}
}

// WithWriter sets the writer for the logger
func WithWriter(writer io.Writer) LoggerOption {
	return func(options *Logger) {
		options.Writer = writer
	}
}

func WithZeroLevel(level zerolog.Level) LoggerOption {
	return func(options *Logger) {
		options.ZeroLevel = level
	}
}

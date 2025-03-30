package zino

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// PinoLevel is a string representation of levels adopted PinoJS logger
type PinoLevel int

// Represents all the Pino log levels accepted
// Note: the Panic level has been added to cover this log level not available in Javascript
const (
	Trace PinoLevel = 10
	Debug PinoLevel = 20
	Info  PinoLevel = 30
	Warn  PinoLevel = 40
	Error PinoLevel = 50
	Fatal PinoLevel = 60
	Panic PinoLevel = 70
)

// ConvertLevel Convert a zerolog log level into the corresponding pino one
func ConvertLevel(level zerolog.Level) int {
	var pinoLevel PinoLevel
	switch level {
	case zerolog.TraceLevel:
		pinoLevel = Trace
	case zerolog.DebugLevel:
		pinoLevel = Debug
	case zerolog.InfoLevel:
		pinoLevel = Info
	case zerolog.WarnLevel:
		pinoLevel = Warn
	case zerolog.ErrorLevel:
		pinoLevel = Error
	case zerolog.FatalLevel:
		pinoLevel = Fatal
	case zerolog.PanicLevel:
		pinoLevel = Panic
	case zerolog.Disabled:
		fallthrough
	case zerolog.NoLevel:
		fallthrough
	default:
		return 0
	}

	return int(pinoLevel)
}

// ParseLevel Parse a string name of the log level and return the corresponding zerolog level
func ParseLevel(level string) (zerolog.Level, error) {
	if len(level) > 0 {
		switch strings.ToLower(level) {
		case "trace":
			return zerolog.TraceLevel, nil
		case "debug":
			return zerolog.DebugLevel, nil
		case "info":
			return zerolog.InfoLevel, nil
		case "warn":
			return zerolog.WarnLevel, nil
		case "error":
			return zerolog.ErrorLevel, nil
		case "fatal":
			return zerolog.FatalLevel, nil
		case "panic":
			return zerolog.PanicLevel, nil
		case "silent":
			return zerolog.Disabled, nil
		default:
			return zerolog.NoLevel, fmt.Errorf("level %s is not recognized", level)
		}
	}

	return zerolog.InfoLevel, nil
}

type LevelHook struct{}

// Add the level as an integer instead of a string to match pino log output
func (h LevelHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Int("level", ConvertLevel(level))
}

// Create a new logger with level removed and default values to match standard pino outpu
func NewLogger(writer io.Writer, level zerolog.Level, disableTimeMs bool) *zerolog.Logger {
	// global default configuration
	zerolog.MessageFieldName = "msg"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.LevelFieldName = ""
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	if disableTimeMs {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	// ignore hostname in case of error
	hostname, _ := os.Hostname()

	log := zerolog.
		New(writer).
		Hook(LevelHook{}).
		With().
		Timestamp().
		Int("pid", os.Getpid()).
		Str("hostname", hostname).
		Logger().
		Level(level)

	return &log
}

// InitOptions are the possible options that can be used to initialize the logger
type InitOptions struct {
	Level         string
	DisableTimeMs bool
	Writer        io.Writer
}

// Init Creates a zerolog logger with custom default properties and custom style
func Init(options InitOptions) (*zerolog.Logger, error) {
	var logWriter io.Writer = os.Stdout
	if options.Writer != nil {
		logWriter = options.Writer
	}

	logLevel, err := ParseLevel(options.Level)
	if err != nil {
		return nil, err
	}

	return NewLogger(logWriter, logLevel, options.DisableTimeMs), nil
}

// InitDefault Creates a zerolog logger with custom default properties
// and custom style using predefined writer and log level
func InitDefault() *zerolog.Logger {
	return NewLogger(os.Stdout, zerolog.InfoLevel, false)
}

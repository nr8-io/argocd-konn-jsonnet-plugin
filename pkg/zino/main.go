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

// InitOptions are the possible options that can be used to initialize the logger
type Logger struct {
	// Zino logger options
	Level     string
	ZeroLevel zerolog.Level
	Writer    io.Writer
}

// Init Creates a zerolog logger with custom default properties and custom style
func NewLogger(level string, options ...LoggerOption) (*zerolog.Logger, error) {
	l := &Logger{
		Level:  level,
		Writer: nil,
	}

	l.Configure(options...)

	zerolog.LevelFieldName = ""      // remove default level field to replace it with pino int level
	zerolog.MessageFieldName = "msg" // replace default message field with pino msg
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs // default time format for pino

	// ignore hostname in case of error
	hostname, _ := os.Hostname()

	log := zerolog.
		New(l.Writer).
		Hook(PinoLevelHook{}).
		With().
		Timestamp().
		Int("pid", os.Getpid()).
		Str("hostname", hostname).
		Logger().
		Level(l.ZeroLevel)

	return &log, nil
}

func (l *Logger) Configure(options ...LoggerOption) error {
	for _, option := range options {
		option(l)
	}

	if l.Level == "" && l.ZeroLevel == 0 {
		l.Level = "info"
	}

	if l.Writer == nil {
		l.Writer = os.Stdout
	}

	if l.ZeroLevel == 0 {
		logLevel, err := ParseLevel(l.Level)
		if err != nil {
			return err
		}
		l.ZeroLevel = logLevel
	}

	return nil
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

type PinoLevelHook struct{}

// Add the level as an integer instead of a string to match pino log output
func (h PinoLevelHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
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
		pinoLevel = 0
	}

	e.Int("level", int(pinoLevel))
}

package logger

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

func NewLogger(service string) zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	return zerolog.New(output).
		With().
		Timestamp().
		Str("service", service).
		Logger()
}

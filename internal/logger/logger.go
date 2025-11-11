package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init() {
	env := os.Getenv("GIN_MODE")
	zerolog.TimeFieldFormat = time.RFC3339

	var level zerolog.Level
	if env == "release" {
		level = zerolog.InfoLevel
	} else {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)

	if env == "release" {
		Log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
		Log = zerolog.New(output).With().Timestamp().Logger()
	}

	Log.Info().
		Str("mode", env).
		Str("level", level.String()).
		Msg("logger initialized")
}

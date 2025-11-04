package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Log zerolog.Logger

func Init() {
	env := os.Getenv("GIN_MODE") // Gin уже использует GIN_MODE (debug/release)
	if env == "release" {
		zerolog.TimeFieldFormat = time.RFC3339
		Log = log.Output(os.Stdout)
	} else {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
		Log = log.Output(output).With().Timestamp().Logger()
	}
	Log.Info().Msg("logger initialized")
}

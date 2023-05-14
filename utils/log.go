package utils

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func GetLogger(module string) zerolog.Logger {
	return log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str("module", module).Logger()
}

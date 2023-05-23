package main

import (
	"autobangumi-go/cmd"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	rootCmd := cobra.Command{}
	rootCmd.AddCommand(cmd.GetBotCommands())
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

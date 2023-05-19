package cmd

import (
	"github.com/spf13/cobra"
)

func GetBotCommands() *cobra.Command {
	rootCmd := cobra.Command{
		Use:   "bot",
		Short: "bot commands",
	}
	rootCmd.AddCommand(GetRunAutoBangumiBotCmd())
	return &rootCmd
}

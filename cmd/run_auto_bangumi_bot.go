package cmd

import (
	"path/filepath"

	"autobangumi-go/bot"
	"autobangumi-go/config"
	"github.com/spf13/cobra"
)

func GetRunAutoBangumiBotCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "run-ab-bot",
		Short: "run auto bangumi bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			configPath, err = filepath.Abs(configPath)
			if err != nil {
				return err
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			abBot, err := bot.NewTGAutoBangumiBot(cfg)
			if err != nil {
				return err
			}
			abBot.Run()
			return nil
		},
	}
	cmd.Flags().String("config", "config.toml", "path to config file")

	return &cmd
}

package cmd

import (
	"os"
	"pikpak-bot/bot"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	ENV_ARIA2_WS_URL    = "ENV_ARIA2_WS_URL"
	ENV_ARIA2_SECRET    = "ENV_ARIA2_SECRET"
	ENV_PIKPAK_USERNAME = "ENV_PIKPAK_USERNAME"
	ENV_PIKPAK_PASSWORD = "ENV_PIKPAK_PASSWORD"
	ENV_TG_BOT_TOKEN    = "ENV_TG_BOT_TOKEN"
	ENV_DOWNLOAD_DIR    = "ENV_DOWNLOAD_DIR"

	FlagAria2WsUrl     = "aria2-ws"
	FlagAria2Secret    = "aria2-secret"
	FlagPikpakUsername = "pikpak-username"
	FlagPikpakPassword = "pikpak-password"
	FlagTGBotToken     = "tg-token"
	FlagDownloadDir    = "dl-dir"
)

func GetBotCommands() *cobra.Command {
	rootCmd := cobra.Command{
		Use:   "bot",
		Short: "bot commands",
	}
	rootCmd.AddCommand(GetRunDLBotCmd())
	return &rootCmd
}

func GetRunDLBotCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "run-dl-bot",
		Short: "run telegram downloader bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			var config *bot.TGDownloaderBotConfig
			configFromEnv := loadConfigFromEnv()
			if configFromEnv.Validate() != nil {
				configFromFlags, err := loadConfigFromCmdFlags(cmd.Flags())
				if err != nil {
					return err
				}
				config = configFromFlags
			} else {
				config = configFromEnv
			}
			downloaderBot, err := bot.NewTGDownloaderBot(config)
			if err != nil {
				return err
			}
			downloaderBot.Run()
			return nil
		},
	}
	cmd.Flags().String(FlagAria2WsUrl, "ws://127.0.0.1:6800/jsonrpc", "aria2 websocket endpoint")
	cmd.Flags().String(FlagAria2Secret, "", "aria2 secret")
	cmd.Flags().String(FlagPikpakUsername, "", "pikpak username")
	cmd.Flags().String(FlagPikpakPassword, "", "pikpak password")
	cmd.Flags().String(FlagTGBotToken, "", "telegram bot token")
	cmd.Flags().String(FlagDownloadDir, "/Volume1/Download", "download directory path")
	return &cmd
}

func loadConfigFromEnv() *bot.TGDownloaderBotConfig {
	config := bot.TGDownloaderBotConfig{}
	config.Aria2WsEndpoint = os.Getenv(ENV_ARIA2_WS_URL)
	config.Aria2Secret = os.Getenv(ENV_ARIA2_SECRET)
	config.PikpakUsername = os.Getenv(ENV_PIKPAK_USERNAME)
	config.PikpakPassword = os.Getenv(ENV_PIKPAK_PASSWORD)
	config.TGBotToken = os.Getenv(ENV_TG_BOT_TOKEN)
	config.DownloadDirectory = os.Getenv(ENV_DOWNLOAD_DIR)
	return &config
}

func loadConfigFromCmdFlags(flags *pflag.FlagSet) (*bot.TGDownloaderBotConfig, error) {
	config := bot.TGDownloaderBotConfig{}
	var err error
	if config.Aria2WsEndpoint, err = flags.GetString(FlagAria2WsUrl); err != nil {
		return nil, err
	}
	if config.Aria2Secret, err = flags.GetString(FlagAria2Secret); err != nil {
		return nil, err
	}
	if config.PikpakUsername, err = flags.GetString(FlagPikpakUsername); err != nil {
		return nil, err
	}
	if config.PikpakPassword, err = flags.GetString(FlagPikpakPassword); err != nil {
		return nil, err
	}
	if config.TGBotToken, err = flags.GetString(FlagTGBotToken); err != nil {
		return nil, err
	}
	if config.DownloadDirectory, err = flags.GetString(FlagDownloadDir); err != nil {
		return nil, err
	}

	return &config, nil
}

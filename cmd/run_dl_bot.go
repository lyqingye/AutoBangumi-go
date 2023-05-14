package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"pikpak-bot/bot"
)

const (
	EnvAria2WsUrl     = "ENV_ARIA2_WS_URL"
	EnvAria2Secret    = "ENV_ARIA2_SECRET"
	EnvPikpakUsername = "ENV_PIKPAK_USERNAME"
	EnvPikpakPassword = "ENV_PIKPAK_PASSWORD"
	EnvTgBotToken     = "ENV_TG_BOT_TOKEN"
	EnvDownloadDir    = "ENV_DOWNLOAD_DIR"
	EnvDbHome         = "ENV_DB_HOME"

	FlagAria2WsUrl     = "aria2-ws"
	FlagAria2Secret    = "aria2-secret"
	FlagPikpakUsername = "pikpak-username"
	FlagPikpakPassword = "pikpak-password"
	FlagTGBotToken     = "tg-token"
	FlagDownloadDir    = "dl-dir"
	FlagDBHome         = "db-home"
)

func GetRunDLBotCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "run-dl-bot",
		Short: "run telegram downloader bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			var config *bot.TGDownloaderBotConfig
			configFromEnv := loadDLBotConfigFromEnv()
			if configFromEnv.Validate() != nil {
				configFromFlags, err := loadDLBotConfigFromCmdFlags(cmd.Flags())
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
	cmd.Flags().String(FlagDBHome, "/Volume1/Cache", "db home directory")

	return &cmd
}

func loadDLBotConfigFromEnv() *bot.TGDownloaderBotConfig {
	config := bot.TGDownloaderBotConfig{}
	config.Aria2WsEndpoint = os.Getenv(EnvAria2WsUrl)
	config.Aria2Secret = os.Getenv(EnvAria2Secret)
	config.PikpakUsername = os.Getenv(EnvPikpakUsername)
	config.PikpakPassword = os.Getenv(EnvPikpakPassword)
	config.TGBotToken = os.Getenv(EnvTgBotToken)
	config.DownloadDirectory = os.Getenv(EnvDownloadDir)
	config.DBHome = os.Getenv(EnvDbHome)

	return &config
}

func loadDLBotConfigFromCmdFlags(flags *pflag.FlagSet) (*bot.TGDownloaderBotConfig, error) {
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
	if config.DBHome, err = flags.GetString(FlagDBHome); err != nil {
		return nil, err
	}
	return &config, nil
}

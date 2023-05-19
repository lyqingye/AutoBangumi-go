package cmd

import (
	"github.com/spf13/pflag"
	"os"
	"pikpak-bot/bot"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

const (
	EnvQbEndpoint      = "ENV_QB_ENDPOINT"
	EnvQbUsername      = "ENV_QB_USERNAME"
	EnvQbPassword      = "ENV_QB_PASSWORD"
	EnvRSSUpdatePeriod = "ENV_RSS_UPDATE_PERIOD"

	FlagQbEndpoint      = "qb-endpoint"
	FlagQbUsername      = "qb-username"
	FlagQbPassword      = "qb-password"
	FlagRssUpdatePeriod = "rss-update-period"
)

func GetRunAutoBangumiBotCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "run-ab-bot",
		Short: "run auto bangumi bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			var config *bot.TGAutoBangumiBotConfig
			configFromEnv, err := loadABBotConfigFromEnv()
			if err != nil || configFromEnv.Validate() != nil {
				configFromFlags, err := loadABBotConfigFromCmdFlags(cmd.Flags())
				if err != nil {
					return err
				}
				config = configFromFlags
			} else {
				config = configFromEnv
			}
			abBot, err := bot.NewTGAutoBangumiBot(config)
			if err != nil {
				return err
			}
			abBot.Run()
			return nil
		},
	}
	cmd.Flags().String(FlagQbEndpoint, "http://localhost:8080", "qb endpoint")
	cmd.Flags().String(FlagQbUsername, "admin", "qb username")
	cmd.Flags().String(FlagQbPassword, "adminadmin", "qb password")
	cmd.Flags().Int64(FlagRssUpdatePeriod, 3600, "rss update period seconds")
	cmd.Flags().String(FlagTGBotToken, "", "telegram bot token")
	cmd.Flags().String(FlagDownloadDir, "/downloads", "download directory path")
	cmd.Flags().String(FlagDBHome, "/cache", "db home directory")
	return &cmd
}

func loadABBotConfigFromEnv() (*bot.TGAutoBangumiBotConfig, error) {
	config := bot.TGAutoBangumiBotConfig{}
	config.QBEndpoint = os.Getenv(EnvQbEndpoint)
	config.QBUsername = os.Getenv(EnvQbUsername)
	config.QBPassword = os.Getenv(EnvQbPassword)
	config.TGBotToken = os.Getenv(EnvTgBotToken)
	config.QBDownloadDir = os.Getenv(EnvDownloadDir)
	config.DBDir = os.Getenv(EnvDbHome)
	rssUpdatePeriod := os.Getenv(EnvRSSUpdatePeriod)
	period, err := strconv.ParseInt(rssUpdatePeriod, 10, 64)
	if err != nil {
		return nil, err
	}
	config.RSSUpdatePeriod = time.Second * time.Duration(period)
	return &config, nil
}

func loadABBotConfigFromCmdFlags(flags *pflag.FlagSet) (*bot.TGAutoBangumiBotConfig, error) {
	config := bot.TGAutoBangumiBotConfig{}
	var err error
	if config.QBEndpoint, err = flags.GetString(FlagQbEndpoint); err != nil {
		return nil, err
	}
	if config.QBUsername, err = flags.GetString(FlagQbUsername); err != nil {
		return nil, err
	}
	if config.QBPassword, err = flags.GetString(FlagQbPassword); err != nil {
		return nil, err
	}
	if rssUpdatePeriod, err := flags.GetInt64(FlagRssUpdatePeriod); err != nil {
		return nil, err
	} else {
		config.RSSUpdatePeriod = time.Second * time.Duration(rssUpdatePeriod)
	}
	if config.TGBotToken, err = flags.GetString(FlagTGBotToken); err != nil {
		return nil, err
	}
	if config.QBDownloadDir, err = flags.GetString(FlagDownloadDir); err != nil {
		return nil, err
	}
	if config.DBDir, err = flags.GetString(FlagDBHome); err != nil {
		return nil, err
	}
	return &config, nil
}

package cmd

import (
	"autobangumi-go/bot"
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

const (
	EnvQbEndpoint           = "ENV_QB_ENDPOINT"
	EnvQbUsername           = "ENV_QB_USERNAME"
	EnvQbPassword           = "ENV_QB_PASSWORD"
	EnvRSSUpdatePeriod      = "ENV_RSS_UPDATE_PERIOD"
	EnvTgBotToken           = "ENV_TG_BOT_TOKEN"
	EnvDownloadDir          = "ENV_DOWNLOAD_DIR"
	EnvDbHome               = "ENV_DB_HOME"
	EnvBangumiTVApiEndpoint = "ENV_BANGUMI_TV_API_ENDPOINT"
	EnvTMDBToken            = "ENV_TMDB_TOKEN"
	EnvBangumiHome          = "ENV_BANGUMI_HOME"
	EnvAria2WsUrl           = "ENV_ARIA2_WS_URL"
	EnvAria2Secret          = "ENV_ARIA2_SECRET"
	EnvAria2DownloadDir     = "ENV_ARIA2_DOWNLOAD_DIR"
	EnvPikpakConfigPath     = "ENV_PIKPAK_CONFIG_PATH"

	FlagQbEndpoint           = "qb-endpoint"
	FlagQbUsername           = "qb-username"
	FlagQbPassword           = "qb-password"
	FlagRssUpdatePeriod      = "rss-update-period"
	FlagTGBotToken           = "tg-token"
	FlagDownloadDir          = "dl-dir"
	FlagDBHome               = "db-home"
	FlagBangumiTVApiEndpoint = "bangumi-tv-api-endpoint"
	FlagTMDBToken            = "tmdb-token"
	FlagBangumiHome          = "bgm-home"
	FlagAria2WsUrl           = "aria2-ws"
	FlagAria2Secret          = "aria2-secret"
	FlagAria2DownloadDir     = "aria2-dir"
	FlagPikpakConfigPath     = "pikpak-config-path"
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
	cmd.Flags().String(FlagDownloadDir, "/downloads", "qb download directory path")
	cmd.Flags().String(FlagDBHome, "/cache", "db home directory")
	cmd.Flags().String(FlagBangumiTVApiEndpoint, "https://api.bgm.tv/v0", "bangumi tv api endpoint")
	cmd.Flags().String(FlagTMDBToken, "", "tmdb token")
	cmd.Flags().String(FlagBangumiHome, "/bangumi", "bangumi config home directory path")
	cmd.Flags().String(FlagAria2WsUrl, "ws://nas.lyqingye.com:8888", "aria2 websocket url")
	cmd.Flags().String(FlagAria2Secret, "", "aria2 secret")
	cmd.Flags().String(FlagAria2DownloadDir, "/downloads", "aria2 download directory path")
	cmd.Flags().String(FlagPikpakConfigPath, "/pikpak.json", "pikpak config file path")
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
	config.TMDBToken = os.Getenv(EnvTMDBToken)
	config.BangumiTVApiEndpoint = os.Getenv(EnvBangumiTVApiEndpoint)
	config.BangumiHome = os.Getenv(EnvBangumiHome)
	rssUpdatePeriod := os.Getenv(EnvRSSUpdatePeriod)
	period, err := strconv.ParseInt(rssUpdatePeriod, 10, 64)
	if err != nil {
		return nil, err
	}
	config.RSSUpdatePeriod = time.Second * time.Duration(period)
	config.Aria2WsUrl = os.Getenv(EnvAria2WsUrl)
	config.Aria2Secret = os.Getenv(EnvAria2Secret)
	config.Aria2DownloadDir = os.Getenv(EnvAria2DownloadDir)
	config.PikPakConfigPath = os.Getenv(EnvPikpakConfigPath)
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
	if config.BangumiTVApiEndpoint, err = flags.GetString(FlagBangumiTVApiEndpoint); err != nil {
		return nil, err
	}
	if config.TMDBToken, err = flags.GetString(FlagTMDBToken); err != nil {
		return nil, err
	}
	if config.DBDir, err = flags.GetString(FlagDBHome); err != nil {
		return nil, err
	}
	if config.BangumiHome, err = flags.GetString(FlagBangumiHome); err != nil {
		return nil, err
	}
	if config.Aria2WsUrl, err = flags.GetString(FlagAria2WsUrl); err != nil {
		return nil, err
	}
	if config.Aria2Secret, err = flags.GetString(FlagAria2Secret); err != nil {
		return nil, err
	}
	if config.Aria2DownloadDir, err = flags.GetString(FlagAria2DownloadDir); err != nil {
		return nil, err
	}
	if config.PikPakConfigPath, err = flags.GetString(FlagPikpakConfigPath); err != nil {
		return nil, err
	}
	return &config, nil
}

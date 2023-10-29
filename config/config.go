package config

import (
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type QBConfig struct {
	Enable      bool   `mapstructure:"Enable"`
	Endpoint    string `mapstructure:"Endpoint"`
	Username    string `mapstructure:"Username"`
	Password    string `mapstructure:"Password"`
	DownloadDir string `mapstructure:"DownloadDir"`
}

func (cfg QBConfig) Validate() error {
	_, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return err
	}
	if cfg.Username == "" || cfg.Password == "" {
		return errors.New("empty username or password")
	}
	return nil
}

type CacheConfig struct {
	CacheDir           string        `mapstructure:"CacheDir"`
	ClearCacheInterval time.Duration `mapstructure:"ClearCacheInterval"`
}

func (cfg CacheConfig) Validate() error {
	return nil
}

type Aria2Config struct {
	WsUrl       string `mapstructure:"WsUrl"`
	Secret      string `mapstructure:"Secret"`
	DownloadDir string `mapstructure:"DownloadDir"`
}

func (cfg Aria2Config) Validate() error {
	if cfg.WsUrl == "" {
		return errors.New("empty ws url")
	}
	return nil
}

type JellyfinConfig struct {
	Endpoint                            string `mapstructure:"Endpoint"`
	Username                            string `mapstructure:"Username"`
	Password                            string `mapstructure:"Password"`
	AutoScanLibraryWhenDownloadFinished bool   `mapstructure:"AutoScanLibraryWhenDownloadFinished"`
}

func (cfg JellyfinConfig) Validate() error {
	if !cfg.AutoScanLibraryWhenDownloadFinished {
		return nil
	}
	if _, err := url.Parse(cfg.Endpoint); err != nil {
		return err
	}
	if cfg.Username == "" {
		return errors.New("username empty")
	}
	if cfg.Password == "" {
		return errors.New("password empty")
	}
	return nil
}

type DBConfig struct {
	LogLevel string `mapstructure:"LogLevel"`

	// Database name
	Name string `mapstructure:"Name"`

	// User name
	User string `mapstructure:"User"`

	// Password of the user
	Password string `mapstructure:"Password"`

	// Host address
	Host string `mapstructure:"Host"`

	// Port number
	Port string `mapstructure:"Port"`

	// MaxConns is the maximum number of connections in the pool.
	MaxConns int `mapstructure:"MaxConns"`
}

func (cfg DBConfig) Validate() error {
	if cfg.Name == "" {
		return errors.New("database name empty")
	}

	if cfg.User == "" {
		return errors.New("user empty")
	}

	if cfg.Password == "" {
		return errors.New("password empty")
	}

	if cfg.Port == "" {
		return errors.New("port  empty")
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return errors.Wrap(err, "parse port error")
	}
	if port <= 0 || port >= 65535 {
		return errors.Errorf("invalid port: %d", port)
	}
	return nil
}

type AutoBangumiConfig struct {
	BangumiCompleteInterval time.Duration `mapstructure:"BangumiCompleteInterval"`
	RssRefreshInterval      time.Duration `mapstructure:"RssRefreshInterval"`
}

type TMDBConfig struct {
	Token string `mapstructure:"Token"`
}

func (cfg TMDBConfig) Validate() error {
	if cfg.Token == "" {
		return errors.New("token is empty")
	}
	return nil
}

type BangumiTVConfig struct {
	Endpoint string `mapstructure:"Endpoint"`
}

func (cfg BangumiTVConfig) Validate() error {
	_, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return err
	}
	return nil
}

type MikanConfig struct {
	Endpoint string `mapstructure:"Endpoint"`
}

func (cfg MikanConfig) Validate() error {
	_, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return err
	}
	return nil
}

type TelegramBotConfig struct {
	Enable bool   `mapstructure:"Enable"`
	Token  string `mapstructure:"Token"`
}

func (cfg TelegramBotConfig) Validate() error {
	if cfg.Enable && cfg.Token == "" {
		return errors.New("token empty")
	}
	return nil
}

type PikpakConfig struct {
	OfflineDownloadTimeout time.Duration `mapstructure:"OfflineDownloadTimeout"`
}

type WebDAVConfig struct {
	ImportBangumiOnStartup bool   `mapstructure:"ImportBangumiOnStartup"`
	Host                   string `mapstructure:"Host"`
	Username               string `mapstructure:"Username"`
	Password               string `mapstructure:"Password"`
	Dir                    string `mapstructure:"Dir"`
}

func (cfg WebDAVConfig) Validate() error {
	if !cfg.ImportBangumiOnStartup {
		return nil
	}
	if _, err := url.Parse(cfg.Host); err != nil {
		return err
	}

	if cfg.Username == "" {
		return errors.New("empty username")
	}
	if cfg.Password == "" {
		return errors.New("empty password")
	}
	if cfg.Dir == "" {
		return errors.New("empty dir")
	}
	return nil
}

type Config struct {
	DB          DBConfig
	Cache       CacheConfig
	AutoBangumi AutoBangumiConfig

	TMDB      TMDBConfig
	BangumiTV BangumiTVConfig
	Mikan     MikanConfig

	Aria2  Aria2Config
	QB     QBConfig
	Pikpak PikpakConfig

	TelegramBot TelegramBotConfig
	WebDAV      WebDAVConfig
	Jellyfin    JellyfinConfig
}

func (config *Config) Validate() error {
	if err := config.QB.Validate(); err != nil {
		return errors.Wrap(err, "QBConfig Validate Error")
	}
	if err := config.Aria2.Validate(); err != nil {
		return errors.Wrap(err, "Aria2Config Validate Error")
	}
	if err := config.TMDB.Validate(); err != nil {
		return errors.Wrap(err, "TMDBConfig Validate Error")
	}
	if err := config.Mikan.Validate(); err != nil {
		return errors.Wrap(err, "MikanConfig Validate Error")
	}
	if err := config.BangumiTV.Validate(); err != nil {
		return errors.Wrap(err, "BangumiTV Validate Error")
	}
	if err := config.DB.Validate(); err != nil {
		return errors.Wrap(err, "DBConfig Validate Error")
	}
	if err := config.TelegramBot.Validate(); err != nil {
		return errors.Wrap(err, "TelegramBotConfig Validate Error")
	}
	if err := config.WebDAV.Validate(); err != nil {
		return errors.Wrap(err, "WebDAVConfig Validate Error")
	}
	if err := config.Jellyfin.Validate(); err != nil {
		return errors.Wrap(err, "JellyfinConfig Validate Error")
	}
	return nil
}

// Load loads the configuration
func Load(configFilePath string) (*Config, error) {
	var cfg Config
	viper.SetConfigType("toml")

	if configFilePath != "" {
		dirName, fileName := filepath.Split(configFilePath)
		fileExtension := strings.TrimPrefix(filepath.Ext(fileName), ".")
		fileNameWithoutExtension := strings.TrimSuffix(fileName, "."+fileExtension)
		viper.AddConfigPath(dirName)
		viper.SetConfigName(fileNameWithoutExtension)
		viper.SetConfigType(fileExtension)
	}
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("AUTO_BANGUMI")
	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		ok := errors.As(err, &configFileNotFoundError)
		if ok {
			return nil, errors.Errorf("config not found in: %s", configFilePath)
		} else {
			return nil, err
		}
	}

	err = viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc()))
	if err != nil {
		return nil, err
	}

	return &cfg, cfg.Validate()
}

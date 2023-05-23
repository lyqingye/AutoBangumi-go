package bot_test

import (
	"autobangumi-go/bot"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAutoBangumi_Start(t *testing.T) {
	config := bot.AutoBangumiConfig{
		QBEndpoint:           "http://localhost:8080",
		QBUsername:           "admin",
		QBPassword:           "adminadmin",
		QBDownloadDir:        "",
		DBDir:                "./db_cache",
		RSSUpdatePeriod:      time.Minute * 5,
		TMDBToken:            "702225c8ca516a5be2f062988438bfda",
		BangumiTVApiEndpoint: "https://api.bgm.tv/v0",
	}
	autoBangumiBot, err := bot.NewAutoBangumi(&config)
	require.NoError(t, err)
	require.NotNil(t, autoBangumiBot)
	require.NoError(t, err)
	autoBangumiBot.Start()
}

func TestAutoBangumi_Start2(t *testing.T) {
	config := bot.AutoBangumiConfig{
		QBEndpoint:           "http://nas.lyqingye.com:8888",
		QBUsername:           "admin",
		QBPassword:           "adminadmin",
		QBDownloadDir:        "",
		DBDir:                "./db_cache",
		TMDBToken:            "702225c8ca516a5be2f062988438bfda",
		BangumiTVApiEndpoint: "https://api.bgm.tv/v0",
		RSSUpdatePeriod:      time.Minute * 5,
	}
	autoBangumiBot, err := bot.NewAutoBangumi(&config)
	require.NoError(t, err)
	require.NotNil(t, autoBangumiBot)
	require.NoError(t, err)
	autoBangumiBot.Start()
}

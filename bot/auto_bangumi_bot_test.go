package bot_test

import (
	"testing"

	"autobangumi-go/bot"
	"autobangumi-go/config"
	"github.com/stretchr/testify/require"
)

func TestAutoBangumi_Start2(t *testing.T) {
	cfg, err := config.Load("../config/config.remote.toml")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	autoBangumiBot, err := bot.NewAutoBangumi(cfg)
	require.NoError(t, err)
	require.NotNil(t, autoBangumiBot)
	require.NoError(t, err)
	//require.NoError(t, autoBangumiBot.AddPikpakAccount("kexed92267@jybra.com", "WOAIxiaokeai1314"))
	//autoBangumiBot.Start()
	err = autoBangumiBot.RemoveBangumi("租借女友")
	require.NoError(t, err)

	//fs, err := bot.NewWebDavFileSystem("http://nas.lyqingye.com:5005", "lyqingye", "WOAIxiaokeai.1314")
	//require.NoError(t, err)

	//err = autoBangumiBot.ScanFileSystemBangumis(fs, "/anime")
	//require.NoError(t, err)
	//err = autoBangumiBot.AddWebDAVBangumiToStorage("http://nas.lyqingye.com:5005", "lyqingye", "WOAIxiaokeai.1314", "/anime")
}

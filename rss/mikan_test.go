package rss_test

import (
	"os"
	"pikpak-bot/bus"
	"pikpak-bot/db"
	"pikpak-bot/mdb"
	"pikpak-bot/rss"
	"regexp"
	"strings"
	"testing"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/stretchr/testify/require"
)

func TestParseMikanRss(t *testing.T) {
	dir := "./parser_cache"
	db, err := db.NewDB(dir)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}()
	tmdbClient, err := tmdb.Init("702225c8ca516a5be2f062988438bfda")
	require.NoError(t, err)
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	eb := bus.NewEventBus()
	eb.Start()
	parser, err := rss.NewMikanRSSParser("https://mikanani.me/RSS/Bangumi?bangumiId=444", eb, db, tmdbClient, bangumiTVClient)
	require.NoError(t, err)
	rssInfo, err := parser.Parse()
	require.NoError(t, err)
	require.NotNil(t, rssInfo)
}

func TestTMDB(t *testing.T) {
	tmdbClient, err := tmdb.Init("702225c8ca516a5be2f062988438bfda")
	require.NoError(t, err)
	options := map[string]string{
		"language": "zh-CN",
	}
	searchResult, err := tmdbClient.GetSearchTVShow("我的青春恋爱物语果然有问题。完 ", options)
	require.NoError(t, err)
	require.NotNil(t, searchResult)
	tvDetails, err := tmdbClient.GetTVDetails(int(searchResult.Results[0].ID), options)
	require.NoError(t, err)
	require.NotNil(t, tvDetails)
}

func TestNormalizationSearchTitle(t *testing.T) {
	t.Log(normalizationSearchTitle("总之就是非常可爱 第二季"))
	t.Log(normalizationSearchTitle("总之就是非常可爱 第三季"))
	t.Log(normalizationSearchTitle("总之就是非常可爱 第三期"))
	t.Log(normalizationSearchTitle("总之就是非常可爱 Season 3"))
	t.Log(normalizationSearchTitle("总之就是非常可爱 Season3"))
}

func normalizationSearchTitle(keyword string) string {
	patterns := []string{
		"第([[:digit:]]+|\\p{Han}+)季",
		"第([[:digit:]]+|\\p{Han}+)期",
		"Season\\s*\\d+",
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		keyword = strings.ReplaceAll(keyword, re.FindString(keyword), "")
	}
	return keyword
}

package mikan_test

import (
	"autobangumi-go/bus"
	"autobangumi-go/db"
	"autobangumi-go/mdb"
	"autobangumi-go/rss/mikan"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	bangumitypes "autobangumi-go/bangumi"

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
	tmdbClient, err := mdb.NewTMDBClient("702225c8ca516a5be2f062988438bfda")
	require.NoError(t, err)
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	eb := bus.NewEventBus()
	eb.Start()
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me/RSS/Bangumi?bangumiId=444", eb, db, tmdbClient, bangumiTVClient)
	require.NoError(t, err)
	rssInfo, err := parser.Parse()
	require.NoError(t, err)
	require.NotNil(t, rssInfo)
}

func TestMikanSearch(t *testing.T) {
	dir := "./parser_cache"
	db, err := db.NewDB(dir)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}()
	tmdbClient, err := mdb.NewTMDBClient("702225c8ca516a5be2f062988438bfda")
	require.NoError(t, err)
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	eb := bus.NewEventBus()
	eb.Start()
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me/RSS/Bangumi?bangumiId=444", eb, db, tmdbClient, bangumiTVClient)
	require.NoError(t, err)
	//result, err := parser.Search("式守同学不只可爱而已")
	//require.NoError(t, err)
	//require.NotNil(t, result)

	result, err := parser.Search2("她去公爵家的理由")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestMikanSearch3(t *testing.T) {
	dir := "./parser_cache"
	db, err := db.NewDB(dir)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}()
	tmdbClient, err := mdb.NewTMDBClient("702225c8ca516a5be2f062988438bfda")
	require.NoError(t, err)
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	eb := bus.NewEventBus()
	eb.Start()
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me/RSS/Bangumi?bangumiId=444", eb, db, tmdbClient, bangumiTVClient)
	require.NoError(t, err)

	result, err := parser.Search3("2984")
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestMikanCompleteBangumi(t *testing.T) {
	resp, err := http.Get("https://mikanani.me/RSS/Search?searchstr=%E6%88%91%E7%9A%84%E9%9D%92%E6%98%A5")
	require.NoError(t, err)
	require.NotNil(t, resp)
	bz, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotNil(t, bz)

	dir := "./parser_cache"
	db, err := db.NewDB(dir)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}()
	tmdbClient, err := mdb.NewTMDBClient("702225c8ca516a5be2f062988438bfda")
	require.NoError(t, err)
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	eb := bus.NewEventBus()
	eb.Start()
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me/RSS/Bangumi?bangumiId=444", eb, db, tmdbClient, bangumiTVClient)
	require.NoError(t, err)
	bangumi := bangumitypes.Bangumi{
		Info: bangumitypes.BangumiInfo{
			Title: "乙女游戏世界对路人角色很不友好",
		},
	}
	err = parser.CompleteBangumi(&bangumi)
	require.NoError(t, err)
	season := bangumi.Seasons[1]
	season.Complete = append(season.Complete, 1, 2, 3, 4)
	season.Episodes = nil
	bangumi.Seasons[1] = season
	err = parser.CompleteBangumi(&bangumi)
	require.NoError(t, err)
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

func TestCollectionRegexp(t *testing.T) {

}

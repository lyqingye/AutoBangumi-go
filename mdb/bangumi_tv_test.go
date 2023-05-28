package mdb_test

import (
	"autobangumi-go/mdb"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBangumiTxSubjects(t *testing.T) {
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	require.NotNil(t, bangumiTVClient)
	subject, err := bangumiTVClient.GetSubjects(404804)
	require.NoError(t, err)
	require.NotNil(t, subject)
}

func TestSearchSubject(t *testing.T) {
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	require.NotNil(t, bangumiTVClient)
	subject, err := bangumiTVClient.SearchAnime("异世界舅舅")
	require.NoError(t, err)
	require.NotNil(t, subject)
	t.Log(subject.GetAliasNames())
}

func TestSearchSubjec3t(t *testing.T) {
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	require.NotNil(t, bangumiTVClient)
	subject, err := bangumiTVClient.SearchAnime("我的青春恋爱物语果然有问题。续")
	require.NoError(t, err)
	require.NotNil(t, subject)
	t.Log(subject.GetAliasNames())
}

func TestSearchAnime2(t *testing.T) {
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	require.NotNil(t, bangumiTVClient)
	subject, err := bangumiTVClient.SearchAnime2("赤发的白雪姬")
	require.NoError(t, err)
	require.NotNil(t, subject)
}

func TestMe(t *testing.T) {
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	require.NotNil(t, bangumiTVClient)
	meInfo, err := bangumiTVClient.Me()
	require.NoError(t, err)
	require.NotNil(t, meInfo)
}

func TestCollections(t *testing.T) {
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	require.NotNil(t, bangumiTVClient)
	err = bangumiTVClient.SetAccessToken("")
	require.NoError(t, err)
	collections, err := bangumiTVClient.Collections(mdb.CollectionTypeDoing, mdb.SubjectTypeAnime)
	require.NoError(t, err)
	require.NotNil(t, collections)
}

func TestGetCalendar(t *testing.T) {
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	require.NoError(t, err)
	require.NotNil(t, bangumiTVClient)
	calendar, err := bangumiTVClient.GetCalendar()
	require.NoError(t, err)
	require.NotNil(t, calendar)
}

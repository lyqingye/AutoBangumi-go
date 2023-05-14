package mdb_test

import (
	"pikpak-bot/mdb"
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

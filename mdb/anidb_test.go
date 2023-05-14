package mdb_test

import (
	"github.com/stretchr/testify/require"
	"pikpak-bot/mdb"
	"testing"
)

func TestAniDBLoadCache(t *testing.T) {
	client, err := mdb.NewAniDBClient("/tmp/anidb_cache", "lyqingye", "1")
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NoError(t, client.InitDumpData())
	aid, err := client.SearchTitle("伊甸星原 第二季")
	require.NoError(t, err)
	t.Log(aid)
	anime, err := client.GetAnime(aid)
	require.NoError(t, err)
	require.NotNil(t, anime)

}

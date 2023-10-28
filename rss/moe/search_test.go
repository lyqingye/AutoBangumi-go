package moe_test

import (
	"autobangumi-go/rss/moe"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSearch(t *testing.T) {
	moe, err := moe.NewBangumiMoe()
	require.NoError(t, err)
	require.NotNil(t, moe)
	resp, err := moe.SearchTagByKeyword("在异世界获得超强能力的我，在现实世界照样无敌～等级提升改变人生命运～")
	require.NoError(t, err)
	require.NotNil(t, resp)

	torrents, err := moe.SearchTorrentByTag([]string{resp.Tag[0].ID}, 1)
	require.NoError(t, err)
	require.NotNil(t, torrents)
}

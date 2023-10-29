package mdb_test

import (
	"testing"

	"autobangumi-go/mdb"

	"github.com/stretchr/testify/require"
)

func TestTmDB(t *testing.T) {
	tmdb, err := mdb.NewTMDBClient("702225c8ca516a5be2f062988438bfda")
	require.NoError(t, err)
	tv, err := tmdb.SearchTVShowByKeyword("辉夜大小姐想让我告白")
	require.NoError(t, err)
	require.NotNil(t, tv)
}

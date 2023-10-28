package mdb_test

import (
	"autobangumi-go/mdb"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTmDB(t *testing.T) {
	tmdb, err := mdb.NewTMDBClient("")
	require.NoError(t, err)
	tv, err := tmdb.SearchTVShowByKeyword("勇者死了")
	require.NoError(t, err)
	require.NotNil(t, tv)
}

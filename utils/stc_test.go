package utils_test

import (
	"autobangumi-go/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStcClear(t *testing.T) {
	stc := utils.NewSimpleTTLCache(time.Second)
	stc.Put("100", "100", time.Second*5)
	stc.Put("101", "100", time.Second*5)
	stc.Put("102", "100", time.Second*5)
	stc.Put("103", "100", time.Second*5)

	time.Sleep(time.Second * 10)
	_, found := stc.Get("100")
	require.False(t, found)
	_, found = stc.Get("101")
	require.False(t, found)
	_, found = stc.Get("102")
	require.False(t, found)
	_, found = stc.Get("103")
	require.False(t, found)
}

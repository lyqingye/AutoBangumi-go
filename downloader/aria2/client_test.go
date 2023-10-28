package aria2_test

import (
	"autobangumi-go/downloader/aria2"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAria2(t *testing.T) {
	client, err := aria2.NewClient("ws://nas.lyqingye.com:9000/jsonrpc", "123456", "/downloads/test")
	require.NoError(t, err)
	require.NotNil(t, client)
	for {
		_, err := client.ListActiveTasks()
		require.NoError(t, err)
		time.Sleep(time.Second * 5)
	}
}

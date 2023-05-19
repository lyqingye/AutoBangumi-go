package rss_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"pikpak-bot/bangumi"
	"pikpak-bot/bus"
	"pikpak-bot/db"
	"pikpak-bot/downloader/qibittorrent"
	"pikpak-bot/rss"
	"testing"
	"time"

	torrent "github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/require"
)

func TestRss(t *testing.T) {
	eb := bus.NewEventBus()
	eb.Start()
	dir := "./parser_cache"
	db, err := db.NewDB(dir)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func() {
		_ = db.Close()
		//_ = os.RemoveAll(dir)
	}()
	rssMan, err := rss.NewRSSManager(eb, db, time.Hour)
	require.NoError(t, err)
	require.NotNil(t, rssMan)
	err = rssMan.AddMikanRss("https://mikanani.me/RSS/Bangumi?bangumiId=3001")
	require.NoError(t, err)

	qb, err := qibittorrent.NewQbittorrentClient("http://nas.lyqingye.com:8888", "admin", "adminadmin", "/downloads")
	require.NoError(t, err)
	require.NotNil(t, qb)
	err = qb.Login()
	require.NoError(t, err)

	eb.Subscribe(bus.RSSTopic, AutoDownload{
		client: qb,
	})

	// refresh rss
	rssMan.Refresh()
}

type AutoDownload struct {
	client *qibittorrent.QbittorrentClient
}

func (dl AutoDownload) HandleEvent(event bus.Event) {
	if event.EventType == bus.RSSUpdateEventType {
		if episode, ok := event.Inner.(bangumi.Episode); ok {
			newName := fmt.Sprintf("[%s] S%02dE%02d", episode.BangumiTitle, episode.Season, episode.EPNumber)
			torr, err := torrent.Load(bytes.NewBuffer(episode.Torrent))
			if err != nil {
				panic(err)
			}
			//FIXME:
			info, err := torr.UnmarshalInfo()
			if err != nil {
				panic(err)
			}
			opts := qibittorrent.AddTorrentOptions{
				Paused: true,
				Rename: newName,
			}
			_, err = dl.client.AddTorrentEx(&opts, episode.Torrent, episode.BangumiTitle)
			if err != nil {
				panic(err)
			}
			hash := torr.HashInfoBytes().HexString()
			for _, fi := range info.Files {
				oldPath := fi.DisplayPath(&info)
				err = dl.client.RenameFile(hash, oldPath, fmt.Sprintf("%s%s", newName, filepath.Ext(oldPath)))
				if err != nil {
					panic(err)
				}
			}
			err = dl.client.ResumeTorrents([]string{hash})
			if err != nil {
				panic(err)
			}
		}
	}
}

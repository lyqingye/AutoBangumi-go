package bangumi

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/studio-b12/gowebdav"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestBangumiManagerConfigMigrate(t *testing.T) {
	var bangumis []Bangumi
	err := filepath.WalkDir("/tmp/bangumi_cache", func(path string, d fs.DirEntry, err error) error {
		bz, err := os.ReadFile(path)
		if err == nil {
			var bangumi Bangumi
			err = json.Unmarshal(bz, &bangumi)
			if err != nil {
				return err
			}
			for seasonNum, season := range bangumi.Seasons {
				for i, ep := range season.Episodes {
					if season.IsComplete(ep.Number) {
						ep.DownloadState = DownloadState{
							Downloader: "qb",
							TaskId:     ep.TorrentHash,
						}
					} else {
						ep.DownloadState = DownloadState{
							Downloader: "",
						}
					}
					season.Episodes[i] = ep
				}

				bangumi.Seasons[seasonNum] = season
			}
			bz, err = json.MarshalIndent(&bangumi, "", "    ")
			if err != nil {
				return err
			}
			err = os.WriteFile(path, bz, os.ModePerm)
			if err != nil {
				return err
			}
			bangumis = append(bangumis, bangumi)
		}
		return nil
	})
	require.NoError(t, err)
}

func TestWebDav(t *testing.T) {
	client := gowebdav.NewClient("http://nas.lyqingye.com:5005", "lyqingye", "WOAIxiaokeai.1314")
	err := client.Connect()
	require.NoError(t, err)
	files, err := client.ReadDir("/anime")
	require.NoError(t, err)
	for _, fi := range files {
		t.Logf(fi.Name())
		if fi.IsDir() {
			seasonDirs, err := client.ReadDir(filepath.Join("/anime", fi.Name()))
			require.NoError(t, err)
			for _, season := range seasonDirs {
				if season.IsDir() {
					episodes, err := client.ReadDir(filepath.Join("/anime", fi.Name(), season.Name()))
					require.NoError(t, err)
					for _, episode := range episodes {
						seasonNumber, episodeNumber, err := ParseEpisodeFilename(episode.Name())
						if err == nil {
							t.Logf("%v %v", seasonNumber, episodeNumber)
						} else {
							t.Error(err)
						}
					}
				}
			}
		}
	}
}

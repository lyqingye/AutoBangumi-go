package bangumi

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
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

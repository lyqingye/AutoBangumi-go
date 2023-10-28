package bangumi

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/nssteinbrenner/anitogo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/studio-b12/gowebdav"
)

type TestManagerSuite struct {
	suite.Suite
	mgr IManager
}

func (s *TestManagerSuite) SetupTest() {

}

func TestRunManagerSuite(t *testing.T) {
	suite.Run(t, new(TestManagerSuite))
}

func (s *TestManagerSuite) TestDownloadService() {

}

func TestWebDav2(t *testing.T) {
	client := gowebdav.NewClient("http://nas.lyqingye.com:5005", "lyqingye", "WOAIxiaokeai.1314")
	err := client.Connect()
	require.NoError(t, err)
	animeName := "樱花庄的宠物女孩"
	season := 1
	baseDir := fmt.Sprintf("/anime/%s/SeasonNum %02d", animeName, season)
	dryRun := false
	files, err := client.ReadDir(baseDir)
	require.NoError(t, err)

	var remainFiles []string
	for _, fi := range files {
		ret := anitogo.Parse(fi.Name(), anitogo.DefaultOptions)
		if strings.HasSuffix(fi.Name(), ".nfo") || strings.HasSuffix(fi.Name(), ".jpg") || strings.HasSuffix(fi.Name(), ".png") || strings.HasSuffix(fi.Name(), ".flac") || strings.Contains(fi.Name(), "予告") {
			if dryRun {
				t.Logf("[Remove] -> %s", fi.Name())
			} else {
				_ = client.Remove(filepath.Join(baseDir, fi.Name()))
			}
			continue
		}
		if len(ret.EpisodeNumber) > 0 {
			epNumber, err := strconv.ParseUint(ret.EpisodeNumber[0], 10, 32)
			if err != nil {
				continue
			}

			ext := filepath.Ext(fi.Name())
			if ext == "" {
				continue
			}
			var newName string
			isSubtitle := false
			for _, suffix := range []string{".sc.ass", ".tc.ass", ".chs.ass", ".cht.ass", ".Chs&Jap.ass", ".Cht&Jap.ass", ".Jap.ass"} {
				if strings.HasSuffix(fi.Name(), suffix) {
					newName = fmt.Sprintf("[%s] S%02dE%02d%s", animeName, season, epNumber, suffix)
					isSubtitle = true
					break
				}
			}
			if !isSubtitle {
				newName = fmt.Sprintf("[%s] S%02dE%02d%s", animeName, season, epNumber, ext)
			}

			if len(ret.AnimeType) == 1 {
				if dryRun {
					t.Logf("[Remove] -> %s", fi.Name())
				} else {
					_ = client.Remove(filepath.Join(baseDir, fi.Name()))
				}
				continue
			}

			if dryRun {
				t.Logf("[Rename] %s -> %s", fi.Name(), newName)
			} else {
				_ = client.Rename(filepath.Join(baseDir, fi.Name()), filepath.Join(baseDir, newName), true)
			}
			remainFiles = append(remainFiles, newName)
		} else {
			if dryRun {
				t.Logf("[Remove] -> %s", fi.Name())
			} else {
				_ = client.Remove(filepath.Join(baseDir, fi.Name()))
			}
		}
	}

	t.Log("------------------------------------------------------------")
	for _, fi := range remainFiles {
		t.Logf(fi)
	}
}

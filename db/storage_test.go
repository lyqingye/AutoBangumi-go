package db_test

import (
	"os"
	"testing"
	"time"

	"autobangumi-go/bangumi"
	"autobangumi-go/config"
	"autobangumi-go/db"
	"autobangumi-go/downloader/pikpak"
	"autobangumi-go/mdb"
	"autobangumi-go/rss/mikan"
	"autobangumi-go/rss/mikan/cache"
	"github.com/stretchr/testify/suite"
)

type TestBackendSuite struct {
	suite.Suite
	backend *db.Backend
}

func (suite *TestBackendSuite) SetupTest() {
	cfg, err := config.Load("../config/config.debug.toml")
	suite.Require().NoError(err)
	suite.Require().NotNil(cfg)

	backend, err := db.NewBackend(cfg.DB)
	suite.NoError(err)
	suite.NotNil(backend)
	suite.backend = backend
}

func TestBackend(t *testing.T) {
	suite.Run(t, new(TestBackendSuite))
}

func (suite *TestBackendSuite) TestQuery() {
	err := suite.backend.ListBangumis(nil, func(bgm bangumi.Bangumi) bool {
		suite.T().Log(bgm.GetTitle())

		seasons, err := bgm.GetSeasons()
		suite.NoError(err)
		for _, season := range seasons {
			suite.T().Log(season.GetNumber(), season.GetEpCount())

			episodes, err := season.GetEpisodes()
			suite.NoError(err)
			for _, episode := range episodes {
				suite.T().Log(episode.GetNumber())
				resources, err := episode.GetResources()
				suite.NoError(err)
				suite.NotNil(resources)
				return false
			}
		}
		return false
	})
	suite.NoError(err)
}

func (suite *TestBackendSuite) TestBasic() {
	var testData = mikan.Bangumi{
		Info: mikan.BangumiInfo{
			Title:  "test bangumi",
			TmDBId: 33,
		},
		Seasons: map[uint]mikan.Season{
			1: {
				SubjectId:      222,
				MikanBangumiId: "11",
				Number:         1,
				EpCount:        24,
				Episodes: map[uint]mikan.Episode{
					1: {
						Number: 1,
						Resources: []mikan.TorrentResource{
							{
								RawFilename:    "11",
								Subgroup:       "11",
								Magnet:         "hash1",
								TorrentHash:    "hash1",
								Torrent:        []byte("test bz"),
								TorrentPubDate: time.Time{},
								FileSize:       0,
								SubtitleLang:   nil,
								Resolution:     "",
								Type:           "",
							},
							{
								RawFilename:    "11",
								Subgroup:       "11",
								Magnet:         "hash2",
								TorrentHash:    "hash2",
								Torrent:        []byte("test bz1"),
								TorrentPubDate: time.Time{},
								FileSize:       0,
								SubtitleLang:   nil,
								Resolution:     "",
								Type:           "",
							},
							{
								RawFilename:    "11",
								Subgroup:       "11",
								Magnet:         "hash3",
								TorrentHash:    "hash3",
								Torrent:        []byte("test bz2"),
								TorrentPubDate: time.Time{},
								FileSize:       0,
								SubtitleLang:   nil,
								Resolution:     "",
								Type:           "",
							},
						},
					},
				},
			},
		},
	}

	err := suite.backend.AddBangumi(nil, &testData)
	suite.NoError(err)

	bgm, err := suite.backend.GetBgmByTitle(nil, "test bangumi")
	suite.NoError(err)
	suite.NotNil(bgm)

	seasons, err := bgm.GetSeasons()
	suite.NoError(err)
	suite.NotNil(seasons)
	suite.Len(seasons, len(testData.Seasons))
	for _, season := range seasons {
		episodes, err := season.GetEpisodes()
		suite.NoError(err)
		suite.NotNil(episodes)
		for _, episode := range episodes {
			suite.NoError(suite.backend.MarkEpisodeDownloaded(nil, episode))
			_, err := episode.GetResources()
			suite.NoError(err)
			resources, err := suite.backend.GetValidEpisodeResources(nil, episode)
			suite.NoError(err)
			suite.NotNil(resources)
		}
		suite.NoError(suite.backend.MarkSeasonDownloaded(nil, season, true))
	}

	suite.NoError(suite.backend.MarkBangumiDownloaded(nil, bgm, true))
}

func (suite *TestBackendSuite) TestAddMikanBangumi() {
	parser, err := NewTestMikanParser()
	suite.NoError(err)
	bgm, err := parser.Search2("因想当冒险者而前往大都市的女儿已经升到了S级")
	suite.NoError(err)
	err = suite.backend.AddBangumi(nil, bgm)
	suite.NoError(err)
}

func NewTestMikanParser() (*mikan.MikanRSSParser, error) {
	tmdbClient, err := mdb.NewTMDBClient("702225c8ca516a5be2f062988438bfda")
	if err != nil {
		return nil, err
	}
	bangumiTVClient, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	if err != nil {
		return nil, err
	}
	cacheDB, err := db.NewDB("test_cache")
	if err != nil {
		return nil, err
	}
	//defer RemoveParserCache()
	cm := cache.NewKVCacheManager(cacheDB)
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me", tmdbClient, bangumiTVClient, cm)
	if err != nil {
		return nil, err
	}
	return parser, nil
}

func RemoveParserCache() {
	_ = os.RemoveAll("test_cache")
}

func (suite *TestBackendSuite) TestPikpakAccounts() {
	testAcc := pikpak.Account{
		Username:       "test_acc",
		Password:       "test_password",
		State:          "normal",
		RestrictedTime: time.Now().Unix(),
	}
	suite.Require().NoError(suite.backend.AddAccount(testAcc))

	accounts, err := suite.backend.ListAccounts()
	suite.Require().NoError(err)
	suite.Require().Len(accounts, 1)
	suite.Require().Equal(testAcc, accounts[0])

	accounts, err = suite.backend.ListAccountsByState(testAcc.State)
	suite.Require().NoError(err)
	suite.Require().Len(accounts, 1)
	suite.Require().Equal(testAcc, accounts[0])

	testAcc.Password = "change password"
	suite.Require().NoError(suite.backend.UpdateAccount(testAcc))

	queryAcc, err := suite.backend.GetAccount(testAcc.Username)
	suite.Require().NoError(err)
	suite.Require().Equal(testAcc, queryAcc)
}

func (suite *TestBackendSuite) TestDownloadHistory() {
	suite.TestBasic()
	err := suite.backend.ListBangumis(nil, func(bgm bangumi.Bangumi) bool {
		seasons, err := bgm.GetSeasons()
		suite.Require().NoError(err)
		for _, season := range seasons {
			episodes, err := season.GetEpisodes()
			suite.Require().NoError(err)
			for _, episode := range episodes {
				resources, err := episode.GetResources()
				if len(resources) == 0 {
					continue
				}
				suite.Require().NoError(err)
				history, err := suite.backend.AddEpisodeDownloadHistory(nil, episode, resources[0].GetTorrentHash())
				suite.Require().NoError(err)
				suite.Require().NotNil(history)

				history.SetDownloader(bangumi.PikpakDownloader, "", bangumi.TryDownload, nil)
				suite.Require().NoError(suite.backend.UpdateDownloadHistory(nil, history))
				suite.Require().NoError(suite.backend.RemoveEpisodeDownloadHistory(nil, episode))

				nullHistory, err := suite.backend.GetEpisodeDownloadHistory(nil, episode)
				suite.Require().NoError(err)
				suite.Require().Nil(nullHistory)

			}
		}
		return true

	})
	suite.Require().NoError(err)
}

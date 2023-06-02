package db_test

import (
	"autobangumi-go/db"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestBackendSuite struct {
	suite.Suite
	backend *db.Backend
}

func (suite *TestBackendSuite) SetupTest() {
	backend, err := db.NewBackend(":memory:")
	suite.NoError(err)
	suite.NotNil(backend)
	suite.backend = backend
}

func TestBackend(t *testing.T) {
	suite.Run(t, new(TestBackendSuite))
}

func (suite *TestBackendSuite) TestCRUD() {
	// 创建 MBangumi 记录
	bangumi := db.MBangumi{
		Title:      "Bangumi 1",
		AliasNames: "Alias 1, Alias 2",
		TMDBId:     "tmdb001",
		Seasons: []db.MSeason{
			{
				Number:     1,
				EpCount:    12,
				SubjectId:  123,
				MikanId:    "mikan001",
				SeasonType: "Type 1",
				State:      db.SeasonStateIncomplete,
			},
			{
				Number:     2,
				EpCount:    12,
				SubjectId:  456,
				MikanId:    "mikan002",
				SeasonType: "Type 2",
				State:      db.SeasonStateComplete,
			},
		},
	}
	err := suite.backend.AddOrUpdateBangumi(&bangumi)
	suite.NoError(err)
	err = suite.backend.ListBangumis(func(bgm *db.MBangumi) bool {
		suite.T().Log(bgm.Title)
		for _, season := range bgm.Seasons {
			suite.T().Log(season.Number)
		}
		return false
	})
	suite.NoError(err)

	var incomplete db.MBangumi
	err = suite.backend.ListIncompleteBangumi(func(bgm *db.MBangumi) bool {
		incomplete = *bgm
		return false
	})
	suite.NoError(err)
	suite.T().Log(incomplete.Title)
}

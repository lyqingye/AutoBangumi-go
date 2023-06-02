package mikan_test

import (
	"regexp"
	"strings"
	"testing"

	"autobangumi-go/mdb"
	"autobangumi-go/rss/mikan"
	"autobangumi-go/rss/mikan/cache"
	"github.com/stretchr/testify/suite"
)

type TestMikanParserSuite struct {
	suite.Suite
	parser *mikan.MikanRSSParser
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
	cm := cache.NewInMemoryCacheManager()
	parser, err := mikan.NewMikanRSSParser("https://mikanani.me", tmdbClient, bangumiTVClient, cm)
	if err != nil {
		return nil, err
	}
	return parser, nil
}

func (suite *TestMikanParserSuite) SetupTest() {
	parser, err := NewTestMikanParser()
	suite.NoError(err)
	suite.parser = parser
}

func TestMikanParser(t *testing.T) {
	suite.Run(t, new(TestMikanParserSuite))
}

func (suite *TestMikanParserSuite) TestSearchBangumi() {
	rs, err := suite.parser.Search2("因想当冒险者而前往大都市的女儿已经升到了S级")
	suite.NoError(err)
	suite.NotNil(rs)
}

func (suite *TestMikanParserSuite) TestSearchBangumiByID() {
	result, err := suite.parser.Search3("2984")
	suite.NoError(err)
	suite.NotNil(result)
}

func (suite *TestMikanParserSuite) TestSearchByTitle() {
	rs, err := suite.parser.Search("式守同学不只可爱而已", 0)
	suite.NoError(err)
	suite.NotNil(rs)
}

func TestNormalizationSearchTitle(t *testing.T) {
	t.Log(normalizationSearchTitle("总之就是非常可爱 第二季"))
	t.Log(normalizationSearchTitle("总之就是非常可爱 第三季"))
	t.Log(normalizationSearchTitle("总之就是非常可爱 第三期"))
	t.Log(normalizationSearchTitle("总之就是非常可爱 SeasonNum 3"))
	t.Log(normalizationSearchTitle("总之就是非常可爱 Season3"))
}

func normalizationSearchTitle(keyword string) string {
	patterns := []string{
		"第([[:digit:]]+|\\p{Han}+)季",
		"第([[:digit:]]+|\\p{Han}+)期",
		"SeasonNum\\s*\\d+",
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		keyword = strings.ReplaceAll(keyword, re.FindString(keyword), "")
	}
	return keyword
}

package mdb_test

import (
	"testing"

	"autobangumi-go/mdb"
	"github.com/stretchr/testify/suite"
)

type TestBangumiTVSuite struct {
	suite.Suite
	cli *mdb.BangumiTVClient
}

func TestBangumiTV(t *testing.T) {
	suite.Run(t, new(TestBangumiTVSuite))
}

func (s *TestBangumiTVSuite) SetupSuite() {
	cli, err := mdb.NewBangumiTVClient("https://api.bgm.tv/v0")
	s.Require().NoError(err)
	s.Require().NotNil(cli)
	s.cli = cli
}

func (s *TestBangumiTVSuite) TestBangumiTxSubjects() {
	subject, err := s.cli.GetSubjects(404804)
	s.Require().NoError(err)
	s.Require().NotNil(subject)
}

func (s *TestBangumiTVSuite) TestSearchSubject() {
	for _, keyword := range []string{"异世界舅舅", "我的青春恋爱物语果然有问题。续", "赤发的白雪姬"} {
		subject, err := s.cli.SearchAnime(keyword)
		s.Require().NoError(err)
		s.Require().NotNil(subject)
		s.T().Log(subject.GetAliasNames())
	}
}

func (s *TestBangumiTVSuite) TestMe() {
	meInfo, err := s.cli.Me()
	s.Require().NoError(err)
	s.Require().NotNil(meInfo)
}

func (s *TestBangumiTVSuite) TestGetCalendar() {
	calendar, err := s.cli.GetCalendar()
	s.Require().NoError(err)
	s.Require().NotNil(calendar)
}

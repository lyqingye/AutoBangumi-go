package jellyfin_test

import (
	"testing"

	"autobangumi-go/jellyfin"
	"github.com/stretchr/testify/suite"
)

type TestJellyfinSutie struct {
	suite.Suite
	client *jellyfin.Client
}

func TestRunJellyfinTest(t *testing.T) {
	suite.Run(t, new(TestJellyfinSutie))
}

func (suite *TestJellyfinSutie) SetupSuite() {
	client, err := jellyfin.NewClient("http://nas.lyqingye.com:8096", "autobangumi", "123456")
	suite.Require().NoError(err)
	suite.Require().NotNil(client)
	suite.client = client
}

func (suite *TestJellyfinSutie) TestScanLibrary() {
	err := suite.client.StartLibraryScan()
	suite.Require().NoError(err)
}

func (suite *TestJellyfinSutie) TestLogin() {
	resp, err := suite.client.Login("lyqingye", "bengbuzhule")
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
}

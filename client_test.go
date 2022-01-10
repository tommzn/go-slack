package slack

import (
	"github.com/stretchr/testify/suite"
	"net/http"
	"os"
	"testing"

	secrets "github.com/tommzn/go-secrets"
)

type ClientTestSuite struct {
	suite.Suite
	tokenKey string
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) SetupSuite() {
	suite.tokenKey = "AuthToken"
	if _, ok := os.LookupEnv(SLACK_TOKEN); !ok {
		os.Setenv(SLACK_TOKEN, suite.tokenKey)
	}
}

func (suite *ClientTestSuite) TearDownSuite() {
	if token, ok := os.LookupEnv(SLACK_TOKEN); ok && token == suite.tokenKey {
		os.Unsetenv(SLACK_TOKEN)
	}
}

func (suite *ClientTestSuite) TestSendMessage() {

	// Test with token provided from env
	client1 := suite.clientForTest(false)
	suite.Nil(client1.SendToChannel("Hello!", "ChannelId", nil))

	// Test with token provided from secrets manager
	client2 := suite.clientWithSecretsManagerForTest()
	suite.Nil(client2.SendToChannel("Hello!", "ChannelId", nil))
}

func (suite *ClientTestSuite) TestSendToDefaultChannel() {

	client := suite.clientForTest(false)
	header := "This is a header"

	suite.NotNil(client.Send("Hello!", &header))
	client.WithChannel("chn01")
	suite.Nil(client.Send("Hello!", &header))
}

func (suite *ClientTestSuite) TestWithResponseError() {

	client := suite.clientForTest(true)
	suite.NotNil(client.SendToChannel("Hello!", "ChannelId", nil))
}

func (suite *ClientTestSuite) TestWithChannelFromConfig() {

	client := suite.clientWithConfigForTest()
	suite.NotNil(client.channel)
	suite.Nil(client.Send("Hello!", nil))
}

func (suite *ClientTestSuite) clientForTest(shouldReturnWithError bool) *Client {
	client := New()
	client.httpClient = &http.Client{Transport: newRoundTripper(nil, newHttpMock(shouldReturnWithError, suite.Assert()))}
	return client
}

func (suite *ClientTestSuite) clientWithSecretsManagerForTest() *Client {
	client := NewFromConfig(nil, secrets.NewSecretsManager())
	client.httpClient = &http.Client{Transport: newRoundTripper(nil, newHttpMock(false, suite.Assert()))}
	return client
}

func (suite *ClientTestSuite) clientWithConfigForTest() *Client {
	client := NewFromConfig(loadConfigForTest(nil), nil)
	client.httpClient = &http.Client{Transport: newRoundTripper(nil, newHttpMock(false, suite.Assert()))}
	return client
}

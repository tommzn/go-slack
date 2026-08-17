package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
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

// TestTokenNotFound exercises the token() error path when SLACK_TOKEN is absent
// and no secrets manager is configured.
func (suite *ClientTestSuite) TestTokenNotFound() {

	prev, hadToken := os.LookupEnv(SLACK_TOKEN)
	os.Unsetenv(SLACK_TOKEN)
	suite.T().Cleanup(func() {
		if hadToken {
			os.Setenv(SLACK_TOKEN, prev)
		}
	})

	client := New()
	err := client.SendToChannel("Hello!", "ChannelId", nil)
	suite.NotNil(err)
	suite.Contains(err.Error(), "auth token")
}

// TestEvalResponseOkFalseNoErrorField exercises the evalResponse path where
// the Slack API returns ok:false but omits the error field.
func (suite *ClientTestSuite) TestEvalResponseOkFalseNoErrorField() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(postMessageResponse{Ok: false}) // Error field absent
	}))
	defer ts.Close()

	// Build a client that injects auth via env token and forwards to the test server.
	client := New()
	client.httpClient = &http.Client{
		Transport: newRoundTripper(nil, &forwardToServer{base: ts.URL}),
	}

	err := client.SendToChannel("Hello!", "ChannelId", nil)
	suite.NotNil(err)
	suite.Contains(err.Error(), "error has occurred")
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

// forwardToServer is a simple RoundTripper that redirects requests to a test server URL.
type forwardToServer struct {
	base string
}

func (f *forwardToServer) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Scheme = "http"
	r.URL.Host = r.Host
	req, err := http.NewRequest(r.Method, f.base+r.URL.Path, r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header
	return http.DefaultTransport.RoundTrip(req)
}

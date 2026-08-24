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

func (suite *ClientTestSuite) SetupTest() {
	suite.tokenKey = "AuthToken"
	// Set SLACK_TOKEN for the current test; Go's testing cleanup (via t.Setenv) restores it after each test.
	suite.T().Setenv(SLACK_TOKEN, suite.tokenKey)
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

	// SetupTest already sets SLACK_TOKEN via t.Setenv; explicitly unsetting here
	// only affects this test — t.Setenv's cleanup restores the previous value.
	os.Unsetenv(SLACK_TOKEN)

	client := New()
	err := client.SendToChannel("Hello!", "ChannelId", nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "auth token")
}

// TestEvalResponseOkFalseNoErrorField exercises the evalResponse path where
// the Slack API returns ok:false but omits the error field.
func (suite *ClientTestSuite) TestEvalResponseOkFalseNoErrorField() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(postMessageResponse{Ok: false}) // Error field absent
	}))
	defer ts.Close()

	// Build a client that injects auth via env token and forwards to the test server.
	client := New()
	client.httpClient = &http.Client{
		Transport: newRoundTripper(nil, &forwardToServer{base: ts.URL}),
	}

	err := client.SendToChannel("Hello!", "ChannelId", nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "unspecified error")
}

// TestEvalResponseDecodeError exercises the evalResponse path where the Slack
// API returns a body that is not valid JSON.
func (suite *ClientTestSuite) TestEvalResponseDecodeError() {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	client := New()
	client.httpClient = &http.Client{
		Transport: newRoundTripper(nil, &forwardToServer{base: ts.URL}),
	}

	err := client.SendToChannel("Hello!", "ChannelId", nil)
	suite.Require().Error(err)
}

// TestSendNoChannelAssigned exercises the Send() guard when no channel has
// been configured on the client.
func (suite *ClientTestSuite) TestSendNoChannelAssigned() {

	client := New()
	err := client.Send("Hello!", nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "no channel assigned")
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

// forwardToServer is a simple RoundTripper that redirects requests to a test server URL
// without mutating the original request (honoring the http.RoundTripper contract).
type forwardToServer struct {
	base string
}

func (f *forwardToServer) RoundTrip(r *http.Request) (*http.Response, error) {
	// Build a new target URL from the test-server base; preserve path and query string
	// without touching r.URL so the original request is not mutated.
	target := f.base + r.URL.Path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	req, err := http.NewRequest(r.Method, target, r.Body)
	if err != nil {
		return nil, err
	}
	// Clone the header map so the new request owns its own copy.
	req.Header = r.Header.Clone()
	return http.DefaultTransport.RoundTrip(req)
}

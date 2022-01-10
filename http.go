package slack

import (
	"errors"
	"net/http"
	"os"

	secrets "github.com/tommzn/go-secrets"
)

// NewRoundTripper returns a middleware to add auth token and content type for Slack api requests.
func newRoundTripper(secretsManager secrets.SecretsManager, roundTripper http.RoundTripper) http.RoundTripper {

	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	return &slackRequests{
		roundTripper:   roundTripper,
		secretsManager: secretsManager,
	}
}

// RoundTrip adds auth token and content type for requests.
func (slackReq *slackRequests) RoundTrip(r *http.Request) (*http.Response, error) {

	tokem, err := slackReq.token()
	if err != nil {
		return nil, err
	}
	r.Header.Add("Authorization", "Bearer: "+*tokem)
	r.Header.Add("Content-Type", "application/jso")
	return slackReq.roundTripper.RoundTrip(r)
}

// Token will try to get an auth token either from assigned secrets manager or directly from environement.
// In both cases SLACK_TOKEN is used as key.
func (slackReq *slackRequests) token() (*string, error) {
	if slackReq.secretsManager != nil {
		return slackReq.secretsManager.Obtain(SLACK_TOKEN)
	}
	if token, ok := os.LookupEnv(SLACK_TOKEN); ok {
		return &token, nil
	}
	return nil, errors.New("Unable to obtain auth token for Slack api request.")
}

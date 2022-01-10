package slack

import (
	"net/http"

	secrets "github.com/tommzn/go-secrets"
)

// SLACK_TOKEN is used to obtain an auth token from environment or if available by assigned secrets mananger.
const SLACK_TOKEN = "SLACK_TOKEN"

// SLACK_POST_MESSAGE_API is the endpoint provides by Slack's web api to send messages.
const SLACK_POST_MESSAGE_API = "https://slack.com/api/chat.postMessage"

// SlackRequests is a middleware to add auth token and content type for request send to Slack web api.
type slackRequests struct {

	// RoundTripper next or default round tripper.
	roundTripper http.RoundTripper

	// SecretsManager is used to obtain auth token.
	secretsManager secrets.SecretsManager
}

// Client is used to send messages to Slack using it's web api.
type Client struct {
	httpClient *http.Client
	channel    *string
}

// PostMessageBody is the request body send in post message requests to the Slack web api.
type postMessageBody struct {
	Channel string             `json:"channel,omitempty"`
	Blocks  []postMessageBlock `json:"blocks"`
}

// PostMessageBlock is a single block, either a header or section, which contains text.
type postMessageBlock struct {
	Type string          `json:"type"`
	Text postMessageText `json:"text"`
}

// PostMessageText is a text used in a header or section.
type postMessageText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PostMessageResponse will be received from Slack api.
type postMessageResponse struct {
	Ok    bool    `json:"ok"`
	Error *string `json:"error,omitempty"`
}

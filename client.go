// Package slack provides a simple client to send messages to Slack channels.
package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	config "github.com/tommzn/go-config"
	secrets "github.com/tommzn/go-secrets"
)

// New will return a client to send messages to Slack.
// Auth token will be taken from environment using SLACK_TOKEN. For other token sources use secrets manager with NewFromConfig.
func New() *Client {
	return &Client{
		httpClient: &http.Client{Transport: newRoundTripper(nil, nil)},
	}
}

// NewFromConfig returns a client to send messages to Slack.
// You can pass a config to set a default channel and secrets managet to obtain auth token. Both are optional.
func NewFromConfig(conf config.Config, secretsManager secrets.SecretsManager) *Client {

	var channel *string
	if conf != nil {
		channel = conf.Get("slack.channel", nil)
	}
	return &Client{
		httpClient: &http.Client{Transport: newRoundTripper(secretsManager, nil)},
		channel:    channel,
	}
}

// WithChannel assign passed channel id to client. Useful if most messages should be send to one channel.
func (client *Client) WithChannel(channel string) *Client {
	client.channel = &channel
	return client
}

// Send will send passed message to previous assigned channel. Passed message is send as section block in plain text.
// Header is optional and will be send as a header block if passed.
func (client *Client) Send(message string, header *string) error {
	if client.channel == nil {
		return errors.New("No channel assigned!")
	}
	return client.SendToChannel(message, *client.channel, header)
}

// SendToChannel will send passed message to given channel. Message is send as section block in plain text.
// Header is optional and will be send as a header block if passed.
func (client *Client) SendToChannel(message, channel string, header *string) error {
	req, err := newRequest(message, channel, header)
	if err != nil {
		return err
	}
	return evalResponse(client.httpClient.Do(req))
}

// NewRequest creates a new HTTP POST request and encode passed message, channel and header as body.
func newRequest(message, channel string, header *string) (*http.Request, error) {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(newPostMessageBody(message, channel, header))
	if err != nil {
		return nil, err
	}
	return http.NewRequest(http.MethodPost, SLACK_POST_MESSAGE_API, &body)
}

// NewPostMessageBody converts passed message, channel and header to a request payload struct.
func newPostMessageBody(message, channel string, header *string) postMessageBody {
	postMessageBody := postMessageBody{
		Channel: channel,
		Blocks:  []postMessageBlock{newPostMessageBlock("section", message)},
	}
	if header != nil {
		postMessageBody.Blocks = append(postMessageBody.Blocks, newPostMessageBlock("header", *header))
	}
	return postMessageBody
}

// NewPostMessageBlock creates a new block for Slack web api request. Test type is plain text always.
func newPostMessageBlock(bloackType, text string) postMessageBlock {
	return postMessageBlock{
		Type: bloackType,
		Text: postMessageText{Type: "plain_text", Text: text},
	}
}

// EvalResponse will process send back from Slack web api. Can be used directly with return values from http.Client.Do.
func evalResponse(response *http.Response, err error) error {

	if err != nil {
		return err
	}

	defer response.Body.Close()
	var postMessageResponse postMessageResponse
	decodeErr := json.NewDecoder(response.Body).Decode(&postMessageResponse)
	if decodeErr != nil {
		return decodeErr
	}
	if !postMessageResponse.Ok {
		if postMessageResponse.Error != nil {
			return errors.New(*postMessageResponse.Error)
		} else {
			return errors.New("An error has occurred!")
		}
	}
	return nil
}

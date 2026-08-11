package slack

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/assert"
)

// HttpMock will catch request send from http client and returns predefined payloads for testing.
type httpMock struct {
	shouldReturnWithError bool
	assert                *assert.Assertions
}

// NewHttpMock returns a new nock for HTTP requests. It have to be added as Transport during creating a *http.Client.
// shouldReturnWithError can be used to return successful or a response error.
func newHttpMock(shouldReturnWithError bool, assert *assert.Assertions) *httpMock {
	return &httpMock{
		shouldReturnWithError: shouldReturnWithError,
		assert:                assert,
	}
}

// Do will look for expected headers in passed request and returns with predefined response payload.
func (mock *httpMock) Do(req *http.Request) (*http.Response, error) {
	expectedHeaders := []string{"Authorization", "Content-Type"}
	for _, header := range expectedHeaders {
		if _, ok := req.Header[header]; !ok {
			return nil, errors.New("Missing header: " + header)
		}
	}
	// Validate header values
	mock.assert.Equal("application/json", req.Header.Get("Content-Type"), "Content-Type must be application/json")
	authHeader := req.Header.Get("Authorization")
	mock.assert.True(strings.HasPrefix(authHeader, "Bearer "), "Authorization must use RFC 6750 'Bearer <token>' format (no colon), got: "+authHeader)
	return mock.createResponse()
}

// RoundTrip to conform interface defined by Tranport in http package.
func (mock *httpMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return mock.Do(req)
}

// CreateResponse will returns a mocked response which looks like a response from Slack web api.
// Ue shouldReturnWithError when creating this mock to define successful or error response.
func (mock *httpMock) createResponse() (*http.Response, error) {
	w := httptest.NewRecorder()
	if mock.shouldReturnWithError {
		msg := "An error has occurred!"
		json.NewEncoder(w).Encode(postMessageResponse{Ok: false, Error: &msg})
	} else {
		json.NewEncoder(w).Encode(postMessageResponse{Ok: true})
	}
	return w.Result(), nil
}

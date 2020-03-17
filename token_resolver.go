package gcs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type defaultResolver struct {
	client  httpClient
	context context.Context
}

func NewTokenResolver(options ...ResolverOption) TokenResolver {
	this := &defaultResolver{}

	WithResolverClient(defaultHTTPClient())(this)
	WithResolverContext(context.Background())(this)
	for _, option := range options {
		option(this)
	}

	return this
}

func (this *defaultResolver) AccessToken(identity ClientIdentity) (AccessToken, error) {
	request, _ := http.NewRequest("POST", tokenURL, this.generateRequestBody(identity))
	request = request.WithContext(this.context)
	response, err := this.client.Do(request)
	return this.processResponse(response, err)
}
func (this *defaultResolver) generateRequestBody(identity ClientIdentity) io.Reader {
	raw, _ := json.Marshal(struct {
		ClientIdentity
		GrantType string `json:"grant_type"`
	}{ClientIdentity: identity, GrantType: "refresh_token"})
	return bytes.NewBuffer(raw)
}
func (this *defaultResolver) processResponse(response *http.Response, err error) (AccessToken, error) {
	if err != nil {
		return emptyToken, err
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return emptyToken, ErrFailedTokenRequest
	}

	return this.unmarshalBody(response.Body)
}
func (this *defaultResolver) unmarshalBody(source io.Reader) (result AccessToken, err error) {
	err = json.NewDecoder(source).Decode(&result)
	return result, err
}

const tokenURL = "https://www.googleapis.com/oauth2/v4/token"

var emptyToken = AccessToken{}
var ErrFailedTokenRequest = errors.New("unable to resolve access token")

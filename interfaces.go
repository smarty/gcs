package gcs

import "net/http"

type (
	TokenResolver interface {
		AccessToken(ClientIdentity) (AccessToken, error)
	}
	ClientIdentity struct {
		ID           string `json:"client_id"`
		Secret       string `json:"client_secret"`
		RefreshToken string `json:"refresh_token"`
	}
	AccessToken struct {
		Value      string `json:"access_token"`
		Expiration uint16 `json:"expires_in"`
		Scope      string `json:"scope"`
		Type       string `json:"token_type"`
		ID         string `json:"id_token"`
	}
)

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

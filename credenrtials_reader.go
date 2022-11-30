package gcs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type CredentialsReader interface {
	Read(context.Context, string) (Credentials, error)
}
type environmentReader interface {
	LookupEnv(string) (string, bool)
}
type fileReader interface {
	ReadFile(string) ([]byte, error)
}

var ErrCredentialsFailure = errors.New("unable to discover credentials")

func NewCredentialsReader(options ...credentialOption) CredentialsReader {
	var config credentialConfig
	CredentialOptions.apply(options...)(&config)
	return &defaultReader{
		client:            config.client,
		fileReader:        config.fileReader,
		environmentReader: config.environmentReader,
		vaultAddress:      config.vaultAddress,
		vaultToken:        config.vaultToken,
		vaultKey:          config.vaultKey,
	}
}

type defaultReader struct {
	client            httpClient
	fileReader        fileReader
	environmentReader environmentReader
	vaultAddress      string
	vaultToken        string
	vaultKey          string
}

func (this *defaultReader) Read(ctx context.Context, value string) (Credentials, error) {
	if value = sanitizeToken(value); len(value) > 0 {
		return Credentials{BearerToken: value}, nil
	}

	if read, found := this.environmentReader.LookupEnv("GOOGLE_OAUTH_ACCESS_TOKEN"); found {
		return Credentials{BearerToken: sanitizeToken(read)}, nil
	}

	if read, found := this.environmentReader.LookupEnv("GOOGLE_CREDENTIALS"); found {
		if raw, err := base64.StdEncoding.DecodeString(read); err != nil {
			return Credentials{}, fmt.Errorf("unable to base64 decode value from environment variable [GOOGLE_CREDENTIALS]: %w", err)
		} else {
			return ParseCredentialsFromJSON(raw, WithResolverClient(this.client), WithResolverContext(ctx))
		}
	}

	if read, found := this.environmentReader.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); found {
		if raw, err := this.fileReader.ReadFile(read); err != nil {
			return Credentials{}, fmt.Errorf("unable to read file specified in [GOOGLE_APPLICATION_CREDENTIALS]: %w", err)
		} else {
			return ParseCredentialsFromJSON(raw, WithResolverClient(this.client), WithResolverContext(ctx))
		}
	}

	if len(this.vaultAddress) > 0 && len(this.vaultToken) > 0 && len(this.vaultKey) > 0 {
		return this.resolveGoogleAccessToken(ctx)
	}

	return Credentials{}, ErrCredentialsFailure

}
func (this *defaultReader) resolveGoogleAccessToken(ctx context.Context) (Credentials, error) {
	parsed, err := url.Parse(this.vaultAddress)
	if err != nil {
		return Credentials{}, fmt.Errorf("unable to parse value specified in VAULT_ADDR: %w", err)
	} else if len(parsed.Scheme) == 0 || len(parsed.Host) == 0 {
		return Credentials{}, fmt.Errorf("unable to parse value specified in VAULT_ADDR")
	}

	request, _ := http.NewRequest("GET", parsed.Scheme+"://"+parsed.Host+"/v1/"+this.vaultKey, nil)
	request.Header["X-Vault-Token"] = []string{this.vaultToken}

	response, err := this.client.Do(request.WithContext(ctx))
	if err != nil {
		return Credentials{}, fmt.Errorf("unable to connect to the configured Vault server: %w", err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return Credentials{}, fmt.Errorf("the Vault token provided did not have permission to read from [%s]", this.vaultKey)
	} else if response.StatusCode != http.StatusOK {
		return Credentials{}, fmt.Errorf("unexpected status from the Vault server [%d]", response.StatusCode)
	}

	if strings.ToLower(response.Header.Get("Content-Type")) != "application/json" {
		return Credentials{}, fmt.Errorf("unknown Content-Type from the Vault server [%s]", response.Header.Get("Content-Type"))
	}

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return Credentials{}, fmt.Errorf("unable to read response from the Vault server: %w", err)
	} else if len(raw) == 0 {
		return Credentials{}, fmt.Errorf("zero-length response returned from the Vault server")
	}

	body := struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}{}

	err = json.Unmarshal(raw, &body)
	if err != nil {
		return Credentials{}, fmt.Errorf("unable to parse response body returned from the Vault server: %w", err)
	}

	accessToken := strings.Trim(body.Data.Token, ". ") // eliminate spaces and periods from the end (and beginning)
	if len(accessToken) == 0 {
		return Credentials{}, fmt.Errorf("no access token was returned from the Vault server: %w", err)
	}

	return Credentials{BearerToken: "Bearer " + accessToken}, nil
}
func sanitizeToken(value string) string {
	if value = strings.TrimSpace(value); len(value) == 0 {
		return ""
	} else if !strings.HasPrefix(value, "Bearer") {
		value = "Bearer " + value
	}

	return strings.TrimSuffix(value, ".")
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type credentialConfig struct {
	client            httpClient
	fileReader        fileReader
	environmentReader environmentReader
	vaultAddress      string
	vaultToken        string
	vaultKey          string
}
type credentialSingleton struct{}
type credentialOption func(*credentialConfig)

var CredentialOptions credentialSingleton

func (credentialSingleton) HTTPClient(value httpClient) credentialOption {
	return func(this *credentialConfig) { this.client = value }
}
func (credentialSingleton) FileReader(value fileReader) credentialOption {
	return func(this *credentialConfig) { this.fileReader = value }
}
func (credentialSingleton) EnvironmentReader(value environmentReader) credentialOption {
	return func(this *credentialConfig) { this.environmentReader = value }
}
func (credentialSingleton) VaultServer(address, token string) credentialOption {
	return func(this *credentialConfig) { this.vaultAddress = address; this.vaultToken = token }
}
func (credentialSingleton) VaultKey(value string) credentialOption {
	return func(this *credentialConfig) { this.vaultKey = strings.TrimLeft(value, "/") }
}
func (credentialSingleton) apply(options ...credentialOption) credentialOption {
	return func(this *credentialConfig) {
		for _, option := range CredentialOptions.defaults(options...) {
			option(this)
		}
	}
}
func (credentialSingleton) defaults(options ...credentialOption) []credentialOption {
	return append([]credentialOption{
		CredentialOptions.HTTPClient(defaultHTTPClient()),
		CredentialOptions.FileReader(&externalSystem{}),
		CredentialOptions.EnvironmentReader(&externalSystem{}),
		CredentialOptions.VaultServer(os.Getenv("VAULT_ADDR"), os.Getenv("VAULT_TOKEN")),
		CredentialOptions.VaultKey(os.Getenv("VAULT_KEY")),
	}, options...)
}

type externalSystem struct{}

func (*externalSystem) ReadFile(value string) ([]byte, error) { return os.ReadFile(value) }
func (*externalSystem) LookupEnv(value string) (string, bool) { return os.LookupEnv(value) }

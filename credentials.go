package gcs

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
)

func ParseCredentialsFromJSON(raw []byte, options ...ResolverOption) (Credentials, error) {
	parsed, err := unmarshalClientCredentials(raw)
	if err != nil {
		return Credentials{}, err
	}

	user := parsed.ClientIdentity()
	if len(user.RefreshToken) == 0 {
		return NewCredentials(parsed.ServiceAccount())
	}

	resolver := NewTokenResolver(options...)
	accessToken, err := resolver.AccessToken(user)
	if err != nil {
		return Credentials{}, err
	}

	bearerToken := fmt.Sprintf("%s %s", accessToken.Type, accessToken.Value)
	return Credentials{BearerToken: bearerToken}, nil
}

/* ////////////////////////////////////////////////////////////////////////////////////////////////////////////////// */

type clientSecrets struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`

	ClientEmail   string `json:"client_email"`
	PrivateKeyPEM string `json:"private_key"`
}

func unmarshalClientCredentials(raw []byte) (result clientSecrets, err error) {
	if err = json.Unmarshal(raw, &result); err != nil {
		return clientSecrets{}, ErrMalformedJSON
	}
	return result, err
}

func (this clientSecrets) ClientIdentity() ClientIdentity {
	return ClientIdentity{
		ID:           this.ClientID,
		Secret:       this.ClientSecret,
		RefreshToken: this.RefreshToken,
	}
}
func (this clientSecrets) ServiceAccount() (string, []byte) {
	return this.ClientEmail, []byte(this.PrivateKeyPEM)
}

/* ////////////////////////////////////////////////////////////////////////////////////////////////////////////////// */

type Credentials struct {
	BearerToken string

	AccessID   string
	PrivateKey PrivateKey
}

func NewCredentials(accessID string, privateKey []byte) (Credentials, error) {
	if parsed, err := newPrivateKey(privateKey); err != nil {
		return Credentials{}, err
	} else {
		return Credentials{AccessID: accessID, PrivateKey: parsed}, nil
	}
}

/* ////////////////////////////////////////////////////////////////////////////////////////////////////////////////// */

type PrivateKey struct {
	inner  *rsa.PrivateKey
	random io.Reader
}

func newPrivateKey(raw []byte) (PrivateKey, error) {
	if parsed, err := tryReadPrivateKey(raw); err != nil {
		return PrivateKey{}, ErrMalformedPrivateKey
	} else {
		return PrivateKey{inner: parsed, random: rand.Reader}, err
	}
}
func tryReadPrivateKey(key []byte) (*rsa.PrivateKey, error) {
	if decoded, _ := pem.Decode(key); decoded != nil {
		key = decoded.Bytes
	}

	return tryParsePKCS8(key)
}
func tryParsePKCS8(key []byte) (*rsa.PrivateKey, error) {
	if parsed, err := x509.ParsePKCS8PrivateKey(key); err != nil {
		return nil, ErrMalformedPrivateKey
	} else if parsed, ok := parsed.(*rsa.PrivateKey); !ok {
		return nil, ErrUnsupportedPrivateKey // e.g. ecdsa.PrivateKey
	} else {
		return parsed, nil
	}
}

func (this *PrivateKey) Sign(raw []byte) ([]byte, error) {
	if this.inner == nil {
		return nil, nil // no private key to sign with
	}

	sum := sha256.Sum256(raw)
	return rsa.SignPKCS1v15(this.random, this.inner, crypto.SHA256, sum[:])
}

var ErrMalformedPrivateKey = errors.New("malformed private key")
var ErrUnsupportedPrivateKey = errors.New("unsupported private key type")
var ErrMalformedJSON = errors.New("malformed JSON")

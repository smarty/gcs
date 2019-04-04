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
	"io"
)

func ParseCredentialsFromJSON(raw []byte) (Credentials, error) {
	type serviceAccount struct {
		PrivateKeyPEM string `json:"private_key"`
		ClientEmail   string `json:"client_email"`
	}

	parsed := serviceAccount{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Credentials{}, ErrMalformedJSON
	} else {
		return NewCredentials(parsed.ClientEmail, []byte(parsed.PrivateKeyPEM))
	}
}

/* ////////////////////////////////////////////////////////////////////////////////////////////////////////////////// */

type Credentials struct {
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

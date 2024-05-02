package gcs

import (
	"errors"
	"net/http"
)

func NewRequest(method string, options ...Option) (*http.Request, error) {
	input := newModel(method, options)
	if err := input.validate(); err != nil {
		return nil, err
	} else {
		return input.buildRequest()
	}
}

const (
	GET  = "GET"
	PUT  = "PUT"
	HEAD = "HEAD"
)

var (
	ErrHTTPMethodMissing      = errors.New("missing HTTP method")
	ErrHTTPMethodUnrecognized = errors.New("unrecognized HTTP method")
	ErrBucketMissing          = errors.New("bucket is required")
	ErrResourceMissing        = errors.New("object resource key is required")
	ErrContentMissing         = errors.New("content payload is required")
)

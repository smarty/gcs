package gcs

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"strconv"
	"strings"
	"time"
)

type Option func(*model)

func WithBearerToken(value string) Option {
	return func(this *model) { this.credentials = Credentials{BearerToken: value} }
}
func WithCredentials(credentials Credentials) Option {
	return func(this *model) { this.credentials = credentials }
}
func WithEndpoint(scheme, host string) Option {
	return func(this *model) { this.scheme = scheme; this.host = host }
}
func WithBucket(value string) Option {
	return func(this *model) { this.bucket = strings.TrimSpace(value) }
}
func WithResource(value string) Option {
	return func(this *model) { this.resource = strings.TrimSpace(value) }
}

// Deprecated
func WithExpiration(value time.Time) Option {
	return WithSignedExpiration(value)
}
func WithSignedExpiration(value time.Time) Option {
	return func(this *model) { this.epoch = strconv.FormatInt(value.Unix(), 10) }
}

func WithContext(value context.Context) Option {
	return func(this *model) { this.context = value }
}

func GetWithETag(value string) Option {
	return func(this *model) { this.etag = strings.TrimSpace(value) }
}
func PutWithGeneration(value string) Option {
	return func(this *model) { this.generation = strings.TrimSpace(value) }
}

func PutWithContentString(value string) Option {
	return func(this *model) { PutWithContentBytes([]byte(value))(this) }
}
func PutWithContentBytes(value []byte) Option {
	return func(this *model) {
		this.content = bytes.NewReader(value)
		this.contentLength = int64(len(value))
	}
}
func PutWithContent(value io.Reader) Option {
	return func(this *model) { this.content = value }
}

func PutWithContentType(value string) Option {
	return func(this *model) { this.contentType = strings.TrimSpace(value) }
}
func PutWithContentLength(value int64) Option {
	return func(this *model) { this.contentLength = value }
}
func PutWithContentMD5(value []byte) Option {
	return func(this *model) { this.contentMD5 = base64.StdEncoding.EncodeToString(value) }
}
func PutWithContentEncoding(value string) Option {
	return func(this *model) { this.contentEncoding = value }
}

func WithCompositeOption(options ...Option) Option {
	return func(this *model) { this.applyOptions(options) }
}
func WithConditionalOption(option Option, condition bool) Option {
	if condition {
		return option
	} else {
		return nil
	}
}

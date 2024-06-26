package gcs

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

type model struct {
	context         context.Context
	credentials     Credentials
	method          string
	host            string
	scheme          string
	bucket          string
	resource        string
	contentMD5      string
	contentType     string
	contentEncoding string
	generation      string
	etag            string
	encryption      bool
	contentLength   int64
	content         io.Reader

	// fields are computed during and after options are applied.
	objectKey string
	epoch     string
	targetURL *url.URL
}

func newModel(method string, options []Option) model {
	this := &model{method: method}

	WithEndpoint(defaultScheme, defaultHost)(this)
	WithSignedExpiration(defaultExpiration())(this)
	WithContext(context.Background())(this)

	this.applyOptions(options)
	this.objectKey = path.Join("/", this.bucket, this.resource)
	this.targetURL = &url.URL{Scheme: this.scheme, Host: this.host, Path: this.objectKey}

	return *this
}

func (this *model) applyOptions(options []Option) {
	for _, option := range options {
		if option != nil {
			option(this)
		}
	}
}
func (this *model) validate() error {
	if len(this.method) == 0 {
		return ErrHTTPMethodMissing
	} else if this.method != GET && this.method != PUT && this.method != HEAD {
		return ErrHTTPMethodUnrecognized
	} else if len(this.bucket) == 0 {
		return ErrBucketMissing
	} else if len(this.resource) == 0 {
		return ErrResourceMissing
	} else if this.method == PUT && this.content == nil {
		return ErrContentMissing
	}
	return nil
}

func (this *model) buildRequest() (request *http.Request, err error) {
	if request, err = http.NewRequest(this.method, this.targetURL.String(), this.content); err != nil {
		return nil, err
	}

	if err = this.authorizeRequest(request); err != nil {
		return nil, err
	}

	request.ContentLength = this.contentLength
	this.appendHeaders(request)
	return request.WithContext(this.context), nil
}
func (this *model) authorizeRequest(request *http.Request) error {
	if len(this.credentials.BearerToken) > 0 {
		request.Header.Set("Authorization", this.credentials.BearerToken)
		return nil
	}

	signature, err := this.calculateSignature()
	if err != nil {
		return err
	}

	request.URL = this.buildSignedURL(signature)
	return nil
}
func (this *model) calculateSignature() (string, error) {
	buffer := bytes.NewBuffer(nil)
	this.appendToBuffer(buffer)

	if signed, err := this.credentials.PrivateKey.Sign(buffer.Bytes()); err != nil {
		return "", err
	} else {
		return base64.StdEncoding.EncodeToString(signed), nil
	}
}
func (this *model) appendToBuffer(buffer io.Writer) {
	// https://cloud.google.com/storage/docs/access-control/signed-urls
	// https://cloud.google.com/storage/docs/access-control/signing-urls-manually
	appendTo(buffer, "%s\n%s\n%s\n%s\n", this.method, this.contentMD5, this.contentType, this.epoch)
	appendIf(len(this.generation) > 0 && this.method == PUT, buffer, "%s:%s\n", headerGeneration, this.generation)
	appendTo(buffer, "%s", this.objectKey)
}
func appendIf(condition bool, writer io.Writer, format string, values ...interface{}) {
	if condition {
		appendTo(writer, format, values...)
	}
}
func appendTo(writer io.Writer, format string, values ...interface{}) {
	_, _ = fmt.Fprintf(writer, format, values...)
}

func (this *model) buildSignedURL(signature string) *url.URL {
	query := this.targetURL.Query()
	query.Set(queryAccessID, this.credentials.AccessID)
	query.Set(queryExpires, this.epoch)
	query.Set(querySignature, signature)

	target := *this.targetURL
	target.RawQuery = query.Encode()
	return &target
}
func (this *model) appendHeaders(request *http.Request) {
	headers := request.Header

	if this.method == GET {
		tryAppendHeaders(len(this.etag) > 0, headers, headerIfNoneMatch, this.etag)
	} else if this.method == PUT {
		tryAppendHeaders(len(this.contentType) > 0, headers, headerContentType, this.contentType)
		tryAppendHeaders(len(this.contentMD5) > 0, headers, headerContentMD5, this.contentMD5)
		tryAppendHeaders(len(this.contentEncoding) > 0, headers, headerContentEncoding, this.contentEncoding)
		tryAppendHeaders(len(this.generation) > 0, headers, headerGeneration, this.generation)
	}
}
func tryAppendHeaders(condition bool, headers http.Header, name, value string) {
	if condition {
		headers.Set(name, value)
	}
}

func defaultExpiration() time.Time { return time.Now().UTC().Add(defaultExpireTime) }

const (
	defaultScheme         = "https"
	defaultHost           = "storage.googleapis.com"
	headerContentType     = "Content-Type"
	headerContentMD5      = "Content-MD5"
	headerContentEncoding = "Content-Encoding"
	headerIfNoneMatch     = "If-None-Match"
	headerGeneration      = "x-goog-if-generation-match"
	queryAccessID         = "GoogleAccessId"
	queryExpires          = "Expires"
	querySignature        = "Signature"
)

var defaultExpireTime = time.Second * 30

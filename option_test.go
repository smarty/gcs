package gcs

import (
	"context"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/smarty/gcs/internal/should"
)

func TestMissingMethod(t *testing.T) {
	request, err := NewRequest("")

	should.So(t, err, should.Equal, ErrHTTPMethodMissing)
	should.So(t, request, should.BeNil)
}

func TestUnrecognizedMethod(t *testing.T) {
	request, err := NewRequest("POST")

	should.So(t, err, should.Equal, ErrHTTPMethodUnrecognized)
	should.So(t, request, should.BeNil)
}

func Test_MissingBucket(t *testing.T) {
	request, err := NewRequest(GET, WithResource("file.txt"))

	should.So(t, err, should.Equal, ErrBucketMissing)
	should.So(t, request, should.BeNil)
}

func TestZeroLengthBucket(t *testing.T) {
	request, err := NewRequest(GET, WithBucket(""), WithResource("file.txt"))

	should.So(t, err, should.Equal, ErrBucketMissing)
	should.So(t, request, should.BeNil)
}

func Test_MissingResource(t *testing.T) {
	request, err := NewRequest(GET, WithBucket("bucket"))

	should.So(t, err, should.Equal, ErrResourceMissing)
	should.So(t, request, should.BeNil)
}

func TestZeroLengthResource(t *testing.T) {
	request, err := NewRequest(GET, WithBucket("bucket"), WithResource(""))

	should.So(t, err, should.Equal, ErrResourceMissing)
	should.So(t, request, should.BeNil)
}

func Test_MissingContent(t *testing.T) {
	request, err := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"))

	should.So(t, err, should.Equal, ErrContentMissing)
	should.So(t, request, should.BeNil)
}

func Test_Composite(t *testing.T) {
	option := WithCompositeOption(WithBucket("bucket"), WithResource("file.txt"))

	request, err := NewRequest(GET, option)

	should.So(t, err, should.BeNil)
	should.So(t, request.URL.Path, should.Equal, "/bucket/file.txt")
}

func Test_Conditional(t *testing.T) {
	option := WithCompositeOption(
		WithConditionalOption(WithBucket("bucket1"), false),
		WithConditionalOption(WithBucket("bucket2"), true),
		WithConditionalOption(WithBucket("bucket1"), false),
		WithConditionalOption(WithResource("file1.txt"), false),
		WithConditionalOption(WithResource("file2.txt"), true),
		WithConditionalOption(WithResource("file3.txt"), false))

	request, err := NewRequest(GET, option)

	should.So(t, err, should.BeNil)
	should.So(t, request.URL.Path, should.Equal, "/bucket2/file2.txt")
}

func Test_RequestContainsContext(t *testing.T) {
	ctx := context.Background()

	request, _ := NewRequest(GET, WithBucket("bucket"), WithResource("file.txt"), WithContext(ctx))
	should.So(t, request.Context(), should.Equal, ctx)
}

func TestRequestPathContainsBucketAndResource(t *testing.T) {
	request, err := NewRequest(GET, WithBucket("bucket"), WithResource("/directory/file.txt"))

	should.So(t, err, should.BeNil)
	should.So(t, request.URL.Path, should.Equal, "/bucket/directory/file.txt")
}

func TestEndpoint(t *testing.T) {
	request, err := NewRequest(GET, WithEndpoint("https", "localhost:9000"), WithBucket("bucket"), WithResource("file.txt"))

	should.So(t, err, should.BeNil)
	should.So(t, request.URL.Scheme, should.Equal, "https")
	should.So(t, request.URL.Host, should.Equal, "localhost:9000")
	should.So(t, request.URL.Path, should.Equal, "/bucket/file.txt")
}

func Test_SignedExpiration(t *testing.T) {
	expiration := time.Now().UTC()
	epoch := strconv.FormatInt(expiration.Unix(), 10)

	requestWithExpiration, _ := NewRequest(GET, WithBucket("bucket"), WithResource("file.txt"),
		WithSignedExpiration(expiration))

	should.So(t, requestWithExpiration.URL.Query().Get("Expires"), should.Equal, epoch)
}

func TestGET_Etag(t *testing.T) {
	request, _ := NewRequest(GET, WithBucket("bucket"), WithResource("file.txt"), GetWithETag("my-etag"))

	should.So(t, request.Header.Get("If-None-Match"), should.Equal, "my-etag")
}

func TestPUT_ContentType(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"), PutWithContentType("application/boink"))

	should.So(t, request.Header.Get("Content-Type"), should.Equal, "application/boink")
}

func TestPUT_ContentEncoding(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"), PutWithContentEncoding("utf-8"))

	should.So(t, request.Header.Get("Content-Encoding"), should.Equal, "utf-8")
}

func TestPUT_ContentMD5(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"), PutWithContentMD5([]byte{0, 1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 13, 14, 15}))

	should.So(t, request.Header.Get("Content-MD5"), should.Equal, "AAECAwQFBwgJCgsMDQ4P")
}

func TestPUT_ContentLength(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"))

	should.So(t, request.ContentLength, should.Equal, int64(2))
}

func TestPUT_ContentBytes(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentBytes([]byte("hi")))

	all, _ := io.ReadAll(request.Body)
	should.So(t, string(all), should.Equal, "hi")
	should.So(t, request.ContentLength, should.Equal, int64(2))
}

func TestPUT_ContentString(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"))

	all, _ := io.ReadAll(request.Body)
	should.So(t, string(all), should.Equal, "hi")
	should.So(t, request.ContentLength, should.Equal, int64(2))
}

func TestPUT_Content(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContent(strings.NewReader("hi")))

	all, _ := io.ReadAll(request.Body)
	should.So(t, string(all), should.Equal, "hi")
}

func TestPUT_ContentAndLength(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContent(strings.NewReader("hi")), PutWithContentLength(17))

	should.So(t, request.ContentLength, should.Equal, int64(17))
}

func TestPUT_Generation(t *testing.T) {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"), PutWithGeneration("42"))

	should.So(t, request.Header.Get("x-goog-if-generation-match"), should.Equal, "42")
}

func TestGET_WithCredentials(t *testing.T) {
	credentials, _ := ParseCredentialsFromJSON(sampleServiceAccountJSON)
	credentials.PrivateKey.random = nil // make it deterministic so the signature doesn't change between test runs
	frozen := time.Unix(1554410829, 0)  // fixed time so signature is deterministic

	request, _ := NewRequest(GET, WithBucket("bucket"), WithResource("file.txt"),
		WithCredentials(credentials), WithSignedExpiration(frozen))

	should.So(t, request.URL.Query().Get("Signature"), should.Equal, "PzVwB1N71A/p6wL7gP/Oh/nZdnsuoXQCszqFr/Q3jo6B5+ozZpuPIcuCW80+wtwUSBnKQJM4lcTVx6DtYrj2F3B/norqJPVdOHSCcG6bvGZ6oUjB2FQNzpQ1DyjY/mN0V8ziXe+FYPZzz6X0ewHJKaTHZb63BNQO92aMj/NMFYlN9FjdfdlE1G2La4oiT+Cjok47ncWw5UwhBXvJBEm8vgTtK2OU4AyqK+2vnOR/5PMBwTtU+82CmrnckOPeNZDyURiJvJenybIxrqOLzaaAsXvphQyz11XWt4Z8b+nqscQezuS6CcqKJiLDFvRcX0wXbzTxeOl00QWX3XGLaMUoGg==")
}

func TestPUT_WithCredentials(t *testing.T) {
	credentials, _ := ParseCredentialsFromJSON(sampleServiceAccountJSON)
	credentials.PrivateKey.random = nil // make it deterministic so the signature doesn't change between test runs
	frozen := time.Unix(1554410829, 0)  // fixed time so signature is deterministic

	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		WithCredentials(credentials),
		WithSignedExpiration(frozen),
		PutWithContentMD5([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5}),
		PutWithContentString("content"),
		PutWithContentType("content-type"),
		PutWithGeneration("12345678"))

	should.So(t, request.URL.Query().Get("Signature"), should.Equal, "VHcBMifvvm1Vg1rbaoXbOs3a2IbMBBx/LInfjRD/lxgA4njeFS7K1CIYHcTlVZNrJFB0vWo8/424wTcgh0WvMRHCsJgN0jm48jjRsASazKriGzO3Y86COcdbpG8Ifs0565ahC0cHY7+/U6TT7W4N11XNYEh6WU+MlMDrFAaPCCUOeHaUwcz6NAUDF5cZQdXAOYQrtFhi2ODGzZ9Y/rlUNiEdXWdIx46+gIWNkYXP6JsIRDHnZGAcZPUhzF6r6YyPMto/MhwKCjx4kxR/jSp2hDa8TAfVULXBTAlqxbWbTpDvht8XcZPx6/T/TnYcZHhKyIQIWCvQzIrrJLCX8rmVpA==")
}

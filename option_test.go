package gcs

import (
	"context"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/smarty/assertions/should"
	"github.com/smarty/gunit"
)

func TestOptionFixture(t *testing.T) {
	gunit.Run(new(OptionFixture), t)
}

type OptionFixture struct {
	*gunit.Fixture
}

func (this *OptionFixture) TestMissingMethod() {
	request, err := NewRequest("")

	this.So(err, should.Equal, ErrHTTPMethodMissing)
	this.So(request, should.BeNil)
}

func (this *OptionFixture) TestUnrecognizedMethod() {
	request, err := NewRequest("POST")

	this.So(err, should.Equal, ErrHTTPMethodUnrecognized)
	this.So(request, should.BeNil)
}

func (this *OptionFixture) Test_MissingBucket() {
	request, err := NewRequest(GET, WithResource("file.txt"))

	this.So(err, should.Equal, ErrBucketMissing)
	this.So(request, should.BeNil)
}

func (this *OptionFixture) TestZeroLengthBucket() {
	request, err := NewRequest(GET, WithBucket(""), WithResource("file.txt"))

	this.So(err, should.Equal, ErrBucketMissing)
	this.So(request, should.BeNil)
}

func (this *OptionFixture) Test_MissingResource() {
	request, err := NewRequest(GET, WithBucket("bucket"))

	this.So(err, should.Equal, ErrResourceMissing)
	this.So(request, should.BeNil)
}

func (this *OptionFixture) TestZeroLengthResource() {
	request, err := NewRequest(GET, WithBucket("bucket"), WithResource(""))

	this.So(err, should.Equal, ErrResourceMissing)
	this.So(request, should.BeNil)
}

func (this *OptionFixture) Test_MissingContent() {
	request, err := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"))

	this.So(err, should.Equal, ErrContentMissing)
	this.So(request, should.BeNil)
}

func (this *OptionFixture) Test_Composite() {
	option := WithCompositeOption(WithBucket("bucket"), WithResource("file.txt"))

	request, err := NewRequest(GET, option)

	this.So(err, should.BeNil)
	this.So(request.URL.Path, should.Equal, "/bucket/file.txt")
}

func (this *OptionFixture) Test_Conditional() {
	option := WithCompositeOption(
		WithConditionalOption(WithBucket("bucket1"), false),
		WithConditionalOption(WithBucket("bucket2"), true),
		WithConditionalOption(WithBucket("bucket1"), false),
		WithConditionalOption(WithResource("file1.txt"), false),
		WithConditionalOption(WithResource("file2.txt"), true),
		WithConditionalOption(WithResource("file3.txt"), false))

	request, err := NewRequest(GET, option)

	this.So(err, should.BeNil)
	this.So(request.URL.Path, should.Equal, "/bucket2/file2.txt")
}

func (this *OptionFixture) Test_RequestContainsContext() {
	ctx := context.Background()

	request, _ := NewRequest(GET, WithBucket("bucket"), WithResource("file.txt"), WithContext(ctx))
	this.So(request.Context(), should.Equal, ctx)
}

func (this *OptionFixture) TestRequestPathContainsBucketAndResource() {
	request, err := NewRequest(GET, WithBucket("bucket"), WithResource("/directory/file.txt"))

	this.So(err, should.BeNil)
	this.So(request.URL.Path, should.Equal, "/bucket/directory/file.txt")
}

func (this *OptionFixture) TestEndpoint() {
	request, err := NewRequest(GET, WithEndpoint("https", "localhost:9000"), WithBucket("bucket"), WithResource("file.txt"))

	this.So(err, should.BeNil)
	this.So(request.URL.Scheme, should.Equal, "https")
	this.So(request.URL.Host, should.Equal, "localhost:9000")
	this.So(request.URL.Path, should.Equal, "/bucket/file.txt")
}

func (this *OptionFixture) Test_SignedExpiration() {
	expiration := time.Now().UTC()
	epoch := strconv.FormatInt(expiration.Unix(), 10)

	requestWithExpiration, _ := NewRequest(GET, WithBucket("bucket"), WithResource("file.txt"),
		WithSignedExpiration(expiration))

	this.So(requestWithExpiration.URL.Query().Get("Expires"), should.Equal, epoch)
}

func (this *OptionFixture) TestGET_Etag() {
	request, _ := NewRequest(GET, WithBucket("bucket"), WithResource("file.txt"), GetWithETag("my-etag"))

	this.So(request.Header.Get("If-None-Match"), should.Equal, "my-etag")
}

func (this *OptionFixture) TestPUT_ContentType() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"), PutWithContentType("application/boink"))

	this.So(request.Header.Get("Content-Type"), should.Equal, "application/boink")
}

func (this *OptionFixture) TestPUT_ContentEncoding() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"), PutWithContentEncoding("utf-8"))

	this.So(request.Header.Get("Content-Encoding"), should.Equal, "utf-8")
}

func (this *OptionFixture) TestPUT_ContentMD5() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"), PutWithContentMD5([]byte{0, 1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 13, 14, 15}))

	this.So(request.Header.Get("Content-MD5"), should.Equal, "AAECAwQFBwgJCgsMDQ4P")
}

func (this *OptionFixture) TestPUT_ContentLength() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"))

	this.So(request.ContentLength, should.Equal, 2)
}

func (this *OptionFixture) TestPUT_ContentBytes() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentBytes([]byte("hi")))

	all, _ := ioutil.ReadAll(request.Body)
	this.So(string(all), should.Equal, "hi")
	this.So(request.ContentLength, should.Equal, 2)
}

func (this *OptionFixture) TestPUT_ContentString() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"))

	all, _ := ioutil.ReadAll(request.Body)
	this.So(string(all), should.Equal, "hi")
	this.So(request.ContentLength, should.Equal, 2)
}

func (this *OptionFixture) TestPUT_Content() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContent(strings.NewReader("hi")))

	all, _ := ioutil.ReadAll(request.Body)
	this.So(string(all), should.Equal, "hi")
}

func (this *OptionFixture) TestPUT_ContentAndLength() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContent(strings.NewReader("hi")), PutWithContentLength(17))

	this.So(request.ContentLength, should.Equal, 17)
}

func (this *OptionFixture) TestPUT_Generation() {
	request, _ := NewRequest(PUT, WithBucket("bucket"), WithResource("file.txt"),
		PutWithContentString("hi"), PutWithGeneration("42"))

	this.So(request.Header.Get("x-goog-if-generation-match"), should.Equal, "42")
}

func (this *OptionFixture) TestGET_WithCredentials() {
	credentials, _ := ParseCredentialsFromJSON(sampleServiceAccountJSON)
	credentials.PrivateKey.random = nil // make it deterministic so the signature doesn't change between test runs
	frozen := time.Unix(1554410829, 0)  // fixed time so signature is deterministic

	request, _ := NewRequest(GET, WithBucket("bucket"), WithResource("file.txt"),
		WithCredentials(credentials), WithSignedExpiration(frozen))

	this.So(request.URL.Query().Get("Signature"), should.Equal, "PzVwB1N71A/p6wL7gP/Oh/nZdnsuoXQCszqFr/Q3jo6B5+ozZpuPIcuCW80+wtwUSBnKQJM4lcTVx6DtYrj2F3B/norqJPVdOHSCcG6bvGZ6oUjB2FQNzpQ1DyjY/mN0V8ziXe+FYPZzz6X0ewHJKaTHZb63BNQO92aMj/NMFYlN9FjdfdlE1G2La4oiT+Cjok47ncWw5UwhBXvJBEm8vgTtK2OU4AyqK+2vnOR/5PMBwTtU+82CmrnckOPeNZDyURiJvJenybIxrqOLzaaAsXvphQyz11XWt4Z8b+nqscQezuS6CcqKJiLDFvRcX0wXbzTxeOl00QWX3XGLaMUoGg==")
}

func (this *OptionFixture) TestPUT_WithCredentials() {
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

	this.So(request.URL.Query().Get("Signature"), should.Equal, "VHcBMifvvm1Vg1rbaoXbOs3a2IbMBBx/LInfjRD/lxgA4njeFS7K1CIYHcTlVZNrJFB0vWo8/424wTcgh0WvMRHCsJgN0jm48jjRsASazKriGzO3Y86COcdbpG8Ifs0565ahC0cHY7+/U6TT7W4N11XNYEh6WU+MlMDrFAaPCCUOeHaUwcz6NAUDF5cZQdXAOYQrtFhi2ODGzZ9Y/rlUNiEdXWdIx46+gIWNkYXP6JsIRDHnZGAcZPUhzF6r6YyPMto/MhwKCjx4kxR/jSp2hDa8TAfVULXBTAlqxbWbTpDvht8XcZPx6/T/TnYcZHhKyIQIWCvQzIrrJLCX8rmVpA==")
}

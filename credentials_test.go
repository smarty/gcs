package gcs

import (
	"bytes"
	"testing"

	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
)

func TestCredentialsFixture(t *testing.T) {
	gunit.Run(new(CredentialsFixture), t)
}

type CredentialsFixture struct {
	*gunit.Fixture
}

func (this *CredentialsFixture) TestParsingCredentialFromJSON() {
	parsed, err := ParseCredentialsFromJSON(sampleJSON)

	this.So(err, should.BeNil)
	this.So(parsed.AccessID, should.Equal, "sample-key@project-id-here.iam.gserviceaccount.com")
}
func (this *CredentialsFixture) TestParsingCredentialsFromMalformedJSON() {
	malformedJSON := sampleJSON[0 : len(sampleJSON)/2]

	parsed, err := ParseCredentialsFromJSON(malformedJSON)

	this.So(err, should.Equal, ErrMalformedJSON)
	this.So(parsed.AccessID, should.Equal, "")
}
func (this *CredentialsFixture) TestMalformedPrivateKey() {
	malformedContents := bytes.ReplaceAll(sampleJSON, []byte("BEGIN PRIVATE KEY"), []byte{})

	parsed, err := ParseCredentialsFromJSON([]byte(malformedContents))

	this.So(err, should.Equal, ErrMalformedPrivateKey)
	this.So(parsed.AccessID, should.Equal, "")

}

var sampleJSON = []byte(`
{
  "type": "service_account",
  "project_id": "project-id-here",
  "private_key_id": "6c264308c88338cee18d31da2290c30e99e01839",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDq8w+G8ghbq177\nU33JJDciwLC3YxDFggo5cHU4FMa6okpFfHkLZVmzot7xs0xuwy5mMdGVRiCxSQ6S\nZQJ6H+0Eh3HZKBceuVZ9V8s0FmAiEQLfi7XqQq1bTFJVmcXGC5CTLek//j9wzfuM\nk3Dfbj3mP/PB/0dySDXBnG+ra1atFDuo+Nwb5yHzbXuLKxTagOlapq5cElxU11QI\nkX6OUynC/alYUdWkL2EeJkw6XNwCF6/fUJ3UtJjtvZH6hBbolUQqpyo7AGTkQIZU\nlRj3zPQfnvxT48MphBh/wjys3zc4WZe7IAlIqU4fzGX/iarA8NrxU3NO53+Bixmf\nsYn8pg4JAgMBAAECggEABAByVAHU1xze4Ah+xLmHb+HY0LLRAxAxNOO/t14ROXYQ\nrkjNtf46ribYxcSgSWWtKWPxVinM3kAojaFHTsWy74gQDhsS87zD5pwjc6Zq6kGg\naybRySTsF3lAEMGu/u3M/1jBl4uw0G7NuUn8mu2hg8W0lOoQcTeeJMdRlpmcLxP4\nCho6NCJlELXoK/vt6AmYL6i6C/39Rni/IH2YrdZUEb/aebojNAwVi2GN5VXzwovB\nQjNZjIXvfVCNbjDnEkP+7tEpNEzwR3tpne5q+rvwveimEfDAqvitIsc7LpoGuCiF\nknKleiA7b5nX6dD/kGT24mEyN1DMQzHnDnYaRTSQoQKBgQD5jfeojsSG5QHyEsy0\ne4uAt7ereOvTCKWOTzQHJ+gCfdmpYxj73Pxz4Zr8uoTkCdQ5UTf7OV9itayRnV7l\n3oPqa0214Vd5zuKFoNazYtuFopP4qCRRaj1qi4Tzkovb6Yiak/OnTbQE6PTRL0qs\n9+BbncMyaGTYfbAlxE+ludeRGQKBgQDxBIaMn78Vf9Ww064KJZ+kUjnmki95/4hh\n9uQ4v5KQWNL90WtSQOmEYQs35xTHaFmGRmcgbbKqJmXM/ktfq9nra7vWDr9uvg9C\nYuTGNqVL5I63n9zjvBngXJ6I1xTYbW9o1bUza1gAm/lZJgoKmH7NkR9liOEfTh+U\n/8DmT81ScQKBgGdPgXZzXCKoDa0kYUBaYP8xj0Tac25TBw6p9VT9DUxywzgfgUlL\nS+vBOwNjR/6LnyL3X6COONHJeh5yMsYg3yWdtHcWSbtwjVBarGdpBo4FJxLqsNZP\nkAtapPic829f96BenaDmRx89PZSX6mc+2s+yuQtWMmF5bwHDimGGVRqJAoGBAJJ7\nVKctA762VhLFZHZoTXFaRDR9TnuQMbyQiD5xOEugoIOA/wAb0ZESRfYw7LERG6//\nI/hSk47UDXUcbIT19lkdviioB/Lvcmi/oBlT5vyMKa0ybNbAYN26jOPQDKxJPrfx\ngtKAgBjGszJaaynras3XUMSt/1y+Z3VwRzXy9HARAoGBAMrtEldUlTrTMjkXiYFB\nR9PQHSfnhPGXgE+NORwI7HkwtAU9fZOKJ7GVYj3CA4fMARKwByDxbYwfCAVwW9ki\n2/8+xCdwQRALV0FhhQYKxC/Vhahv6NqGwuIdjVap0OGDDTZmQeGX7MSMPKrZ7U8l\nr1PtLnLKYCzCHY/ggnj7NaMO\n-----END PRIVATE KEY-----\n",
  "client_email": "sample-key@project-id-here.iam.gserviceaccount.com",
  "client_id": "116921672847611436020",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/sample-key%40project-id-here.iam.gserviceaccount.com"
}
`)

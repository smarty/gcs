package gcs

import (
	"context"
	"net/http"
)

type ResolverOption func(this *defaultResolver)

func WithResolverClient(value httpClient) ResolverOption {
	return func(this *defaultResolver) { this.client = value }
}
func WithResolverContext(value context.Context) ResolverOption {
	return func(this *defaultResolver) { this.context = value }
}

var defaultClient = &http.Client{} // TODO

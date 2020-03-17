package gcs

import (
	"net"
	"net/http"
	"time"
)

type ResolverOption func(this *defaultResolver)

func WithResolverClient(value httpClient) ResolverOption {
	return func(this *defaultResolver) { this.client = value }
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   1 * time.Second,
				KeepAlive: 1 * time.Second,
			}).DialContext,
			MaxIdleConns:          4,
			IdleConnTimeout:       32 * time.Second,
			TLSHandshakeTimeout:   16 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   -1,
			DisableKeepAlives:     true,
		},
	}
}

package destiny

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/taiidani/no-time-to-explain/internal/data"
)

type Client struct {
	cache  data.Cache
	client *http.Client
}

const (
	apiRootPath   = "https://www.bungie.net/Platform"
	assetRootPath = "https://www.bungie.net"
)

func NewTokenClient(cache data.Cache, token string) *Client {
	client := &http.Client{}
	client.Transport = &clientHttpRoundTripper{
		tripper:  http.DefaultTransport,
		apiToken: token,
	}

	return &Client{
		client: client,
		cache:  cache,
	}
}

type clientHttpRoundTripper struct {
	tripper  http.RoundTripper
	apiToken string
}

func (rt *clientHttpRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-API-Key", rt.apiToken)
	return rt.tripper.RoundTrip(req)
}

func (c *Client) lookupCacheItem(ctx context.Context, key string, obj any) bool {
	found, err := c.cache.Get(ctx, key, obj)
	if err != nil {
		slog.WarnContext(ctx, "cache lookup", "error", err.Error(), "key", key)
	}

	return found
}

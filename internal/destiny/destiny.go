package destiny

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

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

	maxAttempts := 10
	for {
		resp, err := rt.tripper.RoundTrip(req)
		if err != nil {
			return resp, err
		}

		switch resp.StatusCode {
		case http.StatusInternalServerError:
			_, _ = io.Copy(os.Stderr, resp.Body)
			fmt.Fprintln(os.Stderr)
			return resp, fmt.Errorf("500 currently having issues with the server")
		case http.StatusServiceUnavailable:
			// Possible rate limiting. Perform some retries
			if maxAttempts <= 0 {
				return resp, fmt.Errorf("request failure due to service unavailable. Try again later")
			}

			slog.Warn("Destiny API unavailable. Possible rate limiting. Trying again in 2 minutes")
			select {
			case <-time.After(time.Minute * 2):
				maxAttempts--
				continue
			case <-req.Context().Done():
				return nil, fmt.Errorf("request aborted: %w", req.Context().Err())
			}
		}

		return resp, err
	}
}

func (c *Client) lookupCacheItem(ctx context.Context, key string, obj any) bool {
	found, err := c.cache.Get(ctx, key, obj)
	if err != nil {
		slog.WarnContext(ctx, "cache lookup", "error", err.Error(), "key", key)
	}

	return found
}

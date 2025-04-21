package destiny

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/taiidani/go-bungie-api/api"
)

// GetManifest contains an index of all the localized assets for API resources.
// It can be used to extract the human-readable names for everything in the API.
// URL: https://bungie-net.github.io/multi/operation_get_Destiny2-GetDestinyManifest.html
func (c *Client) GetManifest(ctx context.Context) (*Manifest, error) {
	var cacheKey = "destiny:manifest:info"
	ret := &Manifest{}
	if found := c.lookupCacheItem(ctx, cacheKey, ret); found {
		return ret, nil
	}

	url := fmt.Sprintf("%s/Destiny2/Manifest", apiRootPath)
	slog.Info(url)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		_, _ = io.Copy(os.Stderr, resp.Body)
		fmt.Fprintln(os.Stderr)
		return nil, fmt.Errorf("500 currently having issues with the server")
	}

	type response struct {
		Response        *Manifest
		ErrorCode       int
		ErrorStatus     string
		Message         string
		ThrottleSeconds int
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsed := &response{}
	err = json.Unmarshal(body, parsed)
	if err != nil {
		slog.Error(string(body))
		return nil, err
	}

	_ = c.cache.Set(ctx, cacheKey, parsed.Response, time.Hour*24)
	return parsed.Response, nil
}

func (c *Client) GetMetricManifestDefinition(ctx context.Context, path string) (map[string]api.Destiny_Definitions_Metrics_DestinyMetricDefinition, error) {
	var cacheKey = "destiny:manifest:metric"
	ret := map[string]api.Destiny_Definitions_Metrics_DestinyMetricDefinition{}
	if found := c.lookupCacheItem(ctx, cacheKey, &ret); found {
		return ret, nil
	}

	url := fmt.Sprintf("%s%s", assetRootPath, path)
	slog.Info(url)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		_, _ = io.Copy(os.Stderr, resp.Body)
		fmt.Fprintln(os.Stderr)
		return nil, fmt.Errorf("500 currently having issues with the server")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsed := map[string]api.Destiny_Definitions_Metrics_DestinyMetricDefinition{}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		slog.Error(string(body))
		return nil, err
	}

	_ = c.cache.Set(ctx, cacheKey, parsed, time.Hour*24)
	return parsed, nil
}

func (c *Client) GetRecordManifestDefinition(ctx context.Context, path string) (map[string]RecordDefinition, error) {
	var cacheKey = "destiny:manifest:record"
	ret := map[string]RecordDefinition{}
	if found := c.lookupCacheItem(ctx, cacheKey, &ret); found {
		return ret, nil
	}

	url := fmt.Sprintf("%s%s", assetRootPath, path)
	slog.Info(url)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		_, _ = io.Copy(os.Stderr, resp.Body)
		fmt.Fprintln(os.Stderr)
		return nil, fmt.Errorf("500 currently having issues with the server")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsed := map[string]RecordDefinition{}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		slog.Error(string(body))
		return nil, err
	}

	_ = c.cache.Set(ctx, cacheKey, parsed, time.Hour*24)
	return parsed, nil
}

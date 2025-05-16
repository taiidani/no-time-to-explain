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

type Manifest struct {
	Version                  string          `json:"version"`
	MobileAssetContentPath   string          `json:"mobileAssetContentPath"`
	MobileGearAssetDataBases []ManifestEntry `json:"mobileGearAssetDataBases"`
	MobileWorldContentPaths  struct {
		English string `json:"en"`
	} `json:"mobileWorldContentPaths"`
	JsonWorldContentPaths struct {
		English string `json:"en"`
	} `json:"jsonWorldContentPaths"`
	JsonWorldComponentContentPaths struct {
		English map[string]string `json:"en"`
	} `json:"jsonWorldComponentContentPaths"`
	MobileClanBannerDatabasePath string `json:"mobileClanBannerDatabasePath"`
	MobileGearCDN                struct {
		Geometry    string
		Texture     string
		PlateRegion string
		Gear        string
		Shader      string
	} `json:"mobileGearCDN"`
	IconImagePyramidInfo []string `json:"iconImagePyramidInfo"`
}

type ManifestEntry struct {
	Version int    `json:"version"`
	Path    string `json:"path"`
}

func (c *Client) GetMetricsManifest(ctx context.Context) (map[string]api.Destiny_Definitions_Metrics_DestinyMetricDefinition, error) {
	manifest, err := c.getManifest(ctx)
	if err != nil {
		return map[string]api.Destiny_Definitions_Metrics_DestinyMetricDefinition{}, err
	}

	metricsManifestURL := manifest.JsonWorldComponentContentPaths.English["DestinyMetricDefinition"]
	metricsManifest := map[string]api.Destiny_Definitions_Metrics_DestinyMetricDefinition{}
	err = c.getManifestDefinition(ctx, metricsManifestURL, &metricsManifest)
	return metricsManifest, err
}

func (c *Client) GetMetricsManifestEntry(ctx context.Context, id string) (api.Destiny_Definitions_Metrics_DestinyMetricDefinition, error) {
	metricsManifest, err := c.GetMetricsManifest(ctx)
	if err != nil {
		return api.Destiny_Definitions_Metrics_DestinyMetricDefinition{}, err
	}

	metricManifestDefinition, ok := metricsManifest[id]
	if !ok {
		return metricManifestDefinition, fmt.Errorf("not found")
	}

	return metricManifestDefinition, nil
}

func (c *Client) GetRecordsManifest(ctx context.Context) (map[string]api.Destiny_Definitions_Records_DestinyRecordDefinition, error) {
	manifest, err := c.getManifest(ctx)
	if err != nil {
		return map[string]api.Destiny_Definitions_Records_DestinyRecordDefinition{}, err
	}

	metricsManifestURL := manifest.JsonWorldComponentContentPaths.English["DestinyRecordDefinition"]
	metricsManifest := map[string]api.Destiny_Definitions_Records_DestinyRecordDefinition{}
	err = c.getManifestDefinition(ctx, metricsManifestURL, &metricsManifest)
	return metricsManifest, err
}

func (c *Client) GetRecordsManifestEntry(ctx context.Context, id string) (api.Destiny_Definitions_Records_DestinyRecordDefinition, error) {
	recordsManifest, err := c.GetRecordsManifest(ctx)
	if err != nil {
		return api.Destiny_Definitions_Records_DestinyRecordDefinition{}, err
	}

	metricsManifestDefinition, ok := recordsManifest[id]
	if !ok {
		return metricsManifestDefinition, fmt.Errorf("not found")
	}

	return metricsManifestDefinition, nil
}

// GetManifest contains an index of all the localized assets for API resources.
// It can be used to extract the human-readable names for everything in the API.
// URL: https://bungie-net.github.io/multi/operation_get_Destiny2-GetDestinyManifest.html
func (c *Client) getManifest(ctx context.Context) (*Manifest, error) {
	var cacheKey = "destiny:manifest:info"
	ret := &Manifest{}
	if found := c.lookupCacheItem(ctx, cacheKey, ret); found {
		return ret, nil
	}

	url := fmt.Sprintf("%s/Destiny2/Manifest", apiRootPath)
	slog.Info(url)
	resp, err := c.client.Get(url)
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

func (c *Client) getManifestDefinition(_ context.Context, path string, target any) error {
	url := fmt.Sprintf("%s%s", assetRootPath, path)
	slog.Info(url)
	resp, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &target)
	if err != nil {
		slog.Error(string(body))
		return err
	}

	return nil
}

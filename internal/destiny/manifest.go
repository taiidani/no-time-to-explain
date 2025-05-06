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

type MetricManifestDefinition struct {
	DisplayProperties struct {
		Description   string `json:"description"`
		Name          string `json:"name"`
		Icon          string `json:"icon"`
		HasIcon       bool   `json:"hasIcon"`
		IconSequences []struct {
			Frames []string `json:"frames"`
		} `json:"iconSequences"`
	} `json:"displayProperties"`

	TrackingObjectiveHash int      `json:"trackingObjectiveHash"`
	LowerValueIsBetter    bool     `json:"lowerValueIsBetter"`
	PresentationNodeType  int      `json:"presentationNodeType"`
	TraitIds              []string `json:"traitIds"`
	TraitHashes           []int    `json:"traitHashes"`
	ParentNodeHashes      []int    `json:"parentNodeHashes"`
	Hash                  int      `json:"hash"`
	Index                 int      `json:"index"`
	Redacted              bool     `json:"redacted"`
	Blacklisted           bool     `json:"blacklisted"`
}

func (c *Client) GetMetricManifestDefinition(ctx context.Context, path string) (map[string]MetricManifestDefinition, error) {
	var cacheKey = "destiny:manifest:metric"
	ret := map[string]MetricManifestDefinition{}
	if found := c.lookupCacheItem(ctx, cacheKey, &ret); found {
		return ret, nil
	}

	url := fmt.Sprintf("%s%s", assetRootPath, path)
	slog.Info(url)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsed := map[string]MetricManifestDefinition{}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		slog.Error(string(body))
		return nil, err
	}

	_ = c.cache.Set(ctx, cacheKey, parsed, time.Hour*24)
	return parsed, nil
}

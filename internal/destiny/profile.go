package destiny

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/taiidani/go-bungie-api/api"
)

// https://bungie-net.github.io/multi/schema_Destiny-DestinyComponentType.html
type ComponentType int

const (
	ComponentTypeNone                  ComponentType = 0
	ComponentTypeProfiles              ComponentType = 100
	ComponentTypeVendorReceipts        ComponentType = 101
	ComponentTypeProfileInventories    ComponentType = 102
	ComponentTypeProfileCurrencies     ComponentType = 103
	ComponentTypeProfileProgression    ComponentType = 104
	ComponentTypePlatformSilver        ComponentType = 105
	ComponentTypeCharacters            ComponentType = 200
	ComponentTypeCharacterInventories  ComponentType = 201
	ComponentTypeCharacterProgressions ComponentType = 202
	ComponentTypeCharacterRenderData   ComponentType = 203
	ComponentTypeCharacterEquipment    ComponentType = 205
	ComponentTypeCharacterLoadouts     ComponentType = 206
	ComponentTypeItemInstances         ComponentType = 300
	ComponentTypeItemObjectives        ComponentType = 301
	ComponentTypeItemPerks             ComponentType = 302
	ComponentTypeItemRenderData        ComponentType = 303
	ComponentTypeItemStats             ComponentType = 304
	ComponentTypeItemSockets           ComponentType = 305
	ComponentTypeItemTalentGrids       ComponentType = 306
	ComponentTypeItemCommonData        ComponentType = 307
	ComponentTypeItemPlugStates        ComponentType = 308
	ComponentTypeItemPlugObjectives    ComponentType = 309
	ComponentTypeItemReusablePlugs     ComponentType = 310
	ComponentTypeVendors               ComponentType = 400
	ComponentTypeVendorCategories      ComponentType = 401
	ComponentTypeVendorSales           ComponentType = 402
	ComponentTypeKiosks                ComponentType = 500
	ComponentTypeCurrencyLookups       ComponentType = 600
	ComponentTypeCollectibles          ComponentType = 800
	ComponentTypeRecords               ComponentType = 900
	ComponentTypeTransitory            ComponentType = 1000
	ComponentTypeMetrics               ComponentType = 1100
	ComponentTypeStringVariables       ComponentType = 1200
	ComponentTypeCraftables            ComponentType = 1300
	ComponentTypeSocialCommendations   ComponentType = 1400
)

type Profile struct {
	// ComponentTypeRecords
	ProfileRecords api.SingleComponentResponseOfDestinyProfileRecordsComponent `json:"profileRecords"`

	// ComponentTypeRecords
	CharacterRecords api.DictionaryComponentResponseOfint64AndDestinyCharacterRecordsComponent `json:"characterRecords"`

	// ComponentTypeMetrics
	Metrics api.SingleComponentResponseOfDestinyMetricsComponent `json:"metrics"`
}

// GetProfile is an omnibus API endpoint allowing for retrieval of multiple sets of information at once
// URL: https://bungie-net.github.io/multi/operation_get_Destiny2-GetProfile.html
// TODO: Cache multiple components properly
func (c *Client) GetProfile(ctx context.Context, membershipType int, membershipID string, components ...ComponentType) (*Profile, error) {
	var cacheKey = fmt.Sprintf("destiny:profile:%d:%s:info", membershipType, membershipID)
	ret := &Profile{}
	if found := c.lookupCacheItem(ctx, cacheKey, ret); found {
		return ret, nil
	}

	// Convert the components to strings
	strComponents := []string{}
	for _, c := range components {
		strComponents = append(strComponents, fmt.Sprintf("%d", c))
	}

	url := fmt.Sprintf("%s/Destiny2/%d/Profile/%s/?components=%s", apiRootPath, membershipType, membershipID, strings.Join(strComponents, ","))
	slog.Info(url)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type response struct {
		Response        *Profile
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

	_ = c.cache.Set(ctx, cacheKey, parsed.Response, time.Hour*2)
	return parsed.Response, nil
}

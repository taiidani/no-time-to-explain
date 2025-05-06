package destiny

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"
)

const (
	groupTypeClan       = 1
	UnknownSpaceGroupID = 3760031
)

type Clan struct {
	ID           string `json:"groupId"`
	Name         string `json:"name"`
	CreationDate string `json:"creationDate"`
	About        string `json:"about"`
	MemberCount  int    `json:"memberCount"`
	Motto        string `json:"motto"`
	Theme        string `json:"theme"`
	BannerPath   string `json:"bannerPath"`
	AvatarPath   string `json:"avatarPath"`
	Features     struct {
		MaximumMembers int `json:"maximumMembers"`
	} `json:"features"`

	ClanInfo struct {
		ClanCallsign string `json:"clanCallsign"`
	} `json:"clanInfo"`
}

func (c *Client) GetClan(ctx context.Context, name string) (*Clan, error) {
	var cacheKey = "destiny:clan:info"
	ret := &Clan{}
	if found := c.lookupCacheItem(ctx, cacheKey, ret); found {
		return ret, nil
	}

	url := fmt.Sprintf("%s/GroupV2/Name/%s/%d/", apiRootPath, name, groupTypeClan)
	slog.Info(url)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type response struct {
		Response struct {
			Detail *Clan `json:"detail"`
		}
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

	_ = c.cache.Set(ctx, cacheKey, parsed.Response.Detail, time.Hour*24)
	return parsed.Response.Detail, nil
}

type ClanMember struct {
	MemberType             int            `json:"memberType"`
	IsOnline               bool           `json:"isOnline"`
	LastOnlineStatusChange string         `json:"lastOnlineStatusChange"`
	DestinyUserInfo        ClanMemberInfo `json:"destinyUserInfo"`
	JoinDate               string         `json:"joinDate"`
}

type ClanMemberInfo struct {
	LastSeenDisplayName         string `json:"LastSeenDisplayName"`
	LastSeenDisplayNameType     int    `json:"LastSeenDisplayNameType"`
	IconPath                    string `json:"iconPath"` // PSN Logo, for example
	CrossSaveOverride           int    `json:"crossSaveOverride"`
	IsPublic                    bool   `json:"isPublic"`
	DisplayName                 string `json:"displayName"`
	MembershipType              int    `json:"membershipType"`
	MembershipID                string `json:"membershipId"`
	BungieGlobalDisplayName     string `json:"bungieGlobalDisplayName"`
	BungieGlobalDisplayNameCode int    `json:"bungieGlobalDisplayNameCode"`
}

func (c *Client) GetClanMembers(ctx context.Context, clanID int) ([]ClanMember, error) {
	var cacheKey = fmt.Sprintf("destiny:clan:%d:members", clanID)
	ret := []ClanMember{}
	if found := c.lookupCacheItem(ctx, cacheKey, &ret); found {
		return ret, nil
	}

	url := fmt.Sprintf("%s/GroupV2/%d/Members/", apiRootPath, clanID)
	slog.Info(url)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type response struct {
		Response struct {
			Results []ClanMember `json:"results"`
		}
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

	_ = c.cache.Set(ctx, cacheKey, parsed.Response.Results, time.Hour*24)
	return parsed.Response.Results, nil
}

type ClanAggregateStat struct {
	Mode   int    `json:"mode"`
	StatId string `json:"statId"`
	Value  struct {
		Basic struct {
			Value        float32 `json:"value"` // Warning: This does not seem to populate correctly
			DisplayValue float32 `json:"displayValue"`
		} `json:"basic"`
	} `json:"value"`
}

// URL: https://bungie-net.github.io/multi/operation_get_Destiny2-GetClanAggregateStats.html
func (c *Client) GetClanAggregateStats(ctx context.Context) ([]ClanAggregateStat, error) {
	url := fmt.Sprintf("%s/Destiny2/Stats/AggregateClanStats/%d/", apiRootPath, UnknownSpaceGroupID)
	slog.Info(url)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type response struct {
		Response        []ClanAggregateStat `json:"response"`
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

	return parsed.Response, nil
}

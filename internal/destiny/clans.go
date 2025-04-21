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

const (
	groupTypeClan       = 1
	unknownSpaceGroupID = 3760031
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

var testFixtureClanMembers = []api.GroupsV2_GroupMember{
	{
		MemberType: 2,
		DestinyUserInfo: api.GroupsV2_GroupUserInfoCard{
			DisplayName:             "taiidani",
			IconPath:                "/img/theme/bungienet/icons/steamLogo.png",
			MembershipType:          3,
			MembershipId:            4611686018467493133,
			BungieGlobalDisplayName: "taiidani",
			// BungieGlobalDisplayNameCode: 2569,
		},
		JoinDate: "2023-01-22T23:28:29Z",
	},
	{
		MemberType: 2,
		DestinyUserInfo: api.GroupsV2_GroupUserInfoCard{
			IconPath:                "/img/theme/bungienet/icons/steamLogo.png",
			MembershipType:          3,
			MembershipId:            4611686018467505428,
			DisplayName:             "The Orange Knight",
			BungieGlobalDisplayName: "The Orange Knight",
			// BungieGlobalDisplayNameCode: 4901,
		},
		JoinDate: "2023-01-22T23:02:06Z",
	},
}

func (c *Client) GetClanMembers(ctx context.Context, clanID int) ([]api.GroupsV2_GroupMember, error) {
	if os.Getenv("DEV") == "true" {
		return testFixtureClanMembers, nil
	}

	var cacheKey = fmt.Sprintf("destiny:clan:%d:members", clanID)
	ret := []api.GroupsV2_GroupMember{}
	if found := c.lookupCacheItem(ctx, cacheKey, &ret); found {
		return ret, nil
	}

	// https://bungie-net.github.io/multi/operation_get_GroupV2-GetMembersOfGroup.html
	url := fmt.Sprintf("%s/GroupV2/%d/Members/", apiRootPath, clanID)
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

	// https://bungie-net.github.io/multi/schema_SearchResultOfGroupMember.html
	type response struct {
		Response struct {
			Results []api.GroupsV2_GroupMember `json:"results"`
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
	url := fmt.Sprintf("%s/Destiny2/Stats/AggregateClanStats/%d/", apiRootPath, unknownSpaceGroupID)
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

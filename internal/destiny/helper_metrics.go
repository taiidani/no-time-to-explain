package destiny

import (
	"context"
)

type Helper struct {
	client *Client
}

func NewHelper(client *Client) *Helper {
	return &Helper{client: client}
}

type HelperClanFish struct {
	TotalFish int
	Member    []HelperClanFishMember
}

type HelperClanFishMember struct {
	Name      string
	TotalFish int
}

// GetClanFish tracks the amount and largest fish caught by clan members.
// Number of fish MetricDefinition index: 24768693
// Number of fish objective hash: 2773717662
func (h *Helper) GetClanFish(ctx context.Context) (*MetricManifestDefinition, *HelperClanFish, error) {
	const fishMetricDefinition = "24768693"
	manifest, err := h.client.GetManifest(ctx)
	if err != nil {
		return nil, nil, err
	}

	metricsManifestURL := manifest.JsonWorldComponentContentPaths.English["DestinyMetricDefinition"]
	metricsManifest, err := h.client.GetMetricManifestDefinition(ctx, metricsManifestURL)
	if err != nil {
		return nil, nil, err
	}
	fishDefinition := metricsManifest[fishMetricDefinition]

	members, err := h.client.GetClanMembers(ctx, unknownSpaceGroupID)
	if err != nil {
		return &fishDefinition, nil, err
	}

	ret := &HelperClanFish{}
	for _, member := range members {
		metrics, err := h.client.GetProfile(ctx, member.DestinyUserInfo.MembershipType, member.DestinyUserInfo.MembershipID, ComponentTypeMetrics)
		if err != nil {
			return &fishDefinition, nil, err
		}

		clanMember := HelperClanFishMember{
			Name:      member.DestinyUserInfo.DisplayName,
			TotalFish: 0,
		}

		if fishMetrics, ok := metrics.Metrics.Data.Metrics[fishMetricDefinition]; ok {
			clanMember.TotalFish = fishMetrics.ObjectiveProgress.Progress
		}

		ret.Member = append(ret.Member, clanMember)
		ret.TotalFish += clanMember.TotalFish
	}

	return &fishDefinition, ret, nil
}

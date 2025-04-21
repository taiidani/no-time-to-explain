package destiny

import (
	"context"

	"github.com/taiidani/go-bungie-api/api"
)

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
func (h *Helper) GetClanFish(ctx context.Context) (*api.Destiny_Definitions_Metrics_DestinyMetricDefinition, *HelperClanFish, error) {
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
		profile, err := h.client.GetProfile(ctx, member.DestinyUserInfo.MembershipType, member.DestinyUserInfo.MembershipId)
		if err != nil {
			return &fishDefinition, nil, err
		}

		clanMember := HelperClanFishMember{
			Name:      member.DestinyUserInfo.DisplayName,
			TotalFish: 0,
		}

		if fishMetrics, ok := profile.Metrics.Data.Metrics[fishMetricDefinition]; ok {
			clanMember.TotalFish += int(*fishMetrics.ObjectiveProgress.Progress)
		}

		ret.Member = append(ret.Member, clanMember)
		ret.TotalFish += clanMember.TotalFish
	}

	return &fishDefinition, ret, nil
}

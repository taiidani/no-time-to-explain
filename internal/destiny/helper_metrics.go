package destiny

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/taiidani/no-time-to-explain/internal/models"
)

type HelperClanFish struct {
	TotalFish int
	Member    []HelperClanFishMember
}

type HelperClanFishMember struct {
	Name      string
	TotalFish int
}

// GetPlayerMetrics tracks the metrics earned by every player.
// Note: This is an expensive operation!
func (h *Helper) GetPlayerMetrics(ctx context.Context) ([]models.PlayerMetric, error) {
	players, err := models.GetPlayers(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get players: %w", err)
	}

	ret := []models.PlayerMetric{}
	for _, player := range players {
		log := slog.With("id", player.ID, "membership-type", player.MembershipType, "membership-id", player.MembershipId)
		log.Info("Refreshing player metrics")
		metrics, err := h.client.GetProfile(ctx, player.MembershipType, player.MembershipId, ComponentTypeMetrics)
		if err != nil {
			return ret, fmt.Errorf("could not get player %q %q profile: %w", player.MembershipType, player.MembershipId, err)
		} else if metrics.Metrics.Data.Metrics == nil {
			log.Warn("Player has no metrics")
			continue
		}

		for key, metric := range metrics.Metrics.Data.Metrics {
			ret = append(ret, models.PlayerMetric{
				PlayerID:        player.ID,
				MetricID:        key,
				ObjectiveHash:   metric.ObjectiveProgress.ObjectiveHash,
				Progress:        metric.ObjectiveProgress.Progress,
				CompletionValue: metric.ObjectiveProgress.CompletionValue,
				Complete:        metric.ObjectiveProgress.Complete,
				Visible:         metric.ObjectiveProgress.Visible,
			})
		}
	}

	return ret, nil
}

package destiny

import (
	"context"
	"fmt"

	"github.com/taiidani/no-time-to-explain/internal/models"
)

func (h *Helper) GetFishMetrics(ctx context.Context) (string, []models.PlayerMetric, error) {
	const fishMetricDefinition = "24768693"

	manifest, err := h.client.GetMetricsManifestEntry(ctx, fishMetricDefinition)
	if err != nil {
		return "", nil, err
	}

	players, err := models.GetPlayers(ctx)
	if err != nil {
		return manifest.DisplayProperties.Name, nil, fmt.Errorf("get players error: %w", err)
	}

	ret := []models.PlayerMetric{}
	for _, player := range players {
		metric, err := models.GetPlayerMetric(ctx, player.ID, fishMetricDefinition)
		if err != nil {
			return manifest.DisplayProperties.Name, nil, fmt.Errorf("player %q metric %q error: %w", player.ID, fishMetricDefinition, err)
		}

		ret = append(ret, *metric)
	}

	return manifest.DisplayProperties.Name, ret, nil
}

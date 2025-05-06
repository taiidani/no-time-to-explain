package destiny

import (
	"context"
	"fmt"
)

func (h *Helper) GetManifestMetricEntry(ctx context.Context, id string) (MetricManifestDefinition, error) {
	manifest, err := h.client.GetManifest(ctx)
	if err != nil {
		return MetricManifestDefinition{}, err
	}

	metricsManifestURL := manifest.JsonWorldComponentContentPaths.English["DestinyMetricDefinition"]
	metricsManifest, err := h.client.GetMetricManifestDefinition(ctx, metricsManifestURL)
	if err != nil {
		return MetricManifestDefinition{}, err
	}

	metricManifestDefinition, ok := metricsManifest[id]
	if !ok {
		return metricManifestDefinition, fmt.Errorf("not found")
	}

	return metricManifestDefinition, nil
}

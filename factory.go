//go:generate mdatagen metadata.yaml

package downsampleprocessor

import (
	"context"

	"github.com/tosuke/downsampleprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithMetrics(createMetrics, metadata.MetricsStability))
}

func createMetrics(_ context.Context, settings processor.CreateSettings, cfg component.Config, next consumer.Metrics) (processor.Metrics, error) {
	return newDownsampleProcessor(settings, cfg.(*Config), next), nil
}

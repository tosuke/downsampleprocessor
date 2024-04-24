//go:generate mdatagen metadata.yaml

package downsampleprocessor

import (
	"context"
	"fmt"

	"github.com/tosuke/downsampleprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

func NewFactory() processor.Factory {
	return processor.NewFactory(metadata.Type, createDefaultConfig, nil)
}

func createMetricProcessor(_ context.Context, _ processor.CreateSettings, _ component.Config, _ consumer.Metrics) (processor.Metrics, error) {
	return nil, fmt.Errorf("unimplemented")
}

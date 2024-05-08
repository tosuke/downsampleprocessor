package downsampleprocessor

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type downsampleProcessor struct {
	m sync.Mutex

	shutdownFn context.CancelFunc

	period         time.Duration
	maxCardinality int
}

func newDownsampleProcessor(c *Config) *downsampleProcessor {
	return &downsampleProcessor{
		period:         c.Period,
		maxCardinality: c.MaxCardinality,
	}
}

func (dp *downsampleProcessor) start(_ context.Context, _ component.Host) error {
	ctx, cancel := context.WithCancel(context.Background())
	dp.shutdownFn = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (dp *downsampleProcessor) shutdown(_ context.Context) error {
	if dp.shutdownFn != nil {
		dp.shutdownFn()
	}
	return nil
}

func (dp *downsampleProcessor) processMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	dp.m.Lock()
	defer dp.m.Unlock()

	return md, nil
}

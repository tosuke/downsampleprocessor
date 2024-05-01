package downsampleprocessor_test

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/tosuke/downsampleprocessor"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestProcessor_SingleMetric(t *testing.T) {
	t.Parallel()

	factory := downsampleprocessor.NewFactory()
	ctx := context.Background()
	sink := new(consumertest.MetricsSink)

	pcs := processortest.NewNopCreateSettings()
	mproc, err := factory.CreateMetricsProcessor(ctx, pcs, factory.CreateDefaultConfig(), sink)
	if err != nil {
		t.Errorf("failed to create processor: %v", err)
        return
	}
	if err := mproc.Start(ctx, componenttest.NewNopHost()); err != nil {
		t.Errorf("failed to start processor: %v", err)
        return
	}
	defer func() {
        if err := mproc.Shutdown(ctx); err != nil {
            t.Errorf("failed to shutdown processor: %v", err)
            return
        }
    }()

	md := generateTestMetrics()
	if err := mproc.ConsumeMetrics(ctx, md); err != nil {
		t.Errorf("failed to process metrics: %v", err)
        return
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Printf("%+v\n", sink.AllMetrics())

	hasData := slices.Contains(sink.AllMetrics(), md)
	if !hasData {
		t.Errorf("expected to find metrics in sink")
	}
}

func generateTestMetrics() pmetric.Metrics {
	metrics := pmetric.NewMetrics()

	rm := metrics.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("resource", "R1")

	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().Attributes().PutStr("scope", "S1")

	m := sm.Metrics().AppendEmpty()
	m.SetName("test_metric")
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.Attributes().PutStr("attr", "A1")
	dp.SetIntValue(1234)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	return metrics
}

package downsampleprocessor_test

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/tosuke/downsampleprocessor"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processortest"
)

const (
	resourceKey = "resource"
	scopeKey    = "scope"
	attrKey     = "attr"
	metricName  = "test_metric"
)

func TestProcessor_SingleMetric(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sink := new(consumertest.MetricsSink)
	mproc, err := setupProcessor(t, ctx, sink, nil)
	if err != nil {
		t.Errorf("failed to setup processor: %v", err)
		return
	}

	now := time.Now()
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr(resourceKey, "R1")
	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().Attributes().PutStr(scopeKey, "S1")
	m := sm.Metrics().AppendEmpty()
	m.SetName(metricName)
	dps := m.SetEmptyGauge().DataPoints()

	dp1 := dps.AppendEmpty()
	dp1.Attributes().PutStr(attrKey, "A1")
	dp1.SetTimestamp(pcommon.NewTimestampFromTime(now))
	dp1.SetIntValue(1)

	dp2 := dps.AppendEmpty()
	dp2.Attributes().PutStr(attrKey, "A1")
	dp2.SetTimestamp(pcommon.NewTimestampFromTime(now.Add(-10 * time.Millisecond)))
    dp2.SetIntValue(2)

	if err := mproc.ConsumeMetrics(ctx, md); err != nil {
		t.Errorf("failed to process metrics: %v", err)
		return
	}

	time.Sleep(200 * time.Millisecond)

	assertHasDatapoint(t, sink.AllMetrics(), "R1", "S1", metricName, "A1", 1)
    assertHasDatapoint(t, sink.AllMetrics(), "R1", "S1", metricName, "A1", 2)
}

func setupProcessor(t testing.TB, ctx context.Context, next consumer.Metrics, configFn func(*downsampleprocessor.Config) *downsampleprocessor.Config) (processor.Metrics, error) {
	factory := downsampleprocessor.NewFactory()

	cfg := factory.CreateDefaultConfig()
	if configFn != nil {
		cfg = configFn(cfg.(*downsampleprocessor.Config))
	}

	if next == nil {
		next = consumertest.NewNop()
	}

	pcs := processortest.NewNopCreateSettings()
	mproc, err := factory.CreateMetricsProcessor(ctx, pcs, cfg, next)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}

	host := componenttest.NewNopHost()
	if err := mproc.Start(ctx, host); err != nil {
		return nil, fmt.Errorf("failed to start processor: %w", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := mproc.Shutdown(ctx); err != nil {
			t.Errorf("failed to shutdown processor: %v", err)
		}
	})

	return mproc, nil
}

func assertHasDatapoint(
	t testing.TB, metrics []pmetric.Metrics,
	resourceValue, scopeValue, metricName, metricAttrValue string,
	datapointValue int64,
) (ok bool) {
	t.Helper()

	ok = hasDatapoint(metrics, resourceValue, scopeValue, metricName, metricAttrValue, datapointValue)
	if !ok {
		t.Errorf("expected to find metrics, resource{%s=%s}; scope{%s=%s}; name=%s; attr{%s=%s}; value=%d\n",
			resourceKey, resourceValue, scopeKey, scopeValue, metricName, attrKey, metricAttrValue, datapointValue)
	}
	return
}

func assertHasNotDatapoint(
	t testing.TB, metrics []pmetric.Metrics,
	resourceValue, scopeValue, metricName, metricAttrValue string,
	datapointValue int64,
) (ok bool) {
	t.Helper()
	ok = !hasDatapoint(metrics, resourceValue, scopeValue, metricName, metricAttrValue, datapointValue)
	if !ok {
		t.Errorf("expected to not find metrics in sink, resource{%s=%s}; scope{%s=%s}; name=%s; attr{%s=%s}; value=%d\n",
			resourceKey, resourceValue, scopeKey, scopeValue, metricName, attrKey, metricAttrValue, datapointValue)
	}
	return
}

func hasDatapoint(
	metrics []pmetric.Metrics,
	resourceValue, scopeValue, metricName, metricAttrValue string,
	datapointValue int64,
) bool {
	return slices.ContainsFunc(metrics, func(md pmetric.Metrics) bool {
		rms := md.ResourceMetrics()
		for i := range rms.Len() {
			rm := rms.At(i)
			if v, ok := rm.Resource().Attributes().Get(resourceKey); !ok || v.Str() != resourceValue {
				continue
			}
			sms := rm.ScopeMetrics()
			for j := range sms.Len() {
				sm := sms.At(j)
				if v, ok := sm.Scope().Attributes().Get(scopeKey); !ok || v.Str() != scopeValue {
					continue
				}
				ms := sm.Metrics()
				for k := range ms.Len() {
					m := ms.At(k)
					if m.Name() != metricName {
						continue
					}
					dps := m.Gauge().DataPoints()
					for l := range dps.Len() {
						dp := dps.At(l)
						if v, ok := dp.Attributes().Get(attrKey); !ok || v.Str() != metricAttrValue {
							continue
						}
						if dp.IntValue() != datapointValue {
							continue
						}
						return true
					}
				}
			}
		}
		return false
	})
}

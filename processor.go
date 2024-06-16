package downsampleprocessor

import (
	"context"
	"sync"
	"time"

	"github.com/tosuke/downsampleprocessor/internal/identity"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type downsampleProcessor struct {
	m sync.Mutex

	shutdownFn context.CancelFunc

	period time.Duration

	tracker *tracker
}

func newDownsampleProcessor(c *Config) *downsampleProcessor {
	return &downsampleProcessor{
		period:  c.Period,
		tracker: newTracker(c),
	}
}

func (dp *downsampleProcessor) start(_ context.Context, _ component.Host) error {
	ctx, cancel := context.WithCancel(context.Background())
	dp.shutdownFn = cancel

	if dp.period == 0 {
		return nil
	}

	ticker := time.NewTicker(dp.period)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dp.sweep()
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

func (dsp *downsampleProcessor) processMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	dsp.m.Lock()
	defer dsp.m.Unlock()

	md.ResourceMetrics().RemoveIf(func(rm pmetric.ResourceMetrics) bool {
		ri := identity.OfResource(rm.Resource())
		rm.ScopeMetrics().RemoveIf(func(sm pmetric.ScopeMetrics) bool {
			isi := identity.OfScope(ri, sm.Scope())
			sm.Metrics().RemoveIf(func(m pmetric.Metric) bool {
				mi := identity.OfMetric(isi, m)
				switch m.Type() {
				case pmetric.MetricTypeGauge:
					return handleDatapointSlice(dsp.tracker, mi, m.Gauge().DataPoints())
				case pmetric.MetricTypeSum:
					return handleDatapointSlice(dsp.tracker, mi, m.Sum().DataPoints())
				case pmetric.MetricTypeHistogram:
					return handleDatapointSlice(dsp.tracker, mi, m.Histogram().DataPoints())
				case pmetric.MetricTypeExponentialHistogram:
					return handleDatapointSlice(dsp.tracker, mi, m.ExponentialHistogram().DataPoints())
				default:
					return false
				}
			})
			return sm.Metrics().Len() == 0
		})
		return rm.ScopeMetrics().Len() == 0
	})

	return md, nil
}

type datapoint interface {
	Attributes() pcommon.Map
	Timestamp() pcommon.Timestamp
}
type datapointSlice[DP datapoint] interface {
	Len() int
	Sort(func(a, b DP) bool)
	RemoveIf(func(DP) bool)
}

func cmpDatapoint[DP datapoint](a, b DP) bool {
	return a.Timestamp().AsTime().Before(b.Timestamp().AsTime())
}

func handleDatapointSlice[DP datapoint, DPS datapointSlice[DP]](tracker *tracker, mi identity.Metric, dps DPS) bool {
	dps.Sort(func(a, b DP) bool {
		return a.Timestamp().AsTime().Before(b.Timestamp().AsTime())
	})
	dps.RemoveIf(func(dp DP) bool {
		si := identity.OfStream(mi, dp)
		return tracker.shouldRemove(si, dp.Timestamp().AsTime())
	})
	return dps.Len() == 0
}

func (dp *downsampleProcessor) sweep() {
	dp.m.Lock()
	defer dp.m.Unlock()
	dp.tracker.sweep()
}

type tracker struct {
	period         time.Duration
	maxCardinality int

	newMap *identity.Map[identity.Stream, *trackerEntry]
	oldMap *identity.Map[identity.Stream, *trackerEntry]
}
type trackerEntry struct {
	timestamp time.Time
}

func newTracker(cfg *Config) *tracker {
	return &tracker{
		period:         cfg.Period,
		maxCardinality: cfg.MaxCardinality,
		newMap:         identity.NewMap[identity.Stream, *trackerEntry](),
		oldMap:         identity.NewMap[identity.Stream, *trackerEntry](),
	}
}

func (t *tracker) shouldRemove(s identity.Stream, timestamp time.Time) bool {
	entry, ok := t.newMap.Get(s)
	if !ok {
		entry, ok = t.oldMap.Get(s)
		if ok {
			t.newMap.Set(s, entry)
		} else {
			if t.newMap.Len() < t.maxCardinality {
				entry = &trackerEntry{
					timestamp: timestamp,
				}
				t.newMap.Set(s, entry)
			}
			return false
		}
	}

	if entry.timestamp.Add(t.period).After(timestamp) {
		return true
	}
	entry.timestamp = timestamp
	return false
}

func (t *tracker) sweep() {
	t.oldMap, t.newMap = t.newMap, identity.NewMap[identity.Stream, *trackerEntry]()
}

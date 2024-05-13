package identity

import (
	"fmt"
	"hash"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type resource = Resource
type Resource struct {
	attrs pcommon.Map
}

func OfResource(r pcommon.Resource) Resource {
	return Resource{attrs: r.Attributes()}
}

func (ra Resource) Equals(rb Resource) bool {
	return equalMap(ra.attrs, rb.attrs)
}

func (r Resource) WriteHash(h hash.Hash) {
	sum := pdatautil.MapHash(r.attrs)
	h.Write(sum[:])
}

type scope = Scope
type Scope struct {
	resource
	name    string
	version string
	attrs   pcommon.Map
}

func OfScope(r Resource, s pcommon.InstrumentationScope) Scope {
	return Scope{
		resource: r,
		name:     s.Name(),
		version:  s.Version(),
		attrs:    s.Attributes(),
	}
}

func (sa Scope) Equals(sb Scope) bool {
	return sa.resource.Equals(sb.resource) &&
		sa.name == sb.name &&
		sa.version == sb.version &&
		equalMap(sa.attrs, sb.attrs)
}

func (s Scope) WriteHash(h hash.Hash) {
	s.resource.WriteHash(h)
	h.Write([]byte(s.name))
	h.Write([]byte(s.version))
	sum := pdatautil.MapHash(s.attrs)
	h.Write(sum[:])
}

type metric = Metric
type Metric struct {
	scope
	name        string
	unit        string
	typ         pmetric.MetricType
	monotonic   bool
	temporality pmetric.AggregationTemporality
}

func OfMetric(s Scope, m pmetric.Metric) Metric {
	id := Metric{
		scope: s,
		name:  m.Name(),
		unit:  m.Unit(),
		typ:   m.Type(),
	}

	switch m.Type() {
	case pmetric.MetricTypeGauge:
	case pmetric.MetricTypeSum:
		sum := m.Sum()
		id.monotonic = sum.IsMonotonic()
		id.temporality = sum.AggregationTemporality()
	case pmetric.MetricTypeHistogram:
		hist := m.Histogram()
		id.monotonic = true
		id.temporality = hist.AggregationTemporality()
	case pmetric.MetricTypeExponentialHistogram:
		exp := m.ExponentialHistogram()
		id.monotonic = true
		id.temporality = exp.AggregationTemporality()
	case pmetric.MetricTypeSummary:
	default:
		panic(fmt.Errorf("unsupported metric type: %v", m.Type()))
	}

	return id
}

func (ma Metric) Equals(mb Metric) bool {
	return ma.scope.Equals(mb.scope) &&
		ma.name == mb.name &&
		ma.unit == mb.unit &&
		ma.typ == mb.typ &&
		ma.monotonic == mb.monotonic &&
		ma.temporality == mb.temporality
}

func (m Metric) WriteHash(h hash.Hash) {
	m.scope.WriteHash(h)
	h.Write([]byte(m.name))
	h.Write([]byte(m.unit))

	var monotonic byte
	if m.monotonic {
		monotonic = 1
	}
	h.Write([]byte{byte(m.typ), monotonic, byte(m.temporality)})
}

type Stream struct {
	metric
	attrs pcommon.Map
}

func OfStream[DP dataPoint](m Metric, dp DP) Stream {
	return Stream{
		metric: m,
		attrs:  dp.Attributes(),
	}
}

type dataPoint interface {
	Attributes() pcommon.Map
}

func (sa Stream) Equals(sb Stream) bool {
	return sa.metric.Equals(sb.metric) && equalMap(sa.attrs, sb.attrs)
}

func (s Stream) WriteHash(h hash.Hash) {
	s.metric.WriteHash(h)
	sum := pdatautil.MapHash(s.attrs)
	h.Write(sum[:])
}

package identity

import (
	"bytes"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

func equalValue(a, b pcommon.Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch a.Type() {
	case pcommon.ValueTypeEmpty:
		return true
	case pcommon.ValueTypeStr:
		return a.Str() == b.Str()
	case pcommon.ValueTypeInt:
		return a.Int() == b.Int()
	case pcommon.ValueTypeDouble:
		return a.Double() == b.Double()
	case pcommon.ValueTypeBool:
		return a.Bool() == b.Bool()
	case pcommon.ValueTypeMap:
		return equalMap(a.Map(), b.Map())
	case pcommon.ValueTypeSlice:
		return equalSlice(a.Slice(), b.Slice())
	case pcommon.ValueTypeBytes:
		return bytes.Equal(a.Bytes().AsRaw(), b.Bytes().AsRaw())
	}
	return false
}

func equalSlice(a, b pcommon.Slice) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := range a.Len() {
		if !equalValue(a.At(i), b.At(i)) {
			return false
		}
	}
	return true
}

var nilMap pcommon.Map

func equalMap(a, b pcommon.Map) bool {
	na, nb := a == nilMap, b == nilMap
	if na || nb {
		return na == nb
	}

	if a.Len() != b.Len() {
		return false
	}
	ma := make(map[string]pcommon.Value, a.Len())
	a.Range(func(k string, v pcommon.Value) bool {
		ma[k] = v
		return true
	})

	equals := true
	b.Range(func(k string, vb pcommon.Value) bool {
		va, ok := ma[k]
		if !ok || !equalValue(va, vb) {
			equals = false
			return false
		}
		return true
	})
	return equals
}

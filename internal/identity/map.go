package identity

import (
	"hash"
	"hash/maphash"
	"slices"
)

type Map[K hashEqualer[K], V any] struct {
	orig map[uint64]*[]mapKV[K, V]
	len  int
	h    *maphash.Hash
}
type hashEqualer[K any] interface {
	Equals(K) bool
	WriteHash(hash.Hash)
}
type mapKV[K hashEqualer[K], V any] struct {
	key   K
	value V
}

func NewMap[K hashEqualer[K], V any]() *Map[K, V] {
	return &Map[K, V]{
		orig: map[uint64]*[]mapKV[K, V]{},
		h:    new(maphash.Hash),
	}
}

func (m *Map[K, V]) Get(key K) (val V, ok bool) {
	kvs, ok := m.orig[m.hash(key)]
	if !ok {
		return
	}
	i := slices.IndexFunc(*kvs, func(kv mapKV[K, V]) bool {
		return kv.key.Equals(key)
	})
	if i == -1 {
		return val, false
	}
	return (*kvs)[i].value, true
}

func (m *Map[_, _]) Len() int {
	return m.len
}

func (m *Map[K, V]) Set(key K, val V) {
	hash := m.hash(key)
	entry, ok := m.orig[hash]
	if !ok {
		kvs := make([]mapKV[K, V], 0, 1)
		entry = &kvs
		m.orig[hash] = entry
	}
	i := slices.IndexFunc(*entry, func(kv mapKV[K, V]) bool {
		return kv.key.Equals(key)
	})
	if i == -1 {
		kvs := append(*entry, mapKV[K, V]{key: key, value: val})
		*entry = kvs
		m.len++
	} else {
		(*entry)[i].value = val
	}
}

func (m *Map[K, V]) Delete(key K) {
	hash := m.hash(key)
	entry, ok := m.orig[hash]
	if !ok {
		return
	}
	l := len(*entry)
	*entry = slices.DeleteFunc(*entry, func(kv mapKV[K, V]) bool {
		return kv.key.Equals(key)
	})
	if len(*entry) != l {
		m.len--
	}
	if len(*entry) == 0 {
		delete(m.orig, hash)
	}
}

func (m *Map[K, _]) hash(key K) uint64 {
	h := m.h
	h.Reset()
	key.WriteHash(h)
	return h.Sum64()
}

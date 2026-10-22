// Package cache provides small, dependency-free in-memory caches for data that
// is read on nearly every request but written rarely — reference lookups and
// derived structures that would otherwise hit the database (and be recomputed)
// on every call.
package cache

import (
	"sync"
	"time"
)

// Value is a concurrency-safe, single-value cache with load-through semantics
// and an optional TTL. The first reader after a miss loads and stores the
// value; later readers get it without touching the source until the entry
// expires or is explicitly invalidated.
//
// It is meant for whole-collection reference data (e.g. the list of faculties
// or the navigation graph), not for per-key query results.
//
// A nil *Value is a valid, caching-disabled cache: Get always loads and
// Invalidate is a no-op. This lets callers switch caching off (ttl <= 0)
// without branching at every call site.
type Value[T any] struct {
	mu      sync.Mutex
	ttl     time.Duration // 0 means entries never expire on their own
	loaded  bool
	expires time.Time
	val     T
}

// New returns a cache whose entries expire after ttl. A ttl of 0 keeps the
// value until it is explicitly invalidated.
func New[T any](ttl time.Duration) *Value[T] {
	return &Value[T]{ttl: ttl}
}

// Get returns the cached value, loading it via load on a miss or once the TTL
// has elapsed. load runs under the cache lock, so concurrent callers racing on
// a cold entry trigger a single load rather than a stampede. A load error is
// returned to the caller and never cached.
func (v *Value[T]) Get(load func() (T, error)) (T, error) {
	if v == nil {
		return load()
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.loaded && (v.ttl == 0 || time.Now().Before(v.expires)) {
		return v.val, nil
	}

	val, err := load()
	if err != nil {
		return val, err
	}

	v.val = val
	v.loaded = true
	if v.ttl > 0 {
		v.expires = time.Now().Add(v.ttl)
	}
	return val, nil
}

// Invalidate drops the cached value so the next Get reloads it. Callers run it
// after a write that changes the underlying data.
func (v *Value[T]) Invalidate() {
	if v == nil {
		return
	}

	v.mu.Lock()
	v.loaded = false
	var zero T
	v.val = zero
	v.mu.Unlock()
}

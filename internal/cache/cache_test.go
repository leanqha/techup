package cache

import (
	"errors"
	"testing"
	"time"
)

func TestValueGetLoadsOnceUntilInvalidated(t *testing.T) {
	c := New[int](time.Minute)

	calls := 0
	load := func() (int, error) {
		calls++
		return calls, nil
	}

	v, err := c.Get(load)
	if err != nil || v != 1 {
		t.Fatalf("first Get = (%d, %v), want (1, nil)", v, err)
	}

	// Warm hit: loader must not run again.
	v, err = c.Get(load)
	if err != nil || v != 1 {
		t.Fatalf("second Get = (%d, %v), want (1, nil)", v, err)
	}
	if calls != 1 {
		t.Fatalf("loader ran %d times, want 1", calls)
	}

	// After invalidation the next Get reloads.
	c.Invalidate()
	v, err = c.Get(load)
	if err != nil || v != 2 {
		t.Fatalf("Get after Invalidate = (%d, %v), want (2, nil)", v, err)
	}
	if calls != 2 {
		t.Fatalf("loader ran %d times after invalidate, want 2", calls)
	}
}

func TestValueGetReloadsAfterTTL(t *testing.T) {
	c := New[int](5 * time.Millisecond)

	calls := 0
	load := func() (int, error) {
		calls++
		return calls, nil
	}

	if _, err := c.Get(load); err != nil {
		t.Fatalf("Get: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if _, err := c.Get(load); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if calls != 2 {
		t.Fatalf("loader ran %d times, want 2 (expired entry should reload)", calls)
	}
}

func TestValueGetDoesNotCacheErrors(t *testing.T) {
	c := New[int](time.Minute)

	boom := errors.New("boom")
	if _, err := c.Get(func() (int, error) { return 0, boom }); !errors.Is(err, boom) {
		t.Fatalf("Get error = %v, want boom", err)
	}

	// A failed load must not populate the cache: the next Get should load again
	// and succeed.
	v, err := c.Get(func() (int, error) { return 42, nil })
	if err != nil || v != 42 {
		t.Fatalf("Get after error = (%d, %v), want (42, nil)", v, err)
	}
}

func TestValueZeroTTLNeverExpires(t *testing.T) {
	c := New[int](0)

	calls := 0
	load := func() (int, error) { calls++; return calls, nil }

	if _, err := c.Get(load); err != nil {
		t.Fatalf("Get: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if _, err := c.Get(load); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if calls != 1 {
		t.Fatalf("loader ran %d times, want 1 (ttl=0 should not expire)", calls)
	}
}

func TestNilValueAlwaysLoads(t *testing.T) {
	var c *Value[int] // caching disabled

	calls := 0
	load := func() (int, error) { calls++; return calls, nil }

	for i := 0; i < 3; i++ {
		v, err := c.Get(load)
		if err != nil || v != i+1 {
			t.Fatalf("Get #%d = (%d, %v), want (%d, nil)", i, v, err, i+1)
		}
	}
	// Must not panic.
	c.Invalidate()
	if calls != 3 {
		t.Fatalf("loader ran %d times, want 3 (nil cache never caches)", calls)
	}
}

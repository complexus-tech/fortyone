package testkit

import (
	"sync"
	"time"
)

// FixedClock is an immutable clock for tests whose decision time never moves.
// It satisfies the application's structural Clock interfaces.
type FixedClock struct {
	now time.Time
}

// NewFixedClock returns a clock that always reports now. The supplied location
// is preserved so tests must choose their timezone deliberately.
func NewFixedClock(now time.Time) FixedClock {
	return FixedClock{now: now}
}

// Now returns the configured fixed time.
func (c FixedClock) Now() time.Time {
	return c.now
}

// ManualClock is a concurrency-safe clock for tests that advance through
// expiry, retry, or scheduling decisions without sleeping.
type ManualClock struct {
	mu  sync.RWMutex
	now time.Time
}

// NewManualClock returns a manually controlled clock at now.
func NewManualClock(now time.Time) *ManualClock {
	return &ManualClock{now: now}
}

// Now returns the current manual time.
func (c *ManualClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

// Set moves the clock to now and returns the new value.
func (c *ManualClock) Set(now time.Time) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = now
	return c.now
}

// Advance moves the clock by delta and returns the new value. Negative deltas
// are permitted for tests that explicitly exercise clock-skew behavior.
func (c *ManualClock) Advance(delta time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(delta)
	return c.now
}

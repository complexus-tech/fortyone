package testkit

import (
	"sync"
	"testing"
	"time"
)

func TestFixedClockPreservesInstantAndLocation(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	now := time.Date(2026, time.August, 28, 9, 30, 0, 0, location)
	clock := NewFixedClock(now)

	if got := clock.Now(); !got.Equal(now) || got.Location() != location {
		t.Fatalf("fixed clock = %v in %v, want %v in %v", got, got.Location(), now, location)
	}
}

func TestManualClockSetAndAdvance(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 28, 9, 30, 0, 0, time.UTC)
	clock := NewManualClock(start)
	if got := clock.Advance(15 * time.Minute); !got.Equal(start.Add(15 * time.Minute)) {
		t.Fatalf("advanced time = %v", got)
	}

	replacement := start.Add(-time.Hour)
	if got := clock.Set(replacement); !got.Equal(replacement) {
		t.Fatalf("set time = %v", got)
	}
	if got := clock.Now(); !got.Equal(replacement) {
		t.Fatalf("current time = %v, want %v", got, replacement)
	}
}

func TestManualClockIsSafeForConcurrentAdvanceAndRead(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.August, 28, 9, 30, 0, 0, time.UTC)
	clock := NewManualClock(start)

	const (
		workers  = 16
		advances = 250
	)
	var waitGroup sync.WaitGroup
	for range workers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for range advances {
				clock.Advance(time.Millisecond)
				_ = clock.Now()
			}
		}()
	}
	waitGroup.Wait()

	want := start.Add(workers * advances * time.Millisecond)
	if got := clock.Now(); !got.Equal(want) {
		t.Fatalf("concurrently advanced time = %v, want %v", got, want)
	}
}

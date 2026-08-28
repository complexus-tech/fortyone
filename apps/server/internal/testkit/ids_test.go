package testkit

import (
	"sync"
	"testing"

	"github.com/google/uuid"
)

func TestUUIDSourceProducesStableUniqueSequence(t *testing.T) {
	t.Parallel()

	first := NewUUIDSource("stories")
	second := NewUUIDSource("stories")
	different := NewUUIDSource("comments")

	for index := range 10 {
		firstID := first.New()
		secondID := second.New()
		if firstID != secondID {
			t.Fatalf("sequence position %d differed: %s != %s", index, firstID, secondID)
		}
		if firstID == different.New() {
			t.Fatalf("different seed matched at sequence position %d", index)
		}
		if firstID.Version() != 5 {
			t.Fatalf("UUID version = %d, want deterministic SHA-1 version 5", firstID.Version())
		}
	}
}

func TestUUIDSourceIsSafeAndUniqueUnderConcurrency(t *testing.T) {
	t.Parallel()

	const total = 1024
	source := NewUUIDSource("concurrent")
	results := make(chan uuid.UUID, total)
	var waitGroup sync.WaitGroup
	for range total {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			results <- source.New()
		}()
	}
	waitGroup.Wait()
	close(results)

	seen := make(map[uuid.UUID]struct{}, total)
	for id := range results {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate deterministic UUID: %s", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != total {
		t.Fatalf("unique UUID count = %d, want %d", len(seen), total)
	}
}

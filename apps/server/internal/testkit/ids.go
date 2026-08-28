package testkit

import (
	"encoding/binary"
	"math"
	"sync"

	"github.com/google/uuid"
)

// UUIDSource emits a reproducible UUID sequence for assertions and fixtures.
// It is safe for concurrent use; concurrent callers receive the same unique set
// of IDs, although goroutine scheduling determines which caller receives each
// sequence position.
type UUIDSource struct {
	mu        sync.Mutex
	namespace uuid.UUID
	next      uint64
}

// NewUUIDSource creates a deterministic source scoped by a stable, descriptive
// seed. Independent sources with the same seed emit the same sequence.
func NewUUIDSource(seed string) *UUIDSource {
	namespace := uuid.NewSHA1(uuid.NameSpaceOID, []byte("fortyone:testkit:"+seed))
	return &UUIDSource{namespace: namespace}
}

// New returns the next deterministic UUID. It panics only if a single test
// source exhausts all uint64 sequence values, because continuing would emit a
// duplicate identifier and invalidate the test.
func (s *UUIDSource) New() uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.next == math.MaxUint64 {
		panic("testkit UUID source exhausted")
	}
	s.next++

	var sequence [8]byte
	binary.BigEndian.PutUint64(sequence[:], s.next)
	return uuid.NewSHA1(s.namespace, sequence[:])
}

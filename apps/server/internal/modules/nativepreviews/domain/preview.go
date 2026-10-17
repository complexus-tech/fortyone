package nativepreviewdomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("preview not found")
	ErrForbidden     = errors.New("preview access denied")
	ErrInvalidTarget = errors.New("invalid preview target")
)

type Target struct {
	Kind     string
	ID       uuid.UUID
	TeamID   uuid.UUID
	TeamCode string
	Sequence int32
}

// Snapshot carries no access decisions. The repository authorizes the actor
// before returning either a revision or its small presentation payload.
type Snapshot struct {
	ID      uuid.UUID
	Version string
	Value   json.RawMessage
}

type Preview struct {
	Kind        string          `json:"kind"`
	Value       json.RawMessage `json:"value"`
	Version     string          `json:"version"`
	RefreshedAt time.Time       `json:"refreshedAt"`
}

package states

import (
	"time"

	"github.com/google/uuid"
)

type CoreState struct {
	ID         uuid.UUID
	Name       string
	Color      string
	Category   string
	OrderIndex int
	Team       uuid.UUID
	Workspace  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CoreNewState struct {
	Name     string
	Color    string
	Category string
	Team     uuid.UUID
}

type CoreUpdateState struct {
	Name       *string
	OrderIndex *int
}

package states

import (
	"time"

	"github.com/google/uuid"
)

type CoreState struct {
	ID         uuid.UUID
	Name       string
	Category   string
	OrderIndex int
	Workflow   uuid.UUID
	Workspace  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CoreNewState struct {
	Name     string
	Category string
	Workflow uuid.UUID
}

type CoreUpdateState struct {
	Name       *string
	OrderIndex *int
}

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
	Team       uuid.UUID
	Workspace  uuid.UUID
	IsDefault  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CoreNewState struct {
	Name      string
	Category  string
	Team      uuid.UUID
	IsDefault bool
}

type CoreUpdateState struct {
	Name       *string
	OrderIndex *int
	IsDefault  *bool
}

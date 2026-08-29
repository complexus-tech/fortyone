package statesdomain

import (
	"time"

	"github.com/google/uuid"
)

type State struct {
	ID         uuid.UUID
	Name       string
	Category   string
	OrderIndex int
	Team       uuid.UUID
	Workspace  uuid.UUID
	IsDefault  bool
	Color      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type NewState struct {
	Name      string
	Category  string
	Team      uuid.UUID
	IsDefault bool
	Color     string
}

type UpdateState struct {
	Name       *string
	OrderIndex *int
	IsDefault  *bool
	Color      *string
}

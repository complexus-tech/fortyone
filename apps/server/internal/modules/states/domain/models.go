package statesdomain

import (
	"time"

	"github.com/google/uuid"
)

type State struct {
	WIPLimit    *int
	ActiveCount int
	ID          uuid.UUID
	Name        string
	Category    string
	OrderIndex  int
	Team        uuid.UUID
	Workspace   uuid.UUID
	IsDefault   bool
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewState struct {
	Name      string
	Category  string
	Team      uuid.UUID
	IsDefault bool
	Color     string
}

type UpdateState struct {
	// Zero removes the advisory limit; nil leaves it unchanged.
	WIPLimit   *int
	Name       *string
	OrderIndex *int
	IsDefault  *bool
	Color      *string
}

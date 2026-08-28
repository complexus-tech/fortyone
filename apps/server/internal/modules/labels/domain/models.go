package labelsdomain

import (
	"time"

	"github.com/google/uuid"
)

type Label struct {
	ID          uuid.UUID
	Name        string
	TeamID      *uuid.UUID
	WorkspaceID uuid.UUID
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewLabel struct {
	Name        string
	TeamID      *uuid.UUID
	WorkspaceID uuid.UUID
	Color       string
}

// Filters is intentionally typed so callers cannot introduce SQL fields.
type Filters struct {
	TeamID *uuid.UUID
	Search string
	Limit  *int
	Offset int
}

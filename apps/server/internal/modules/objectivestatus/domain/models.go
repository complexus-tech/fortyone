package objectivestatusdomain

import (
	"time"

	"github.com/google/uuid"
)

type Status struct {
	ID          uuid.UUID
	Name        string
	Category    string
	OrderIndex  int
	WorkspaceID uuid.UUID
	IsDefault   bool
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewStatus struct {
	Name      string
	Category  string
	IsDefault bool
	Color     string
}

type UpdateStatus struct {
	Name       *string
	OrderIndex *int
	IsDefault  *bool
	Color      *string
}

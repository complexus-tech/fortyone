package objectivestatus

import (
	"time"

	"github.com/google/uuid"
)

type CoreObjectiveStatus struct {
	ID         uuid.UUID
	Name       string
	Category   string
	OrderIndex int
	Workspace  uuid.UUID
	IsDefault  bool
	Color      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CoreNewObjectiveStatus struct {
	Name      string
	Category  string
	IsDefault bool
	Color     string
}

type CoreUpdateObjectiveStatus struct {
	Name       *string
	OrderIndex *int
	IsDefault  *bool
	Color      *string
}

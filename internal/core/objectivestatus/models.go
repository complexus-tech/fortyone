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
	Workflow   uuid.UUID
	Workspace  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CoreNewObjectiveStatus struct {
	Name     string
	Category string
	Workflow uuid.UUID
}

type CoreUpdateObjectiveStatus struct {
	Name       *string
	OrderIndex *int
}

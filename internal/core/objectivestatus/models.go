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
	Team       uuid.UUID
	Workspace  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CoreNewObjectiveStatus struct {
	Name     string
	Category string
	Team     uuid.UUID
}

type CoreUpdateObjectiveStatus struct {
	Name       *string
	OrderIndex *int
}

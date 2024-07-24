package objectives

import (
	"time"

	"github.com/google/uuid"
)

type CoreObjective struct {
	ID          uuid.UUID
	Name        string
	Description *string
	LeadUser    *uuid.UUID
	Team        uuid.UUID
	Workspace   uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	IsPrivate   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

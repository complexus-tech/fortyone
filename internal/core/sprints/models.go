package sprints

import (
	"time"

	"github.com/google/uuid"
)

type CoreSprint struct {
	ID               uuid.UUID
	Name             string
	Goal             *string
	Objective        *uuid.UUID
	Team             uuid.UUID
	Workspace        uuid.UUID
	StartDate        time.Time
	EndDate          time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	TotalStories     int
	CancelledStories int
	CompletedStories int
	StartedStories   int
	UnstartedStories int
	BacklogStories   int
}

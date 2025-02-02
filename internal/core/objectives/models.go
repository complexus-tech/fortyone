package objectives

import (
	"time"

	"github.com/google/uuid"
)

type CoreObjective struct {
	ID               uuid.UUID
	Name             string
	Description      *string
	LeadUser         *uuid.UUID
	Team             uuid.UUID
	Workspace        uuid.UUID
	StartDate        *time.Time
	EndDate          *time.Time
	IsPrivate        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Status           uuid.UUID
	Priority         *string
	TotalStories     int
	CancelledStories int
	CompletedStories int
	StartedStories   int
	UnstartedStories int
	BacklogStories   int
}

type CoreNewObjective struct {
	Name        string
	Description *string
	LeadUser    *uuid.UUID
	Team        uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	IsPrivate   bool
	Status      uuid.UUID
	Priority    *string
}

type CoreUpdateObjective struct {
	Name        *string
	Description *string
	LeadUser    *uuid.UUID
	Team        *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	IsPrivate   *bool
	Visibility  *string
	Status      *uuid.UUID
	Priority    *string
}

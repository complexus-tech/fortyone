package objectives

import (
	"time"

	"github.com/google/uuid"
)

// ObjectiveHealth represents the possible health states of an objective
type ObjectiveHealth string

const (
	HealthAtRisk   ObjectiveHealth = "At Risk"
	HealthOnTrack  ObjectiveHealth = "On Track"
	HealthOffTrack ObjectiveHealth = "Off Track"
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
	CreatedBy        uuid.UUID
	Priority         *string
	Health           *ObjectiveHealth
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
	CreatedBy   uuid.UUID
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
	Health      *ObjectiveHealth
}

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

// Objective Analytics Models

type CoreObjectiveAnalytics struct {
	ObjectiveID       uuid.UUID
	PriorityBreakdown []CorePriorityBreakdown
	ProgressBreakdown CoreProgressBreakdown
	TeamAllocation    []CoreTeamMemberAllocation
	ProgressChart     []CoreObjectiveProgressDataPoint
}

type CorePriorityBreakdown struct {
	Priority string `db:"priority"`
	Count    int    `db:"count"`
}

type CoreProgressBreakdown struct {
	Total      int `db:"total"`
	Completed  int `db:"completed"`
	InProgress int `db:"in_progress"`
	Todo       int `db:"todo"`
	Blocked    int `db:"blocked"`
	Cancelled  int `db:"cancelled"`
}

type CoreTeamMemberAllocation struct {
	MemberID  uuid.UUID `db:"user_id"`
	Username  string    `db:"username"`
	AvatarURL *string   `db:"avatar_url"`
	Assigned  int       `db:"assigned"`
	Completed int       `db:"completed"`
}

type CoreObjectiveProgressDataPoint struct {
	Date       time.Time `json:"date" db:"completion_date"`
	Completed  int       `json:"completed" db:"stories_completed"`
	InProgress int       `json:"inProgress" db:"stories_in_progress"`
	Total      int       `json:"total" db:"total_stories"`
}

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

type CoreNewSprint struct {
	Name      string
	Goal      *string
	Objective *uuid.UUID
	Team      uuid.UUID
	Workspace uuid.UUID
	StartDate time.Time
	EndDate   time.Time
}

type CoreUpdateSprint struct {
	Name      *string
	Goal      *string
	Objective *uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
}

// Sprint Analytics Models

type CoreSprintAnalytics struct {
	SprintID         uuid.UUID
	Overview         CoreSprintOverview
	StoryBreakdown   CoreStoryBreakdown
	Burndown         []CoreBurndownDataPoint
	TeamAllocation   []CoreTeamMemberAllocation
	HealthIndicators CoreSprintHealthIndicators
}

type CoreSprintOverview struct {
	CompletionPercentage int
	DaysElapsed          int
	DaysRemaining        int
	Status               string // "on_track", "at_risk", "behind"
}

type CoreStoryBreakdown struct {
	Total      int `db:"total"`
	Completed  int `db:"completed"`
	InProgress int `db:"in_progress"`
	Todo       int `db:"todo"`
	Blocked    int `db:"blocked"`
	Cancelled  int `db:"cancelled"`
}

type CoreBurndownDataPoint struct {
	Date      time.Time
	Remaining int
}

type CoreTeamMemberAllocation struct {
	MemberID  uuid.UUID
	Username  string
	AvatarURL string
	Assigned  int
	Completed int
}

type CoreSprintHealthIndicators struct {
	BlockedCount   int `db:"blocked_count"`
	OverdueCount   int `db:"overdue_count"`
	AddedMidSprint int `db:"added_mid_sprint"`
}

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
	SprintID       uuid.UUID
	Overview       CoreSprintOverview
	StoryBreakdown CoreStoryBreakdown
	Burndown       []CoreBurndownDataPoint
	TeamAllocation []CoreTeamMemberAllocation
}

type CoreSprintOverview struct {
	CompletionPercentage int
	DaysElapsed          int
	DaysRemaining        int
	Status               string // "not_started", "on_track", "at_risk", "behind", "completed"
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
	Date      time.Time `db:"date"`
	Remaining int       `db:"remaining"`
	Ideal     int       `db:"ideal"`
}

type CoreTeamMemberAllocation struct {
	MemberID  uuid.UUID `db:"user_id"`
	Username  string    `db:"username"`
	AvatarURL string    `db:"avatar_url"`
	Assigned  int       `db:"assigned"`
	Completed int       `db:"completed"`
}

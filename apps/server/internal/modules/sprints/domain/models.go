// Package domain owns sprint concepts independently from HTTP, SQLC, pgx, and
// other modules' implementation details.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Sprint is a planning interval and its current story summary.
type Sprint struct {
	ID                          uuid.UUID
	Name                        string
	Goal                        *string
	ObjectiveID                 *uuid.UUID
	TeamID                      uuid.UUID
	WorkspaceID                 uuid.UUID
	StartDate                   time.Time
	EndDate                     time.Time
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	ScheduleManagedByAutomation bool
	TotalStories                int
	CancelledStories            int
	CompletedStories            int
	StartedStories              int
	UnstartedStories            int
	BacklogStories              int
}

// Analytics combines progress, burndown, and allocation for one sprint.
type Analytics struct {
	SprintID       uuid.UUID
	WorkingDays    []int
	Overview       Overview
	StoryBreakdown StoryBreakdown
	Burndown       []BurndownDataPoint
	TeamAllocation []TeamMemberAllocation
}

type Overview struct {
	CompletionPercentage int
	DaysElapsed          int
	DaysRemaining        int
	Status               string
}

type StoryBreakdown struct {
	Total      int
	Completed  int
	InProgress int
	Todo       int
	Blocked    int
	Cancelled  int
}

type BurndownDataPoint struct {
	Date      time.Time
	Remaining int
	Ideal     int
}

type TeamMemberAllocation struct {
	MemberID  uuid.UUID
	Username  string
	AvatarURL string
	Assigned  int
	Completed int
}

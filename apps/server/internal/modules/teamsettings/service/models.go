package teamsettings

import (
	"time"

	"github.com/google/uuid"
)

type CoreTeamSprintSettings struct {
	TeamID                       uuid.UUID
	WorkspaceID                  uuid.UUID
	AutoCreateSprints            bool
	UpcomingSprintsCount         int
	SprintDurationWeeks          int
	SprintStartDay               string
	WorkingDays                  []int
	MoveIncompleteStoriesEnabled bool
	LastAutoSprintNumber         int
	NextAutoSprintNumber         int
	AutoCreateDisabledAt         *time.Time
	AutoCreateDisabledReason     *string
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

type CoreTeamStoryAutomationSettings struct {
	TeamID                   uuid.UUID
	WorkspaceID              uuid.UUID
	AutoCloseInactiveEnabled bool
	AutoCloseInactiveMonths  int
	AutoArchiveEnabled       bool
	AutoArchiveMonths        int
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type CoreTeamEstimationSettings struct {
	TeamID      uuid.UUID
	WorkspaceID uuid.UUID
	Scheme      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CoreTeamSettings struct {
	SprintSettings          CoreTeamSprintSettings
	StoryAutomationSettings CoreTeamStoryAutomationSettings
	EstimationSettings      CoreTeamEstimationSettings
}

type CoreUpdateTeamSprintSettings struct {
	AutoCreateSprints            *bool
	UpcomingSprintsCount         *int
	SprintDurationWeeks          *int
	SprintStartDay               *string
	WorkingDays                  *[]int
	MoveIncompleteStoriesEnabled *bool
	NextAutoSprintNumber         *int
}

type CoreUpdateTeamStoryAutomationSettings struct {
	AutoCloseInactiveEnabled *bool
	AutoCloseInactiveMonths  *int
	AutoArchiveEnabled       *bool
	AutoArchiveMonths        *int
}

type CoreUpdateTeamEstimationSettings struct {
	Scheme *string
}

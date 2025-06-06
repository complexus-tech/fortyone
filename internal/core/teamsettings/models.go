package teamsettings

import (
	"time"

	"github.com/google/uuid"
)

type CoreTeamSprintSettings struct {
	TeamID                       uuid.UUID
	WorkspaceID                  uuid.UUID
	SprintsEnabled               bool
	AutoCreateSprints            bool
	UpcomingSprintsCount         int
	SprintDurationWeeks          int
	SprintStartDay               string
	MoveIncompleteStoriesEnabled bool
	LastAutoSprintNumber         int
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

type CoreTeamSettings struct {
	SprintSettings          CoreTeamSprintSettings
	StoryAutomationSettings CoreTeamStoryAutomationSettings
}

type CoreUpdateTeamSprintSettings struct {
	SprintsEnabled               *bool
	AutoCreateSprints            *bool
	UpcomingSprintsCount         *int
	SprintDurationWeeks          *int
	SprintStartDay               *string
	MoveIncompleteStoriesEnabled *bool
}

type CoreUpdateTeamStoryAutomationSettings struct {
	AutoCloseInactiveEnabled *bool
	AutoCloseInactiveMonths  *int
	AutoArchiveEnabled       *bool
	AutoArchiveMonths        *int
}

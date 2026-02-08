package teamsettingsrepository

import (
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	"github.com/google/uuid"
)

type dbTeamSprintSettings struct {
	TeamID                       uuid.UUID `db:"team_id"`
	WorkspaceID                  uuid.UUID `db:"workspace_id"`
	AutoCreateSprints            bool      `db:"auto_create_sprints"`
	UpcomingSprintsCount         int       `db:"upcoming_sprints_count"`
	SprintDurationWeeks          int       `db:"sprint_duration_weeks"`
	SprintStartDay               string    `db:"sprint_start_day"`
	MoveIncompleteStoriesEnabled bool      `db:"move_incomplete_stories_enabled"`
	LastAutoSprintNumber         int       `db:"last_auto_sprint_number"`
	CreatedAt                    time.Time `db:"created_at"`
	UpdatedAt                    time.Time `db:"updated_at"`
}

type dbTeamStoryAutomationSettings struct {
	TeamID                   uuid.UUID `db:"team_id"`
	WorkspaceID              uuid.UUID `db:"workspace_id"`
	AutoCloseInactiveEnabled bool      `db:"auto_close_inactive_enabled"`
	AutoCloseInactiveMonths  int       `db:"auto_close_inactive_months"`
	AutoArchiveEnabled       bool      `db:"auto_archive_enabled"`
	AutoArchiveMonths        int       `db:"auto_archive_months"`
	CreatedAt                time.Time `db:"created_at"`
	UpdatedAt                time.Time `db:"updated_at"`
}

func toCoreTeamSprintSettings(s dbTeamSprintSettings) teamsettings.CoreTeamSprintSettings {
	return teamsettings.CoreTeamSprintSettings{
		TeamID:                       s.TeamID,
		WorkspaceID:                  s.WorkspaceID,
		AutoCreateSprints:            s.AutoCreateSprints,
		UpcomingSprintsCount:         s.UpcomingSprintsCount,
		SprintDurationWeeks:          s.SprintDurationWeeks,
		SprintStartDay:               s.SprintStartDay,
		MoveIncompleteStoriesEnabled: s.MoveIncompleteStoriesEnabled,
		LastAutoSprintNumber:         s.LastAutoSprintNumber,
		CreatedAt:                    s.CreatedAt,
		UpdatedAt:                    s.UpdatedAt,
	}
}

func toCoreTeamSprintSettingsList(settings []dbTeamSprintSettings) []teamsettings.CoreTeamSprintSettings {
	result := make([]teamsettings.CoreTeamSprintSettings, len(settings))
	for i, setting := range settings {
		result[i] = toCoreTeamSprintSettings(setting)
	}
	return result
}

func toCoreTeamStoryAutomationSettings(s dbTeamStoryAutomationSettings) teamsettings.CoreTeamStoryAutomationSettings {
	return teamsettings.CoreTeamStoryAutomationSettings{
		TeamID:                   s.TeamID,
		WorkspaceID:              s.WorkspaceID,
		AutoCloseInactiveEnabled: s.AutoCloseInactiveEnabled,
		AutoCloseInactiveMonths:  s.AutoCloseInactiveMonths,
		AutoArchiveEnabled:       s.AutoArchiveEnabled,
		AutoArchiveMonths:        s.AutoArchiveMonths,
		CreatedAt:                s.CreatedAt,
		UpdatedAt:                s.UpdatedAt,
	}
}

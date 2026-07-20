package teamsettingsrepository

import (
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type dbTeamSprintSettings struct {
	TeamID                       uuid.UUID     `db:"team_id"`
	WorkspaceID                  uuid.UUID     `db:"workspace_id"`
	AutoCreateSprints            bool          `db:"auto_create_sprints"`
	UpcomingSprintsCount         int           `db:"upcoming_sprints_count"`
	SprintDurationWeeks          int           `db:"sprint_duration_weeks"`
	SprintStartDay               string        `db:"sprint_start_day"`
	WorkingDays                  pq.Int64Array `db:"working_days"`
	MoveIncompleteStoriesEnabled bool          `db:"move_incomplete_stories_enabled"`
	LastAutoSprintNumber         int           `db:"last_auto_sprint_number"`
	NextAutoSprintNumber         int           `db:"next_auto_sprint_number"`
	AutoCreateDisabledAt         *time.Time    `db:"auto_create_disabled_at"`
	AutoCreateDisabledReason     *string       `db:"auto_create_disabled_reason"`
	CreatedAt                    time.Time     `db:"created_at"`
	UpdatedAt                    time.Time     `db:"updated_at"`
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

type dbTeamEstimationSettings struct {
	TeamID      uuid.UUID `db:"team_id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	Scheme      string    `db:"scheme"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toCoreTeamSprintSettings(s dbTeamSprintSettings) teamsettings.CoreTeamSprintSettings {
	return teamsettings.CoreTeamSprintSettings{
		TeamID:                       s.TeamID,
		WorkspaceID:                  s.WorkspaceID,
		AutoCreateSprints:            s.AutoCreateSprints,
		UpcomingSprintsCount:         s.UpcomingSprintsCount,
		SprintDurationWeeks:          s.SprintDurationWeeks,
		SprintStartDay:               s.SprintStartDay,
		WorkingDays:                  int64sToInts(s.WorkingDays),
		MoveIncompleteStoriesEnabled: s.MoveIncompleteStoriesEnabled,
		LastAutoSprintNumber:         s.LastAutoSprintNumber,
		NextAutoSprintNumber:         s.NextAutoSprintNumber,
		AutoCreateDisabledAt:         s.AutoCreateDisabledAt,
		AutoCreateDisabledReason:     s.AutoCreateDisabledReason,
		CreatedAt:                    s.CreatedAt,
		UpdatedAt:                    s.UpdatedAt,
	}
}

func int64sToInts(values []int64) []int {
	result := make([]int, len(values))
	for i, value := range values {
		result[i] = int(value)
	}
	return result
}

func intsToInt64s(values []int) []int64 {
	result := make([]int64, len(values))
	for i, value := range values {
		result[i] = int64(value)
	}
	return result
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

func toCoreTeamEstimationSettings(s dbTeamEstimationSettings) teamsettings.CoreTeamEstimationSettings {
	return teamsettings.CoreTeamEstimationSettings{
		TeamID:      s.TeamID,
		WorkspaceID: s.WorkspaceID,
		Scheme:      s.Scheme,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

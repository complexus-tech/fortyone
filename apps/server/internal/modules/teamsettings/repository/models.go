package teamsettingsrepository

import (
	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
)

func mapGetSprintSettings(row teamsettingssql.GetSprintSettingsRow) teamsettings.CoreTeamSprintSettings {
	return newSprintSettings(
		row.TeamID, row.WorkspaceID, row.AutoCreateSprints,
		row.UpcomingSprintsCount, row.SprintDurationWeeks, row.SprintStartDay,
		row.WorkingDays, row.MoveIncompleteStoriesEnabled, row.LastAutoSprintNumber,
		row.NextAutoSprintNumber, row.AutoCreateDisabledAt, row.AutoCreateDisabledReason,
		row.CreatedAt, row.UpdatedAt,
	)
}

func mapLockedSprintSettings(row teamsettingssql.LockSprintSettingsRow) teamsettings.CoreTeamSprintSettings {
	return newSprintSettings(
		row.TeamID, row.WorkspaceID, row.AutoCreateSprints,
		row.UpcomingSprintsCount, row.SprintDurationWeeks, row.SprintStartDay,
		row.WorkingDays, row.MoveIncompleteStoriesEnabled, row.LastAutoSprintNumber,
		row.NextAutoSprintNumber, row.AutoCreateDisabledAt, row.AutoCreateDisabledReason,
		row.CreatedAt, row.UpdatedAt,
	)
}

func mapUpdatedSprintSettings(row teamsettingssql.UpdateSprintSettingsRow) teamsettings.CoreTeamSprintSettings {
	return newSprintSettings(
		row.TeamID, row.WorkspaceID, row.AutoCreateSprints,
		row.UpcomingSprintsCount, row.SprintDurationWeeks, row.SprintStartDay,
		row.WorkingDays, row.MoveIncompleteStoriesEnabled, row.LastAutoSprintNumber,
		row.NextAutoSprintNumber, row.AutoCreateDisabledAt, row.AutoCreateDisabledReason,
		row.CreatedAt, row.UpdatedAt,
	)
}

func mapStoryAutomationSettings(row teamsettingssql.TeamStoryAutomationSetting) teamsettings.CoreTeamStoryAutomationSettings {
	return teamsettings.CoreTeamStoryAutomationSettings{
		TeamID:                   row.TeamID,
		WorkspaceID:              row.WorkspaceID,
		AutoCloseInactiveEnabled: row.AutoCloseInactiveEnabled,
		AutoCloseInactiveMonths:  int(row.AutoCloseInactiveMonths),
		AutoArchiveEnabled:       row.AutoArchiveEnabled,
		AutoArchiveMonths:        int(row.AutoArchiveMonths),
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
	}
}

func mapEstimationSettings(row teamsettingssql.TeamEstimationSetting) teamsettings.CoreTeamEstimationSettings {
	return teamsettings.CoreTeamEstimationSettings{
		TeamID:      row.TeamID,
		WorkspaceID: row.WorkspaceID,
		Scheme:      row.Scheme,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

package teamsettingsrepository

import (
	"context"
	"fmt"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

func (r *repo) UpdateSprintSettings(
	ctx context.Context,
	teamID, workspaceID uuid.UUID,
	updates teamsettings.CoreUpdateTeamSprintSettings,
	actor teamsettings.AuditActor,
) (teamsettings.CoreTeamSprintSettings, error) {
	if updates.Empty() {
		return teamsettings.CoreTeamSprintSettings{}, teamsettings.ErrNoSettingsChanges
	}
	params, err := sprintUpdateParams(teamID, workspaceID, updates)
	if err != nil {
		return teamsettings.CoreTeamSprintSettings{}, err
	}
	var updated teamsettings.CoreTeamSprintSettings
	err = r.withinTransaction(ctx, func(queries teamsettingssql.Querier) error {
		if _, err := queries.EnsureSprintSettings(ctx, teamsettingssql.EnsureSprintSettingsParams{
			TeamID: teamID, WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("ensure sprint settings: %w", err)
		}
		row, err := queries.UpdateSprintSettings(ctx, params)
		if err != nil {
			return fmt.Errorf("update sprint settings: %w", err)
		}
		updated = mapUpdatedSprintSettings(row)
		if cadenceChanged(updates) && updated.AutoCreateSprints {
			if _, err := reconcileSprintSchedule(ctx, queries, updated, actor); err != nil {
				return err
			}
		}
		return insertAuditEvent(
			ctx, queries, workspaceID, teamID, actor,
			"team_settings", teamID, "team_settings.sprint_updated",
			settingsAuditMetadata{ChangedFields: sprintChangedFields(updates)},
		)
	})
	if err != nil {
		return teamsettings.CoreTeamSprintSettings{}, err
	}
	return updated, nil
}

func sprintUpdateParams(
	teamID, workspaceID uuid.UUID,
	updates teamsettings.CoreUpdateTeamSprintSettings,
) (teamsettingssql.UpdateSprintSettingsParams, error) {
	upcomingSprintsCount, err := safecast.Int32(updates.UpcomingSprintsCount.Value)
	if err != nil {
		return teamsettingssql.UpdateSprintSettingsParams{}, teamsettings.ErrInvalidUpcomingCount
	}
	sprintDurationWeeks, err := safecast.Int32(updates.SprintDurationWeeks.Value)
	if err != nil {
		return teamsettingssql.UpdateSprintSettingsParams{}, teamsettings.ErrInvalidSprintDuration
	}
	workingDays, err := intsToSmallInts(updates.WorkingDays.Value)
	if err != nil {
		return teamsettingssql.UpdateSprintSettingsParams{}, err
	}
	nextAutoSprintNumber, err := safecast.Int32(updates.NextAutoSprintNumber.Value)
	if err != nil {
		return teamsettingssql.UpdateSprintSettingsParams{}, teamsettings.ErrInvalidNextAutoNumber
	}
	return teamsettingssql.UpdateSprintSettingsParams{
		SetAutoCreateSprints:            updates.AutoCreateSprints.Present,
		AutoCreateSprints:               updates.AutoCreateSprints.Value,
		SetUpcomingSprintsCount:         updates.UpcomingSprintsCount.Present,
		UpcomingSprintsCount:            upcomingSprintsCount,
		SetSprintDurationWeeks:          updates.SprintDurationWeeks.Present,
		SprintDurationWeeks:             sprintDurationWeeks,
		SetSprintStartDay:               updates.SprintStartDay.Present,
		SprintStartDay:                  updates.SprintStartDay.Value,
		SetWorkingDays:                  updates.WorkingDays.Present,
		WorkingDays:                     workingDays,
		SetMoveIncompleteStoriesEnabled: updates.MoveIncompleteStoriesEnabled.Present,
		MoveIncompleteStoriesEnabled:    updates.MoveIncompleteStoriesEnabled.Value,
		SetNextAutoSprintNumber:         updates.NextAutoSprintNumber.Present,
		NextAutoSprintNumber:            nextAutoSprintNumber,
		TeamID:                          teamID,
		WorkspaceID:                     workspaceID,
	}, nil
}

func cadenceChanged(updates teamsettings.CoreUpdateTeamSprintSettings) bool {
	return updates.SprintDurationWeeks.Present || updates.SprintStartDay.Present
}

func sprintChangedFields(updates teamsettings.CoreUpdateTeamSprintSettings) []string {
	fields := make([]string, 0, 7)
	if updates.AutoCreateSprints.Present {
		fields = append(fields, "auto_create_sprints")
	}
	if updates.UpcomingSprintsCount.Present {
		fields = append(fields, "upcoming_sprints_count")
	}
	if updates.SprintDurationWeeks.Present {
		fields = append(fields, "sprint_duration_weeks")
	}
	if updates.SprintStartDay.Present {
		fields = append(fields, "sprint_start_day")
	}
	if updates.WorkingDays.Present {
		fields = append(fields, "working_days")
	}
	if updates.MoveIncompleteStoriesEnabled.Present {
		fields = append(fields, "move_incomplete_stories_enabled")
	}
	if updates.NextAutoSprintNumber.Present {
		fields = append(fields, "next_auto_sprint_number")
	}
	return fields
}

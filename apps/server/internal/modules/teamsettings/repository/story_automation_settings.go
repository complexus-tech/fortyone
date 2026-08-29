package teamsettingsrepository

import (
	"context"
	"fmt"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

func (r *repo) UpdateStoryAutomationSettings(
	ctx context.Context,
	teamID, workspaceID uuid.UUID,
	updates teamsettings.CoreUpdateTeamStoryAutomationSettings,
	actor teamsettings.AuditActor,
) (teamsettings.CoreTeamStoryAutomationSettings, error) {
	if updates.Empty() {
		return teamsettings.CoreTeamStoryAutomationSettings{}, teamsettings.ErrNoSettingsChanges
	}
	params, err := storyAutomationUpdateParams(teamID, workspaceID, updates)
	if err != nil {
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}
	var updated teamsettings.CoreTeamStoryAutomationSettings
	err = r.withinTransaction(ctx, func(queries teamsettingssql.Querier) error {
		if _, err := queries.EnsureStoryAutomationSettings(ctx, teamsettingssql.EnsureStoryAutomationSettingsParams{
			TeamID: teamID, WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("ensure story automation settings: %w", err)
		}
		row, err := queries.UpdateStoryAutomationSettings(ctx, params)
		if err != nil {
			return fmt.Errorf("update story automation settings: %w", err)
		}
		updated = mapStoryAutomationSettings(row)
		return insertAuditEvent(
			ctx, queries, workspaceID, teamID, actor,
			"team_settings", teamID, "team_settings.story_automation_updated",
			settingsAuditMetadata{ChangedFields: storyChangedFields(updates)},
		)
	})
	if err != nil {
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}
	return updated, nil
}

func storyAutomationUpdateParams(
	teamID, workspaceID uuid.UUID,
	updates teamsettings.CoreUpdateTeamStoryAutomationSettings,
) (teamsettingssql.UpdateStoryAutomationSettingsParams, error) {
	autoCloseMonths, err := safecast.Int32(updates.AutoCloseInactiveMonths.Value)
	if err != nil {
		return teamsettingssql.UpdateStoryAutomationSettingsParams{}, teamsettings.ErrInvalidCloseMonths
	}
	autoArchiveMonths, err := safecast.Int32(updates.AutoArchiveMonths.Value)
	if err != nil {
		return teamsettingssql.UpdateStoryAutomationSettingsParams{}, teamsettings.ErrInvalidArchiveMonths
	}
	return teamsettingssql.UpdateStoryAutomationSettingsParams{
		SetAutoCloseInactiveEnabled: updates.AutoCloseInactiveEnabled.Present,
		AutoCloseInactiveEnabled:    updates.AutoCloseInactiveEnabled.Value,
		SetAutoCloseInactiveMonths:  updates.AutoCloseInactiveMonths.Present,
		AutoCloseInactiveMonths:     autoCloseMonths,
		SetAutoArchiveEnabled:       updates.AutoArchiveEnabled.Present,
		AutoArchiveEnabled:          updates.AutoArchiveEnabled.Value,
		SetAutoArchiveMonths:        updates.AutoArchiveMonths.Present,
		AutoArchiveMonths:           autoArchiveMonths,
		TeamID:                      teamID,
		WorkspaceID:                 workspaceID,
	}, nil
}

func storyChangedFields(updates teamsettings.CoreUpdateTeamStoryAutomationSettings) []string {
	fields := make([]string, 0, 4)
	if updates.AutoCloseInactiveEnabled.Present {
		fields = append(fields, "auto_close_inactive_enabled")
	}
	if updates.AutoCloseInactiveMonths.Present {
		fields = append(fields, "auto_close_inactive_months")
	}
	if updates.AutoArchiveEnabled.Present {
		fields = append(fields, "auto_archive_enabled")
	}
	if updates.AutoArchiveMonths.Present {
		fields = append(fields, "auto_archive_months")
	}
	return fields
}

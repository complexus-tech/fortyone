package teamsettingsrepository

import (
	"context"
	"fmt"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) UpdateEstimationSettings(
	ctx context.Context,
	teamID, workspaceID uuid.UUID,
	updates teamsettings.CoreUpdateTeamEstimationSettings,
	actor teamsettings.AuditActor,
) (teamsettings.CoreTeamEstimationSettings, error) {
	if updates.Empty() {
		return teamsettings.CoreTeamEstimationSettings{}, teamsettings.ErrNoSettingsChanges
	}
	var updated teamsettings.CoreTeamEstimationSettings
	err := r.withinTransaction(ctx, func(queries teamsettingssql.Querier) error {
		if _, err := queries.EnsureEstimationSettings(ctx, teamsettingssql.EnsureEstimationSettingsParams{
			TeamID: teamID, WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("ensure estimation settings: %w", err)
		}
		row, err := queries.UpdateEstimationSettings(ctx, teamsettingssql.UpdateEstimationSettingsParams{
			SetScheme:   updates.Scheme.Present,
			Scheme:      updates.Scheme.Value,
			TeamID:      teamID,
			WorkspaceID: workspaceID,
		})
		if err != nil {
			return fmt.Errorf("update estimation settings: %w", err)
		}
		updated = mapEstimationSettings(row)
		return insertAuditEvent(
			ctx, queries, workspaceID, teamID, actor,
			"team_settings", teamID, "team_settings.estimation_updated",
			settingsAuditMetadata{ChangedFields: []string{"scheme"}},
		)
	})
	if err != nil {
		return teamsettings.CoreTeamEstimationSettings{}, err
	}
	return updated, nil
}

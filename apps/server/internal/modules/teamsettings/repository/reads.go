package teamsettingsrepository

import (
	"context"
	"fmt"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) IsActiveTeamMember(
	ctx context.Context,
	teamID, workspaceID, actorID uuid.UUID,
) (bool, error) {
	isMember, err := r.queries.IsActiveTeamMember(ctx, teamsettingssql.IsActiveTeamMemberParams{
		TeamID: teamID, WorkspaceID: workspaceID, ActorID: actorID,
	})
	if err != nil {
		return false, fmt.Errorf("check active team membership: %w", mapDatabaseError(err))
	}
	return isMember, nil
}

func (r *repo) GetSprintSettings(
	ctx context.Context,
	teamID, workspaceID uuid.UUID,
) (teamsettings.CoreTeamSprintSettings, error) {
	params := teamsettingssql.EnsureSprintSettingsParams{TeamID: teamID, WorkspaceID: workspaceID}
	if _, err := r.queries.EnsureSprintSettings(ctx, params); err != nil {
		return teamsettings.CoreTeamSprintSettings{}, fmt.Errorf("ensure sprint settings: %w", mapDatabaseError(err))
	}
	row, err := r.queries.GetSprintSettings(ctx, teamsettingssql.GetSprintSettingsParams(params))
	if err != nil {
		return teamsettings.CoreTeamSprintSettings{}, fmt.Errorf("get sprint settings: %w", mapDatabaseError(err))
	}
	return mapGetSprintSettings(row), nil
}

func (r *repo) GetStoryAutomationSettings(
	ctx context.Context,
	teamID, workspaceID uuid.UUID,
) (teamsettings.CoreTeamStoryAutomationSettings, error) {
	params := teamsettingssql.EnsureStoryAutomationSettingsParams{TeamID: teamID, WorkspaceID: workspaceID}
	if _, err := r.queries.EnsureStoryAutomationSettings(ctx, params); err != nil {
		return teamsettings.CoreTeamStoryAutomationSettings{}, fmt.Errorf("ensure story automation settings: %w", mapDatabaseError(err))
	}
	row, err := r.queries.GetStoryAutomationSettings(ctx, teamsettingssql.GetStoryAutomationSettingsParams(params))
	if err != nil {
		return teamsettings.CoreTeamStoryAutomationSettings{}, fmt.Errorf("get story automation settings: %w", mapDatabaseError(err))
	}
	return mapStoryAutomationSettings(row), nil
}

func (r *repo) GetEstimationSettings(
	ctx context.Context,
	teamID, workspaceID uuid.UUID,
) (teamsettings.CoreTeamEstimationSettings, error) {
	params := teamsettingssql.EnsureEstimationSettingsParams{TeamID: teamID, WorkspaceID: workspaceID}
	if _, err := r.queries.EnsureEstimationSettings(ctx, params); err != nil {
		return teamsettings.CoreTeamEstimationSettings{}, fmt.Errorf("ensure estimation settings: %w", mapDatabaseError(err))
	}
	row, err := r.queries.GetEstimationSettings(ctx, teamsettingssql.GetEstimationSettingsParams(params))
	if err != nil {
		return teamsettings.CoreTeamEstimationSettings{}, fmt.Errorf("get estimation settings: %w", mapDatabaseError(err))
	}
	return mapEstimationSettings(row), nil
}

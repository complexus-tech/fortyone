package teamsrepository

import (
	"context"
	"errors"
	"fmt"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	teamsql "github.com/complexus-tech/projects-api/internal/modules/teams/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) Create(ctx context.Context, team teamsdomain.Team) (teamsdomain.Team, error) {
	var created teamsdomain.Team
	err := r.withinTransaction(ctx, func(queries teamsql.Querier) error {
		result, err := createTeamWithDefaults(ctx, queries, team)
		if err != nil {
			return err
		}
		created = result
		return nil
	})
	if err != nil {
		return teamsdomain.Team{}, err
	}
	return created, nil
}

func createTeamWithDefaults(
	ctx context.Context,
	queries teamsql.Querier,
	team teamsdomain.Team,
) (teamsdomain.Team, error) {
	row, err := queries.CreateTeam(ctx, teamsql.CreateTeamParams{
		Name:        team.Name,
		Code:        team.Code,
		Color:       team.Color,
		IsPrivate:   team.IsPrivate,
		WorkspaceID: team.Workspace,
	})
	if err != nil {
		if database.Classify(err) == database.ErrorClassUniqueViolation {
			return teamsdomain.Team{}, teamsdomain.ErrCodeExists
		}
		return teamsdomain.Team{}, fmt.Errorf("create team: %w", err)
	}

	if err := createTeamDefaults(ctx, queries, row.TeamID, row.WorkspaceID); err != nil {
		return teamsdomain.Team{}, err
	}
	return toCoreCreatedTeam(row), nil
}

func (transaction *workspaceTransaction) CreateTeam(ctx context.Context, input WorkspaceTeamInput) (WorkspaceTeam, error) {
	team, err := createTeamWithDefaults(ctx, transaction.queries, teamsdomain.Team{
		Name:      input.Name,
		Code:      input.Code,
		Color:     input.Color,
		Workspace: input.Workspace,
	})
	if err != nil {
		return WorkspaceTeam{}, err
	}
	return WorkspaceTeam{ID: team.ID}, nil
}

func createTeamDefaults(
	ctx context.Context,
	queries teamsql.Querier,
	teamID uuid.UUID,
	workspaceID uuid.UUID,
) error {
	rowsAffected, err := queries.CreateDefaultStoryAutomationSettings(
		ctx,
		teamsql.CreateDefaultStoryAutomationSettingsParams{
			TeamID:      teamID,
			WorkspaceID: workspaceID,
		},
	)
	if err != nil {
		return fmt.Errorf("create default story automation settings: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("create default story automation settings: affected %d rows, want 1", rowsAffected)
	}

	for _, status := range teamsdomain.DefaultStoryStatuses {
		orderIndex, err := safecast.Int32(status.OrderIndex)
		if err != nil {
			return fmt.Errorf("convert default story status %q order: %w", status.Name, err)
		}
		rowsAffected, err := queries.CreateDefaultStoryStatus(ctx, teamsql.CreateDefaultStoryStatusParams{
			Name:        status.Name,
			Category:    status.Category,
			OrderIndex:  orderIndex,
			Color:       status.Color,
			TeamID:      teamID,
			WorkspaceID: workspaceID,
		})
		if err != nil {
			return fmt.Errorf("create default story status %q: %w", status.Name, err)
		}
		if rowsAffected != 1 {
			return fmt.Errorf("create default story status %q: affected %d rows, want 1", status.Name, rowsAffected)
		}
	}
	return nil
}

func (r *repo) Update(
	ctx context.Context,
	teamID uuid.UUID,
	updates teamsdomain.Team,
) (teamsdomain.Team, error) {
	row, err := r.queries.UpdateTeamForWorkspace(ctx, teamsql.UpdateTeamForWorkspaceParams{
		Name:        updates.Name,
		Code:        updates.Code,
		Color:       updates.Color,
		IsPrivate:   updates.IsPrivate,
		TeamID:      teamID,
		WorkspaceID: updates.Workspace,
	})
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return teamsdomain.Team{}, teamsdomain.ErrNotFound
		case database.Classify(err) == database.ErrorClassUniqueViolation:
			return teamsdomain.Team{}, teamsdomain.ErrCodeExists
		default:
			return teamsdomain.Team{}, fmt.Errorf("update scoped team: %w", err)
		}
	}
	return toCoreUpdatedTeam(row), nil
}

func (r *repo) Delete(ctx context.Context, teamID uuid.UUID, workspaceID uuid.UUID) error {
	rowsAffected, err := r.queries.DeleteTeamForWorkspace(ctx, teamsql.DeleteTeamForWorkspaceParams{
		TeamID:      teamID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("delete scoped team: %w", err)
	}
	if rowsAffected == 0 {
		return teamsdomain.ErrNotFound
	}
	return nil
}

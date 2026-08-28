package githubrepository

import (
	"context"
	"database/sql"
	"errors"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) CreateIssueSyncLink(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	input githubshared.CoreIssueSyncLinkInput,
) (githubshared.CoreIssueSyncLink, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return githubshared.CoreIssueSyncLink{}, err
	}
	id, err := queries.CreateGitHubIssueSyncLink(ctx, githubsql.CreateGitHubIssueSyncLinkParams{
		WorkspaceID:     workspaceID,
		SyncDirection:   input.SyncDirection,
		CreatedByUserID: &userID,
		TeamID:          input.TeamID,
		RepositoryID:    input.RepositoryID,
	})
	if err != nil {
		return githubshared.CoreIssueSyncLink{}, mapDatabaseError(err)
	}
	return r.GetIssueSyncLink(ctx, workspaceID, id)
}

func (r *Repo) GetIssueSyncLink(
	ctx context.Context,
	workspaceID, linkID uuid.UUID,
) (githubshared.CoreIssueSyncLink, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return githubshared.CoreIssueSyncLink{}, err
	}
	row, err := queries.GetGitHubIssueSyncLink(ctx, githubsql.GetGitHubIssueSyncLinkParams{
		WorkspaceID: workspaceID,
		LinkID:      linkID,
	})
	if err != nil {
		return githubshared.CoreIssueSyncLink{}, mapDatabaseError(err)
	}
	return githubshared.CoreIssueSyncLink{
		ID:             row.ID,
		RepositoryID:   row.RepositoryID,
		RepositoryName: row.RepositoryName,
		TeamID:         row.TeamID,
		TeamName:       row.TeamName,
		TeamColor:      row.TeamColor,
		SyncDirection:  row.SyncDirection,
		IsActive:       row.IsActive,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}

func (r *Repo) UpdateIssueSyncLink(
	ctx context.Context,
	workspaceID, linkID uuid.UUID,
	input githubshared.CoreUpdateIssueSyncLinkInput,
) (githubshared.CoreIssueSyncLink, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return githubshared.CoreIssueSyncLink{}, err
	}
	if _, err := queries.UpdateGitHubIssueSyncLink(ctx, githubsql.UpdateGitHubIssueSyncLinkParams{
		SyncDirection: input.SyncDirection,
		IsActive:      input.IsActive,
		WorkspaceID:   workspaceID,
		LinkID:        linkID,
	}); err != nil {
		return githubshared.CoreIssueSyncLink{}, mapDatabaseError(err)
	}
	return r.GetIssueSyncLink(ctx, workspaceID, linkID)
}

func (r *Repo) DeleteIssueSyncLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.DeleteGitHubIssueSyncLink(ctx, githubsql.DeleteGitHubIssueSyncLinkParams{
		WorkspaceID: workspaceID,
		LinkID:      linkID,
	})
}

func (r *Repo) GetTeamWorkflowSettings(
	ctx context.Context,
	workspaceID, teamID uuid.UUID,
) (githubshared.CoreTeamGitHubSettings, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return githubshared.CoreTeamGitHubSettings{}, err
	}
	rows, err := queries.ListGitHubTeamWorkflowRules(ctx, githubsql.ListGitHubTeamWorkflowRulesParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
	})
	if err != nil {
		return githubshared.CoreTeamGitHubSettings{}, err
	}
	rules := make([]githubshared.CoreWorkflowRule, 0, len(rows))
	for _, row := range rows {
		rules = append(rules, githubshared.CoreWorkflowRule{
			ID:                row.ID,
			EventKey:          row.EventKey,
			TargetStatusID:    row.TargetStatusID,
			BaseBranchPattern: row.BaseBranchPattern,
			IsActive:          row.IsActive,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
		})
	}
	return githubshared.CoreTeamGitHubSettings{TeamID: teamID, Rules: rules}, nil
}

func (r *Repo) ReplaceTeamWorkflowSettings(
	ctx context.Context,
	workspaceID, teamID uuid.UUID,
	rules []githubshared.CoreWorkflowRuleInput,
) (githubshared.CoreTeamGitHubSettings, error) {
	err := r.withinTransaction(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}, func(queries githubsql.Querier) error {
		if _, err := queries.LockGitHubTeam(ctx, githubsql.LockGitHubTeamParams{
			WorkspaceID: workspaceID,
			TeamID:      teamID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return sql.ErrNoRows
			}
			return err
		}
		if err := queries.DeleteGitHubTeamWorkflowRules(ctx, githubsql.DeleteGitHubTeamWorkflowRulesParams{
			WorkspaceID: workspaceID,
			TeamID:      teamID,
		}); err != nil {
			return err
		}
		for _, rule := range rules {
			if _, err := queries.InsertGitHubTeamWorkflowRule(ctx, githubsql.InsertGitHubTeamWorkflowRuleParams{
				WorkspaceID:       workspaceID,
				TeamID:            teamID,
				EventKey:          rule.EventKey,
				TargetStatusID:    rule.TargetStatusID,
				BaseBranchPattern: rule.BaseBranchPattern,
				IsActive:          rule.IsActive,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return githubshared.CoreTeamGitHubSettings{}, err
	}
	return r.GetTeamWorkflowSettings(ctx, workspaceID, teamID)
}

func (r *Repo) ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]statusRow, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	rows, err := queries.ListGitHubTeamStatuses(ctx, githubsql.ListGitHubTeamStatusesParams{TeamID: teamID})
	if err != nil {
		return nil, err
	}
	statuses := make([]statusRow, 0, len(rows))
	for _, row := range rows {
		statuses = append(statuses, statusRow{
			ID:       row.StatusID,
			Name:     row.Name,
			Category: row.Category,
			Color:    row.Color,
		})
	}
	return statuses, nil
}

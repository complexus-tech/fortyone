package teamsrepository

import (
	"context"
	"errors"
	"fmt"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	teamsql "github.com/complexus-tech/projects-api/internal/modules/teams/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

type DeletionTransaction interface {
	Lock(context.Context, teamsdomain.Deletion) error
	RemoveEntityReferences(context.Context, teamsdomain.Deletion) error
	Delete(context.Context, teamsdomain.Deletion) error
}

type deletionTransaction struct{ queries teamsql.Querier }

func (r *repo) BindDeletionTransaction(tx pgx.Tx) (DeletionTransaction, error) {
	queries, err := r.transactionQueries(tx)
	if err != nil {
		return nil, err
	}
	return &deletionTransaction{queries: queries}, nil
}

func (transaction *deletionTransaction) Lock(ctx context.Context, command teamsdomain.Deletion) error {
	_, err := transaction.queries.LockTeamForDeletion(ctx, teamsql.LockTeamForDeletionParams{
		TeamID: command.TeamID, WorkspaceID: command.WorkspaceID, ActorID: command.Actor.PrincipalID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return teamsdomain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock team for deletion: %w", err)
	}
	if err := transaction.queries.LockTeamDeletionObjectives(ctx, teamsql.LockTeamDeletionObjectivesParams{
		TeamID: &command.TeamID, WorkspaceID: &command.WorkspaceID,
	}); err != nil {
		return fmt.Errorf("lock team objectives: %w", err)
	}
	if err := transaction.queries.LockTeamDeletionBoards(ctx, teamsql.LockTeamDeletionBoardsParams{
		TeamID: command.TeamID, WorkspaceID: command.WorkspaceID,
	}); err != nil {
		return fmt.Errorf("lock team feedback boards: %w", err)
	}
	if err := transaction.queries.LockTeamDeletionFeedback(ctx, teamsql.LockTeamDeletionFeedbackParams{
		TeamID: command.TeamID, WorkspaceID: command.WorkspaceID,
	}); err != nil {
		return fmt.Errorf("lock team feedback: %w", err)
	}
	return nil
}

func (transaction *deletionTransaction) RemoveEntityReferences(ctx context.Context, command teamsdomain.Deletion) error {
	if err := transaction.queries.DeleteTeamDocumentRelationships(ctx, teamsql.DeleteTeamDocumentRelationshipsParams{
		TeamID: command.TeamID, WorkspaceID: command.WorkspaceID,
	}); err != nil {
		return fmt.Errorf("remove deleted team's document links: %w", err)
	}
	if err := transaction.queries.DeleteTeamEntityNotifications(ctx, teamsql.DeleteTeamEntityNotificationsParams{
		TeamID: command.TeamID, WorkspaceID: command.WorkspaceID,
	}); err != nil {
		return fmt.Errorf("remove deleted team's notifications: %w", err)
	}
	return nil
}

func (transaction *deletionTransaction) Delete(ctx context.Context, command teamsdomain.Deletion) error {
	return newWithQueries(transaction.queries).Delete(ctx, command.TeamID, command.WorkspaceID)
}

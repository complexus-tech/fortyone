package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LockTeamDeletionStories is a capability for the team deletion unit of work.
// Its caller holds the team lock and owns the surrounding transaction.
func (r *repo) LockTeamDeletionStories(ctx context.Context, tx pgx.Tx, teamID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	if tx == nil || teamID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, errors.New("team deletion transaction and scope are required")
	}
	ids, err := storyreadsql.New(tx).LockTeamDeletionStories(ctx, storyreadsql.LockTeamDeletionStoriesParams{
		TeamID: teamID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("lock team stories: %w", err)
	}
	return ids, nil
}

// DeleteTeamStories removes already-locked stories and writes the same durable
// integration events as interactive deletion. The caller captures attachments
// first and retires orphaned files after all team-owned relations are removed.
func (r *repo) DeleteTeamStories(ctx context.Context, tx pgx.Tx, actor platformauth.Actor, ids []uuid.UUID, deletedAt time.Time) error {
	if tx == nil || actor.Validate() != nil || actor.WorkspaceID == uuid.Nil || deletedAt.IsZero() {
		return errors.New("team story deletion transaction and actor are required")
	}
	queries := storyreadsql.New(tx)
	var credentialID *uuid.UUID
	if actor.CredentialID != uuid.Nil {
		credentialID = &actor.CredentialID
	}
	for start := 0; start < len(ids); start += storydomain.MaximumSecondaryMutationTargets {
		end := min(start+storydomain.MaximumSecondaryMutationTargets, len(ids))
		batch := ids[start:end]
		deleted, err := queries.DeleteTeamStoriesWithEvents(ctx, storyreadsql.DeleteTeamStoriesWithEventsParams{
			StoryIds: batch, WorkspaceID: actor.WorkspaceID, ActorKind: string(actor.Kind),
			ActorID: actor.PrincipalID, ActorCredentialID: credentialID, DeletedAt: deletedAt.UTC(),
		})
		if err != nil {
			return fmt.Errorf("delete team stories and record events: %w", err)
		}
		if deleted != int64(len(batch)) {
			return fmt.Errorf("%w: team stories changed during deletion", storydomain.ErrMutationConflict)
		}
	}
	return nil
}

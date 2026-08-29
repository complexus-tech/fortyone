package storiesrepository

import (
	"context"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) AuthorizedMayaScheduleBlocksExist(
	ctx context.Context,
	scope storydomain.MutationScope,
	storyID uuid.UUID,
) (bool, error) {
	if err := scope.Validate(); err != nil {
		return false, err
	}
	var exists bool
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		targets, err := authorizeSecondaryTargets(ctx, queries, scope, []uuid.UUID{storyID}, time.Now().UTC())
		if err != nil {
			return err
		}
		if targets[0].DeletedAt != nil {
			return storydomain.ErrNotFound
		}
		exists, err = queries.MayaScheduleBlocksExist(ctx, storyreadsql.MayaScheduleBlocksExistParams{
			StoryID: &storyID, WorkspaceID: scope.WorkspaceID,
		})
		if err != nil {
			return fmt.Errorf("check authorized Maya schedule blocks: %w", err)
		}
		return nil
	})
	if err != nil {
		return false, mapMutationDatabaseError(err)
	}
	return exists, nil
}

func (r *repo) UpdateAuthorizedAutoSchedulingStateIfUnchanged(
	ctx context.Context,
	scope storydomain.MutationScope,
	storyID uuid.UUID,
	expectedUpdatedAt time.Time,
	status string,
	reason *string,
	stateUpdatedAt time.Time,
	locked *bool,
) (bool, error) {
	if err := scope.Validate(); err != nil {
		return false, err
	}
	if storyID == uuid.Nil || expectedUpdatedAt.IsZero() || stateUpdatedAt.IsZero() || status == "" {
		return false, storydomain.ErrInvalidMutation
	}
	var updated bool
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		targets, err := authorizeSecondaryTargets(ctx, queries, scope, []uuid.UUID{storyID}, stateUpdatedAt.UTC())
		if err != nil {
			return err
		}
		if targets[0].DeletedAt != nil {
			return storydomain.ErrNotFound
		}
		lockedValue := false
		if locked != nil {
			lockedValue = *locked
		}
		rows, err := queries.UpdateAutoSchedulingState(ctx, storyreadsql.UpdateAutoSchedulingStateParams{
			AutoSchedulingStatus:    status,
			AutoSchedulingReason:    reason,
			AutoSchedulingUpdatedAt: &stateUpdatedAt,
			SetAutoSchedulingLocked: locked != nil,
			AutoSchedulingLocked:    lockedValue,
			StoryID:                 storyID,
			WorkspaceID:             scope.WorkspaceID,
			ExpectedUpdatedAt:       expectedUpdatedAt.UTC(),
		})
		if err != nil {
			return fmt.Errorf("update authorized auto-scheduling state: %w", err)
		}
		updated = rows == 1
		return nil
	})
	if err != nil {
		return false, mapMutationDatabaseError(err)
	}
	return updated, nil
}

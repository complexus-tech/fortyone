package storiesrepository

import (
	"context"
	"errors"
	"fmt"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) DuplicateStoryMutation(
	ctx context.Context,
	command storydomain.DuplicateStoryCommand,
) (storydomain.DuplicateStoryResult, error) {
	if err := r.mutationConfigured(); err != nil {
		return storydomain.DuplicateStoryResult{}, err
	}
	if err := command.Validate(); err != nil {
		return storydomain.DuplicateStoryResult{}, err
	}

	var (
		created      storydomain.Story
		sourceTeamID uuid.UUID
		err          error
	)
	for attempt := 1; attempt <= storyMutationTransactionAttempts; attempt++ {
		err = r.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
			queries := storyreadsql.New(tx)
			targets, authorizeErr := authorizeSecondaryTargets(
				ctx, queries, command.Scope, []uuid.UUID{command.SourceStoryID}, command.OccurredAt,
			)
			if authorizeErr != nil {
				return authorizeErr
			}
			source := targets[0]
			if source.DeletedAt != nil {
				return storydomain.ErrNotFound
			}
			if !source.UpdatedAt.Equal(command.ExpectedSourceUpdatedAt) {
				return storydomain.ErrMutationConflict
			}
			sourceTeamID = source.TeamID

			sequence, sequenceErr := queries.NextStorySequence(ctx, storyreadsql.NextStorySequenceParams{
				WorkspaceID: command.Scope.WorkspaceID,
				TeamID:      source.TeamID,
			})
			if sequenceErr != nil {
				return fmt.Errorf("advance duplicate story sequence: %w", sequenceErr)
			}
			reporterID := command.ReporterID
			insertedID, duplicateErr := queries.DuplicateAuthorizedStory(
				ctx,
				storyreadsql.DuplicateAuthorizedStoryParams{
					TargetStoryID:   command.TargetStoryID,
					SequenceID:      &sequence,
					SourceMediaPath: storyMediaPath(command.SourceStoryID),
					TargetMediaPath: storyMediaPath(command.TargetStoryID),
					ReporterID:      &reporterID,
					CreatedAt:       command.OccurredAt.UTC(),
					SourceStoryID:   command.SourceStoryID,
					WorkspaceID:     command.Scope.WorkspaceID,
				},
			)
			if errors.Is(duplicateErr, pgx.ErrNoRows) {
				return storydomain.ErrNotFound
			}
			if duplicateErr != nil {
				return fmt.Errorf("duplicate authorized story: %w", duplicateErr)
			}
			if insertedID != command.TargetStoryID {
				return fmt.Errorf("%w: duplicate story id mismatch", storydomain.ErrMutationConflict)
			}
			if err := queries.CopyDuplicatedStoryMediaLinks(ctx, storyreadsql.CopyDuplicatedStoryMediaLinksParams{
				TargetStoryID: command.TargetStoryID,
				CreatedBy:     command.ReporterID,
				SourceStoryID: command.SourceStoryID,
				WorkspaceID:   command.Scope.WorkspaceID,
			}); err != nil {
				return fmt.Errorf("copy duplicate story media links: %w", err)
			}
			if err := upsertMutationActivity(ctx, queries, command.Activity, false); err != nil {
				return err
			}
			if err := insertMutationEvent(ctx, queries, command.Event); err != nil {
				return err
			}

			params := mutationQueryScope(command.Scope, command.OccurredAt)
			params.StoryID = command.TargetStoryID
			row, snapshotErr := queries.GetStoryMutationSnapshot(ctx, params)
			if errors.Is(snapshotErr, pgx.ErrNoRows) {
				return storydomain.ErrMutationForbidden
			}
			if snapshotErr != nil {
				return fmt.Errorf("load duplicated story snapshot: %w", snapshotErr)
			}
			created = mutationSnapshotToStory(row)
			created.CreatedNow = true
			return nil
		})
		if err == nil {
			return storydomain.DuplicateStoryResult{Story: created}, nil
		}
		if !isStorySequenceUniqueConflict(err) && !platformdatabase.IsRetryableTransactionError(err) {
			return storydomain.DuplicateStoryResult{}, mapMutationDatabaseError(err)
		}
		if isStorySequenceUniqueConflict(err) && sourceTeamID != uuid.Nil {
			if syncErr := r.reads.SynchronizeStorySequence(ctx, storyreadsql.SynchronizeStorySequenceParams{
				WorkspaceID: command.Scope.WorkspaceID,
				TeamID:      sourceTeamID,
			}); syncErr != nil {
				return storydomain.DuplicateStoryResult{}, fmt.Errorf("synchronize duplicate story sequence: %w", syncErr)
			}
		}
	}
	return storydomain.DuplicateStoryResult{}, fmt.Errorf(
		"duplicate story mutation after %d attempts: %w", storyMutationTransactionAttempts, err,
	)
}

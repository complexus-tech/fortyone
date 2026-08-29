package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) PrepareStoryAssociationMutation(
	ctx context.Context,
	scope storydomain.MutationScope,
	associationID uuid.UUID,
) (storydomain.AssociationSnapshot, error) {
	if err := scope.Validate(); err != nil {
		return storydomain.AssociationSnapshot{}, err
	}
	if associationID == uuid.Nil {
		return storydomain.AssociationSnapshot{}, storydomain.ErrInvalidMutation
	}
	var result storydomain.AssociationSnapshot
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		row, err := queries.LockStoryAssociation(ctx, storyreadsql.LockStoryAssociationParams{
			AssociationID: associationID, WorkspaceID: scope.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return storydomain.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("lock story association for preparation: %w", err)
		}
		targets, err := authorizeSecondaryTargets(
			ctx, queries, scope, []uuid.UUID{row.FromStoryID, row.ToStoryID}, nowUTC(),
		)
		if err != nil {
			return err
		}
		if targets[0].DeletedAt != nil || targets[1].DeletedAt != nil {
			return storydomain.ErrNotFound
		}
		titles := secondaryTargetTitles(targets)
		result = storydomain.AssociationSnapshot{
			ID: row.ID, FromStoryID: row.FromStoryID, ToStoryID: row.ToStoryID,
			Type: row.AssociationType, FromStoryTitle: titles[row.FromStoryID], ToStoryTitle: titles[row.ToStoryID],
		}
		return nil
	})
	if err != nil {
		return storydomain.AssociationSnapshot{}, mapMutationDatabaseError(err)
	}
	return result, nil
}

func (r *repo) ApplyStoryAssociationMutation(
	ctx context.Context,
	command storydomain.AssociationMutationCommand,
) (storydomain.AssociationSnapshot, error) {
	if err := command.Validate(); err != nil {
		return storydomain.AssociationSnapshot{}, err
	}
	result := command.Association
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		var current *storyreadsql.LockStoryAssociationRow
		if command.Action != storydomain.AssociationMutationAdd {
			row, err := queries.LockStoryAssociation(ctx, storyreadsql.LockStoryAssociationParams{
				AssociationID: command.Association.ID, WorkspaceID: command.Scope.WorkspaceID,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				return storydomain.ErrNotFound
			}
			if err != nil {
				return fmt.Errorf("lock story association: %w", err)
			}
			current = &row
			if !associationMatchesExpected(row, *command.Expected) {
				return storydomain.ErrMutationConflict
			}
		}

		targetIDs := associationTargetIDs(command, current)
		targets, err := authorizeSecondaryTargets(ctx, queries, command.Scope, targetIDs, command.OccurredAt)
		if err != nil {
			return err
		}
		for _, target := range targets {
			if target.DeletedAt != nil {
				return storydomain.ErrNotFound
			}
		}

		switch command.Action {
		case storydomain.AssociationMutationAdd:
			row, err := queries.InsertStoryAssociation(ctx, storyreadsql.InsertStoryAssociationParams{
				AssociationID: command.Association.ID, FromStoryID: command.Association.FromStoryID,
				ToStoryID: command.Association.ToStoryID, AssociationType: command.Association.Type,
				WorkspaceID: command.Scope.WorkspaceID,
			})
			if err != nil {
				return fmt.Errorf("insert story association: %w", err)
			}
			result.ID, result.FromStoryID, result.ToStoryID, result.Type = row.ID, row.FromStoryID, row.ToStoryID, row.AssociationType
		case storydomain.AssociationMutationUpdate:
			row, err := queries.UpdateStoryAssociation(ctx, storyreadsql.UpdateStoryAssociationParams{
				FromStoryID: command.Association.FromStoryID, ToStoryID: command.Association.ToStoryID,
				AssociationType: command.Association.Type, AssociationID: command.Association.ID,
				WorkspaceID: command.Scope.WorkspaceID,
			})
			if err != nil {
				return fmt.Errorf("update story association: %w", err)
			}
			previousType := current.AssociationType
			result.ID, result.FromStoryID, result.ToStoryID, result.Type = row.ID, row.FromStoryID, row.ToStoryID, row.AssociationType
			result.PreviousType = &previousType
		case storydomain.AssociationMutationRemove:
			row, err := queries.DeleteStoryAssociation(ctx, storyreadsql.DeleteStoryAssociationParams{
				AssociationID: command.Association.ID, WorkspaceID: command.Scope.WorkspaceID,
			})
			if err != nil {
				return fmt.Errorf("delete story association: %w", err)
			}
			result.ID, result.FromStoryID, result.ToStoryID, result.Type = row.ID, row.FromStoryID, row.ToStoryID, row.AssociationType
		default:
			return storydomain.ErrInvalidMutation
		}

		titles := secondaryTargetTitles(targets)
		result.FromStoryTitle = titles[result.FromStoryID]
		result.ToStoryTitle = titles[result.ToStoryID]
		for _, activity := range command.Activities {
			if err := upsertMutationActivity(ctx, queries, activity, false); err != nil {
				return err
			}
		}
		for _, event := range command.Events {
			if err := insertMutationEvent(ctx, queries, event); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return storydomain.AssociationSnapshot{}, mapMutationDatabaseError(err)
	}
	return result, nil
}

func associationMatchesExpected(row storyreadsql.LockStoryAssociationRow, expected storydomain.AssociationSnapshot) bool {
	return row.ID == expected.ID && row.FromStoryID == expected.FromStoryID &&
		row.ToStoryID == expected.ToStoryID && row.AssociationType == expected.Type
}

func associationTargetIDs(
	command storydomain.AssociationMutationCommand,
	current *storyreadsql.LockStoryAssociationRow,
) []uuid.UUID {
	values := []uuid.UUID{command.Association.FromStoryID, command.Association.ToStoryID}
	if current != nil {
		values = append(values, current.FromStoryID, current.ToStoryID)
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sortUUIDs(result)
	return result
}

func secondaryTargetTitles(targets []authorizedSecondaryTarget) map[uuid.UUID]string {
	result := make(map[uuid.UUID]string, len(targets))
	for _, target := range targets {
		result[target.ID] = target.Title
	}
	return result
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

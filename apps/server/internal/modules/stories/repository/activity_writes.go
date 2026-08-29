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

func (r *repo) RecordStoryActivities(
	ctx context.Context,
	command storydomain.RecordActivitiesCommand,
) ([]storydomain.MutationActivity, error) {
	if err := command.Validate(); err != nil {
		return nil, err
	}
	storyIDs := make([]uuid.UUID, 0, len(command.Activities))
	seen := make(map[uuid.UUID]struct{}, len(command.Activities))
	for _, write := range command.Activities {
		if _, exists := seen[write.Activity.StoryID]; exists {
			continue
		}
		seen[write.Activity.StoryID] = struct{}{}
		storyIDs = append(storyIDs, write.Activity.StoryID)
	}
	result := make([]storydomain.MutationActivity, 0, len(command.Activities))
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := storyreadsql.New(tx)
		targets, err := authorizeSecondaryTargets(ctx, queries, command.Scope, storyIDs, time.Now().UTC())
		if err != nil {
			return err
		}
		for _, target := range targets {
			if target.DeletedAt != nil {
				return storydomain.ErrNotFound
			}
		}
		for _, write := range command.Activities {
			row, err := queries.UpsertStoryMutationActivity(ctx, activityWriteParams(write))
			if err != nil {
				return fmt.Errorf("persist authorized story activity: %w", err)
			}
			result = append(result, activityRowToDomain(row))
		}
		return nil
	})
	if err != nil {
		return nil, mapMutationDatabaseError(err)
	}
	return result, nil
}

func activityWriteParams(write storydomain.ActivityWrite) storyreadsql.UpsertStoryMutationActivityParams {
	activity := write.Activity
	workspaceID := activity.WorkspaceID
	return storyreadsql.UpsertStoryMutationActivityParams{
		CurrentValue: activity.CurrentValue,
		NewValue:     activity.NewValue,
		Reason:       activity.Reason,
		CreatedAt:    activity.CreatedAt.UTC(),
		StoryID:      activity.StoryID,
		UserID:       activity.UserID,
		WorkspaceID:  &workspaceID,
		ActivityType: activity.Type,
		FieldChanged: activity.Field,
		Compact:      write.Compact,
		ActivityID:   activity.ID,
		OldValue:     activity.OldValue,
	}
}

func activityRowToDomain(row storyreadsql.UpsertStoryMutationActivityRow) storydomain.MutationActivity {
	workspaceID := uuid.Nil
	if row.WorkspaceID != nil {
		workspaceID = *row.WorkspaceID
	}
	return storydomain.MutationActivity{
		ID: row.ActivityID, StoryID: row.StoryID, UserID: row.UserID,
		Type: row.ActivityType, Field: row.FieldChanged, CurrentValue: row.CurrentValue,
		OldValue: row.OldValue, NewValue: row.NewValue, Reason: row.Reason,
		WorkspaceID: workspaceID, CreatedAt: row.CreatedAt,
	}
}

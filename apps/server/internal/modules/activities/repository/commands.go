package activitiesrepository

import (
	"context"
	"fmt"

	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
)

// Create inserts a new activity into the database.
func (r *repo) Create(ctx context.Context, na activities.CoreNewActivity) error {
	const query = `
		INSERT INTO story_activities (
			activity_id, story_id, user_id, activity_type, field_changed, current_value, created_at, workspace_id
		) VALUES (
			:activity_id, :story_id, :user_id, :activity_type, :field_changed, :current_value, :created_at, :workspace_id
		)`

	activity := dbActivity{
		StoryID:      na.StoryID,
		UserID:       na.UserID,
		Type:         na.Type,
		Field:        na.Field,
		CurrentValue: na.CurrentValue,
		WorkspaceID:  na.WorkspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, activity); err != nil {
		r.log.Error(ctx, "failed to create activity", "error", err)
		return fmt.Errorf("creating activity: %w", err)
	}

	return nil
}

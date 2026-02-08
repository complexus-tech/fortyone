package okractivitiesrepository

import (
	"context"
	"fmt"

	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
)

// Create inserts a new OKR activity into the database.
func (r *repo) Create(ctx context.Context, na okractivities.CoreNewActivity) error {
	const query = `
        INSERT INTO okr_activities (
            objective_id, key_result_id, user_id, activity_type, update_type,
            field_changed, current_value, comment, workspace_id
        ) VALUES (
            :objective_id, :key_result_id, :user_id, :activity_type, :update_type,
            :field_changed, :current_value, :comment, :workspace_id
        )`

	activity := dbActivity{
		ObjectiveID:  na.ObjectiveID,
		KeyResultID:  na.KeyResultID,
		UserID:       na.UserID,
		Type:         string(na.Type),
		UpdateType:   string(na.UpdateType),
		Field:        na.Field,
		CurrentValue: na.CurrentValue,
		Comment:      na.Comment,
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

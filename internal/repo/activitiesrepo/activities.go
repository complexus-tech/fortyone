package activitiesrepo

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/activities"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

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

// GetActivities gets activities for a user.
func (r *repo) GetActivities(ctx context.Context, userID uuid.UUID, limit int, workspaceId uuid.UUID) ([]activities.CoreActivity, error) {
	const query = `
SELECT
		sa.activity_id, sa.story_id, sa.user_id, sa.activity_type, sa.field_changed, sa.current_value, sa.created_at, sa.workspace_id,
		u.username, u.full_name, u.avatar_url, u.is_active
		FROM story_activities sa
		JOIN users u ON sa.user_id = u.user_id
		WHERE sa.user_id = :user_id
		AND sa.workspace_id = :workspace_id
		AND u.is_active = true
		ORDER BY sa.created_at DESC
		LIMIT :limit`

	params := map[string]interface{}{
		"user_id":      userID,
		"workspace_id": workspaceId,
		"limit":        limit,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return nil, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	var dbActs []dbActivity
	if err := stmt.SelectContext(ctx, &dbActs, params); err != nil {
		r.log.Error(ctx, "failed to get activities", "error", err)
		return nil, fmt.Errorf("selecting activities: %w", err)
	}

	return toCoreActivities(dbActs), nil
}

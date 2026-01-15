package okractivitiesrepo

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/okractivities"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository provides access to the OKR activities storage.
type Repository interface {
	Create(ctx context.Context, na okractivities.CoreNewActivity) error
	GetObjectiveActivities(ctx context.Context, objectiveID uuid.UUID, page, pageSize int) ([]okractivities.CoreActivity, bool, error)
	GetKeyResultActivities(ctx context.Context, keyResultID uuid.UUID, page, pageSize int) ([]okractivities.CoreActivity, bool, error)
}

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

// GetObjectiveActivities returns activities for a specific objective with pagination.
func (r *repo) GetObjectiveActivities(ctx context.Context, objectiveID uuid.UUID, page, pageSize int) ([]okractivities.CoreActivity, bool, error) {
	offset := (page - 1) * pageSize
	limit := pageSize + 1

	const query = `
        SELECT 
            oa.activity_id, oa.objective_id, oa.key_result_id, oa.user_id, oa.activity_type, oa.update_type,
            oa.field_changed, oa.current_value, oa.comment, oa.created_at, oa.workspace_id,
            u.username, u.full_name, u.avatar_url, u.is_active
        FROM okr_activities oa
        JOIN users u ON oa.user_id = u.user_id
        WHERE oa.objective_id = :objective_id
            AND u.is_active = true
        ORDER BY oa.created_at DESC
        LIMIT :limit OFFSET :offset`

	params := map[string]any{
		"objective_id": objectiveID,
		"limit":        limit,
		"offset":       offset,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return nil, false, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	var dbActs []dbActivity
	if err := stmt.SelectContext(ctx, &dbActs, params); err != nil {
		r.log.Error(ctx, "failed to get objective activities", "error", err)
		return nil, false, fmt.Errorf("selecting activities: %w", err)
	}

	hasMore := len(dbActs) > pageSize
	if hasMore {
		dbActs = dbActs[:pageSize]
	}

	return toCoreActivities(dbActs), hasMore, nil
}

// GetKeyResultActivities returns activities for a specific key result with pagination.
func (r *repo) GetKeyResultActivities(ctx context.Context, keyResultID uuid.UUID, page, pageSize int) ([]okractivities.CoreActivity, bool, error) {
	offset := (page - 1) * pageSize
	limit := pageSize + 1

	const query = `
        SELECT 
            oa.activity_id, oa.objective_id, oa.key_result_id, oa.user_id, oa.activity_type, oa.update_type,
            oa.field_changed, oa.current_value, oa.comment, oa.created_at, oa.workspace_id,
            u.username, u.full_name, u.avatar_url, u.is_active
        FROM okr_activities oa
        JOIN users u ON oa.user_id = u.user_id
        WHERE oa.key_result_id = :key_result_id
            AND u.is_active = true
        ORDER BY oa.created_at DESC
        LIMIT :limit OFFSET :offset`

	params := map[string]any{
		"key_result_id": keyResultID,
		"limit":         limit,
		"offset":        offset,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare named statement", "error", err)
		return nil, false, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	var dbActs []dbActivity
	if err := stmt.SelectContext(ctx, &dbActs, params); err != nil {
		r.log.Error(ctx, "failed to get key result activities", "error", err)
		return nil, false, fmt.Errorf("selecting activities: %w", err)
	}

	hasMore := len(dbActs) > pageSize
	if hasMore {
		dbActs = dbActs[:pageSize]
	}

	return toCoreActivities(dbActs), hasMore, nil
}

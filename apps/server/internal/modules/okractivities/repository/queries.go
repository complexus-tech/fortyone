package okractivitiesrepository

import (
	"context"
	"fmt"

	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	"github.com/google/uuid"
)

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

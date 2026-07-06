package mayarepository

import (
	"context"
	"encoding/json"
	"fmt"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *Repo {
	return &Repo{db: db, log: log}
}

func (r *Repo) CreateRun(ctx context.Context, input maya.CreateRunInput) (maya.CoreRun, error) {
	const query = `
		INSERT INTO maya_agent_runs (
			workspace_id,
			story_id,
			triggered_by_user_id,
			trigger_type,
			status,
			context
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			run_id,
			workspace_id,
			story_id,
			triggered_by_user_id,
			trigger_type,
			status,
			summary,
			context,
			error_message,
			started_at,
			completed_at,
			created_at,
			updated_at
	`
	contextPayload := input.Context
	if len(contextPayload) == 0 {
		contextPayload = json.RawMessage(`{}`)
	}
	var row dbRun
	if err := r.db.GetContext(ctx, &row, query, input.WorkspaceID, input.StoryID, input.TriggeredBy, string(input.Trigger), string(maya.RunStatusRunning), contextPayload); err != nil {
		return maya.CoreRun{}, fmt.Errorf("create maya run: %w", err)
	}
	return toCoreRun(row), nil
}

func (r *Repo) CompleteRun(ctx context.Context, runID uuid.UUID, status maya.RunStatus, summary string, message *string) (maya.CoreRun, error) {
	const query = `
		UPDATE maya_agent_runs
		SET status = $2,
			summary = $3,
			error_message = $4,
			completed_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE run_id = $1
		RETURNING
			run_id,
			workspace_id,
			story_id,
			triggered_by_user_id,
			trigger_type,
			status,
			summary,
			context,
			error_message,
			started_at,
			completed_at,
			created_at,
			updated_at
	`
	var row dbRun
	if err := r.db.GetContext(ctx, &row, query, runID, string(status), summary, message); err != nil {
		return maya.CoreRun{}, fmt.Errorf("complete maya run: %w", err)
	}
	return toCoreRun(row), nil
}

func (r *Repo) CreateActions(ctx context.Context, actions []maya.CoreAction) ([]maya.CoreAction, error) {
	if len(actions) == 0 {
		return []maya.CoreAction{}, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin maya actions transaction: %w", err)
	}
	defer tx.Rollback()

	const query = `
		INSERT INTO maya_agent_actions (
			run_id,
			workspace_id,
			story_id,
			action_type,
			status,
			reason,
			payload
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			action_id,
			run_id,
			workspace_id,
			story_id,
			action_type,
			status,
			reason,
			payload,
			error_message,
			applied_at,
			created_at,
			updated_at
	`

	created := make([]maya.CoreAction, 0, len(actions))
	for _, action := range actions {
		payload, err := json.Marshal(action.Payload)
		if err != nil {
			return nil, fmt.Errorf("marshal maya action payload: %w", err)
		}
		var row dbAction
		if err := tx.GetContext(ctx, &row, query, action.RunID, action.WorkspaceID, action.StoryID, string(action.Type), string(action.Status), action.Reason, payload); err != nil {
			return nil, fmt.Errorf("create maya action: %w", err)
		}
		core, err := toCoreAction(row)
		if err != nil {
			return nil, fmt.Errorf("decode maya action payload: %w", err)
		}
		created = append(created, core)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit maya actions transaction: %w", err)
	}
	return created, nil
}

func (r *Repo) MarkActionApplied(ctx context.Context, actionID uuid.UUID) error {
	const query = `
		UPDATE maya_agent_actions
		SET status = $2,
			applied_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP,
			error_message = NULL
		WHERE action_id = $1
	`
	if _, err := r.db.ExecContext(ctx, query, actionID, string(maya.ActionStatusApplied)); err != nil {
		return fmt.Errorf("mark maya action applied: %w", err)
	}
	return nil
}

func (r *Repo) MarkActionFailed(ctx context.Context, actionID uuid.UUID, message string) error {
	const query = `
		UPDATE maya_agent_actions
		SET status = $2,
			error_message = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE action_id = $1
	`
	if _, err := r.db.ExecContext(ctx, query, actionID, string(maya.ActionStatusFailed), message); err != nil {
		return fmt.Errorf("mark maya action failed: %w", err)
	}
	return nil
}

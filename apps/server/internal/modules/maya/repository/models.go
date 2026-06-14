package mayarepository

import (
	"encoding/json"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/google/uuid"
)

type dbRun struct {
	ID          uuid.UUID       `db:"run_id"`
	WorkspaceID uuid.UUID       `db:"workspace_id"`
	StoryID     uuid.UUID       `db:"story_id"`
	TriggeredBy uuid.UUID       `db:"triggered_by_user_id"`
	Trigger     string          `db:"trigger_type"`
	Status      string          `db:"status"`
	Summary     string          `db:"summary"`
	Context     json.RawMessage `db:"context"`
	Error       *string         `db:"error_message"`
	StartedAt   time.Time       `db:"started_at"`
	CompletedAt *time.Time      `db:"completed_at"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

type dbAction struct {
	ID          uuid.UUID       `db:"action_id"`
	RunID       uuid.UUID       `db:"run_id"`
	WorkspaceID uuid.UUID       `db:"workspace_id"`
	StoryID     uuid.UUID       `db:"story_id"`
	Type        string          `db:"action_type"`
	Status      string          `db:"status"`
	Reason      string          `db:"reason"`
	Payload     json.RawMessage `db:"payload"`
	Error       *string         `db:"error_message"`
	AppliedAt   *time.Time      `db:"applied_at"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

func toCoreRun(row dbRun) maya.CoreRun {
	return maya.CoreRun{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		StoryID:     row.StoryID,
		TriggeredBy: row.TriggeredBy,
		Trigger:     maya.RunTrigger(row.Trigger),
		Status:      maya.RunStatus(row.Status),
		Summary:     row.Summary,
		Context:     row.Context,
		Error:       row.Error,
		StartedAt:   row.StartedAt,
		CompletedAt: row.CompletedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toCoreAction(row dbAction) (maya.CoreAction, error) {
	var payload maya.ActionPayload
	if len(row.Payload) > 0 {
		if err := json.Unmarshal(row.Payload, &payload); err != nil {
			return maya.CoreAction{}, err
		}
	}
	return maya.CoreAction{
		ID:          row.ID,
		RunID:       row.RunID,
		WorkspaceID: row.WorkspaceID,
		StoryID:     row.StoryID,
		Type:        maya.ActionType(row.Type),
		Status:      maya.ActionStatus(row.Status),
		Reason:      row.Reason,
		Payload:     payload,
		PayloadJSON: row.Payload,
		Error:       row.Error,
		AppliedAt:   row.AppliedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

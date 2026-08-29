package mayarepository

import (
	"encoding/json"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
)

func toCoreRun(row mayasql.MayaAgentRun) mayadomain.CoreRun {
	return mayadomain.CoreRun{
		ID:          row.RunID,
		WorkspaceID: row.WorkspaceID,
		StoryID:     row.StoryID,
		TriggeredBy: row.TriggeredByUserID,
		Trigger:     mayadomain.RunTrigger(row.TriggerType),
		Status:      mayadomain.RunStatus(row.Status),
		Summary:     row.Summary,
		Context:     append(json.RawMessage(nil), row.Context...),
		Error:       row.ErrorMessage,
		StartedAt:   row.StartedAt,
		CompletedAt: row.CompletedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toCoreAction(row mayasql.MayaAgentAction) (mayadomain.CoreAction, error) {
	var payload mayadomain.ActionPayload
	if len(row.Payload) > 0 {
		if err := json.Unmarshal(row.Payload, &payload); err != nil {
			return mayadomain.CoreAction{}, err
		}
	}
	return mayadomain.CoreAction{
		ID:          row.ActionID,
		RunID:       row.RunID,
		WorkspaceID: row.WorkspaceID,
		StoryID:     row.StoryID,
		Type:        mayadomain.ActionType(row.ActionType),
		Status:      mayadomain.ActionStatus(row.Status),
		Reason:      row.Reason,
		Payload:     payload,
		PayloadJSON: append(json.RawMessage(nil), row.Payload...),
		Error:       row.ErrorMessage,
		AppliedAt:   row.AppliedAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

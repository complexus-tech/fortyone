package reportsrepository

import (
	"context"
	"encoding/json"
	"fmt"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
)

func (r *repo) CreateWorkspaceAnalyticsEvent(ctx context.Context, input reports.CoreWorkspaceAnalyticsEventInput) error {
	properties, err := json.Marshal(input.Properties)
	if err != nil {
		return fmt.Errorf("marshaling analytics event properties: %w", err)
	}

	const query = `
		INSERT INTO workspace_analytics_events (
			workspace_id,
			user_id,
			team_id,
			story_id,
			objective_id,
			sprint_id,
			key_result_id,
			event_name,
			surface,
			properties,
			occurred_at
		)
		VALUES (
			:workspace_id,
			:user_id,
			:team_id,
			:story_id,
			:objective_id,
			:sprint_id,
			:key_result_id,
			:event_name,
			:surface,
			:properties,
			:occurred_at
		)
	`

	params := map[string]any{
		"workspace_id":  input.WorkspaceID,
		"user_id":       input.UserID,
		"team_id":       input.TeamID,
		"story_id":      input.StoryID,
		"objective_id":  input.ObjectiveID,
		"sprint_id":     input.SprintID,
		"key_result_id": input.KeyResultID,
		"event_name":    input.EventName,
		"surface":       input.Surface,
		"properties":    properties,
		"occurred_at":   input.OccurredAt,
	}

	if _, err := r.db.NamedExecContext(ctx, query, params); err != nil {
		return fmt.Errorf("inserting workspace analytics event: %w", err)
	}

	return nil
}

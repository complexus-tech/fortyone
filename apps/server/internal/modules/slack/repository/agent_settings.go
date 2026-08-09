package slackrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxSlackAgentGuidanceRunes = 4000

type AgentSettingsRecord struct {
	WorkspaceID            uuid.UUID `db:"workspace_id"`
	AssistantEnabled       bool      `db:"assistant_enabled"`
	WorkflowActionsEnabled bool      `db:"workflow_actions_enabled"`
	Guidance               string    `db:"guidance"`
	CreatedAt              time.Time `db:"created_at"`
	UpdatedAt              time.Time `db:"updated_at"`
}

type AgentSettingsInput struct {
	AssistantEnabled       bool
	WorkflowActionsEnabled bool
	Guidance               string
}

func (r *Repo) GetAgentSettings(ctx context.Context, workspaceID uuid.UUID) (AgentSettingsRecord, error) {
	if workspaceID == uuid.Nil {
		return AgentSettingsRecord{}, errors.New("workspace is required")
	}
	var record AgentSettingsRecord
	if err := r.db.GetContext(ctx, &record, `
		SELECT workspace.workspace_id,
		       COALESCE(settings.assistant_enabled, true) AS assistant_enabled,
		       COALESCE(settings.workflow_actions_enabled, true) AS workflow_actions_enabled,
		       COALESCE(settings.guidance, '') AS guidance,
		       COALESCE(settings.created_at, workspace.created_at) AS created_at,
		       COALESCE(settings.updated_at, workspace.updated_at) AS updated_at
		FROM workspaces workspace
		LEFT JOIN slack_agent_settings settings
		  ON settings.workspace_id = workspace.workspace_id
		WHERE workspace.workspace_id = $1
	`, workspaceID); err != nil {
		return AgentSettingsRecord{}, fmt.Errorf("get Slack agent settings: %w", err)
	}
	return record, nil
}

func (r *Repo) UpsertAgentSettings(
	ctx context.Context,
	workspaceID uuid.UUID,
	input AgentSettingsInput,
) (AgentSettingsRecord, error) {
	guidance := strings.TrimSpace(input.Guidance)
	if workspaceID == uuid.Nil {
		return AgentSettingsRecord{}, errors.New("workspace is required")
	}
	if len([]rune(guidance)) > MaxSlackAgentGuidanceRunes {
		return AgentSettingsRecord{}, fmt.Errorf("Slack agent guidance must be %d characters or fewer", MaxSlackAgentGuidanceRunes)
	}
	var record AgentSettingsRecord
	if err := r.db.GetContext(ctx, &record, `
		INSERT INTO slack_agent_settings (
			workspace_id,
			assistant_enabled,
			workflow_actions_enabled,
			guidance
		)
		SELECT workspace_id, $2, $3, $4
		FROM workspaces
		WHERE workspace_id = $1
		ON CONFLICT (workspace_id) DO UPDATE SET
			assistant_enabled = EXCLUDED.assistant_enabled,
			workflow_actions_enabled = EXCLUDED.workflow_actions_enabled,
			guidance = EXCLUDED.guidance,
			updated_at = NOW()
		RETURNING workspace_id, assistant_enabled, workflow_actions_enabled,
		          guidance, created_at, updated_at
	`, workspaceID, input.AssistantEnabled, input.WorkflowActionsEnabled, guidance); err != nil {
		return AgentSettingsRecord{}, fmt.Errorf("upsert Slack agent settings: %w", err)
	}
	return record, nil
}

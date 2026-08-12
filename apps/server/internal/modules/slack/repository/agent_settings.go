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
	WorkspaceID uuid.UUID `db:"workspace_id"`
	Guidance    string    `db:"guidance"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type AgentSettingsInput struct {
	Guidance string
}

func (r *Repo) GetAgentSettings(ctx context.Context, workspaceID uuid.UUID) (AgentSettingsRecord, error) {
	if workspaceID == uuid.Nil {
		return AgentSettingsRecord{}, errors.New("workspace is required")
	}
	var record AgentSettingsRecord
	if err := r.db.GetContext(ctx, &record, `
		SELECT workspace.workspace_id,
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
			INSERT INTO slack_agent_settings (workspace_id, guidance)
			SELECT workspace_id, $2
		FROM workspaces
		WHERE workspace_id = $1
		ON CONFLICT (workspace_id) DO UPDATE SET
			guidance = EXCLUDED.guidance,
			updated_at = NOW()
		RETURNING workspace_id, guidance, created_at, updated_at
		`, workspaceID, guidance); err != nil {
		return AgentSettingsRecord{}, fmt.Errorf("upsert Slack agent settings: %w", err)
	}
	return record, nil
}

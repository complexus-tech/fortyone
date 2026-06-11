package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type auditEvent struct {
	WorkspaceID uuid.UUID  `db:"workspace_id"`
	TeamID      *uuid.UUID `db:"team_id"`
	ActorType   string     `db:"actor_type"`
	ActorID     *uuid.UUID `db:"actor_id"`
	EntityType  string     `db:"entity_type"`
	EntityID    *uuid.UUID `db:"entity_id"`
	EventType   string     `db:"event_type"`
	Metadata    string     `db:"metadata"`
}

func recordAuditEvent(ctx context.Context, exec sqlx.ExtContext, event auditEvent) error {
	query := `
		INSERT INTO audit_events (
			workspace_id,
			team_id,
			actor_type,
			actor_id,
			entity_type,
			entity_id,
			event_type,
			metadata
		) VALUES (
			:workspace_id,
			:team_id,
			:actor_type,
			:actor_id,
			:entity_type,
			:entity_id,
			:event_type,
			CAST(:metadata AS jsonb)
		)
	`

	_, err := sqlx.NamedExecContext(ctx, exec, query, event)
	if err != nil {
		return fmt.Errorf("failed to record audit event %s: %w", event.EventType, err)
	}

	return nil
}

func auditMetadata(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal audit metadata: %w", err)
	}

	return string(data), nil
}

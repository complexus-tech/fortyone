package developercredentialsrepository

import (
	"context"
	"encoding/json"
	"fmt"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentialssql "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository/sqlc"
	"github.com/google/uuid"
)

type auditMetadata struct {
	ScopeCount           int        `json:"scope_count,omitempty"`
	TeamRestrictionCount int        `json:"team_restriction_count,omitempty"`
	WorkspaceRole        string     `json:"workspace_role,omitempty"`
	RotatedFromID        *uuid.UUID `json:"rotated_from_id,omitempty"`
	RotationOverlapSecs  int64      `json:"rotation_overlap_seconds,omitempty"`
}

func insertAuditEvent(
	ctx context.Context,
	queries developercredentialssql.Querier,
	event developercredentialsdomain.AuditEvent,
) error {
	metadata, err := json.Marshal(auditMetadata{
		ScopeCount:           event.ScopeCount,
		TeamRestrictionCount: event.TeamCount,
		WorkspaceRole:        string(event.WorkspaceRole),
		RotatedFromID:        event.RotatedFromID,
		RotationOverlapSecs:  int64(event.RotationOverlap.Seconds()),
	})
	if err != nil {
		return fmt.Errorf("marshal developer credential audit metadata: %w", err)
	}
	var actorCredentialID *uuid.UUID
	if event.Actor.CredentialID != uuid.Nil {
		actorCredentialID = uuidPointer(event.Actor.CredentialID)
	}
	var reasonCode *string
	if event.ReasonCode != "" {
		reasonCode = stringPointer(event.ReasonCode)
	}
	var requestID *string
	if event.RequestID != "" {
		requestID = stringPointer(event.RequestID)
	}
	if err := queries.InsertDeveloperCredentialAuditEvent(ctx, developercredentialssql.InsertDeveloperCredentialAuditEventParams{
		EventID:           event.ID,
		WorkspaceID:       event.WorkspaceID,
		ActorKind:         string(event.Actor.Kind),
		ActorID:           event.Actor.PrincipalID,
		ActorCredentialID: actorCredentialID,
		Operation:         event.Operation,
		SubjectType:       event.SubjectType,
		SubjectID:         event.SubjectID,
		Result:            event.Result,
		ReasonCode:        reasonCode,
		RequestID:         requestID,
		Metadata:          metadata,
		CreatedAt:         event.CreatedAt,
	}); err != nil {
		return fmt.Errorf("insert developer credential audit event: %w", err)
	}
	return nil
}

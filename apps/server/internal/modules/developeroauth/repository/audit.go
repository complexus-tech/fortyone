package developeroauthrepository

import (
	"context"
	"fmt"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	"github.com/google/uuid"
)

func createAuditEvent(
	ctx context.Context,
	queries *developeroauthsql.Queries,
	event developeroauthdomain.AuditEvent,
) error {
	metadata := event.Metadata
	if len(metadata) == 0 {
		metadata = []byte(`{}`)
	}
	var actorKind *string
	if event.ActorKind != "" {
		value := string(event.ActorKind)
		actorKind = &value
	}
	var requestID *string
	if event.RequestID != "" {
		requestID = &event.RequestID
	}
	var subjectType *string
	if event.SubjectType != "" {
		subjectType = &event.SubjectType
	}
	if err := queries.CreateOAuthAuditEvent(ctx, developeroauthsql.CreateOAuthAuditEventParams{
		EventID: event.ID, ApplicationID: event.ApplicationID, GrantID: event.GrantID,
		UserID: event.UserID, WorkspaceID: event.WorkspaceID, InstallationID: event.InstallationID,
		PrincipalID: event.PrincipalID, SecretID: event.SecretID, ActorKind: actorKind,
		ActorID: event.ActorID, ActorCredentialID: event.ActorCredentialID, RequestID: requestID,
		SubjectType: subjectType, SubjectID: event.SubjectID, Operation: event.Operation,
		Result: event.Result, ReasonCode: event.ReasonCode, Metadata: metadata, CreatedAt: event.CreatedAt,
	}); err != nil {
		return fmt.Errorf("create OAuth audit event: %w", err)
	}
	return nil
}

func uuidForAudit() uuid.UUID {
	return uuid.New()
}

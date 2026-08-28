package webhooksrepository

import (
	"fmt"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksql "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository/sqlc"
	"github.com/google/uuid"
)

func toRecord(row webhooksql.MessagingInboundEvent) (webhooks.Record, error) {
	status, ok := webhooks.ParseStatus(row.Status)
	if !ok {
		return webhooks.Record{}, fmt.Errorf("map webhook delivery: %w", webhooks.ErrInvalidState)
	}
	return webhooks.Record{
		ID: row.ID,
		Envelope: webhooks.Envelope{
			Version:                row.EnvelopeVersion,
			Provider:               integrations.ProviderKey(row.Provider),
			DeliveryID:             row.ExternalEventID,
			EventType:              row.EventType,
			ExternalAccountID:      row.ExternalWorkspaceID,
			WorkspaceID:            valueOrNil(row.WorkspaceID),
			InstallationID:         valueOrNil(row.InstallationID),
			InstallationGeneration: valueOrNil(row.InstallationGeneration),
			TraceID:                stringOrEmpty(row.TraceID),
			ReceivedAt:             row.ReceivedAt,
		},
		Status:             status,
		AttemptCount:       row.AttemptCount,
		RecoveryGeneration: row.RecoveryGeneration,
		RecoveryEnqueuedAt: row.RecoveryEnqueuedAt,
		ProcessedAt:        row.ProcessedAt,
		UpdatedAt:          row.UpdatedAt,
		EncryptedPayload:   row.PayloadEncrypted,
		PayloadExpiresAt:   row.PayloadExpiresAt,
	}, nil
}

func valueOrNil(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

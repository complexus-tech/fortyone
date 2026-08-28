package messagingrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type InboundEventInput = messagingdomain.InboundEventInput
type InboundEventRecord = messagingdomain.InboundEventRecord

// RegisterInboundEvent persists the provider event identity before an HTTP
// acknowledgement. created is false for a retry already present in the inbox.
func (repository *Repository) RegisterInboundEvent(
	ctx context.Context,
	input InboundEventInput,
) (InboundEventRecord, bool, error) {
	if !repository.configured() {
		return InboundEventRecord{}, false, errors.New("messaging repository is not configured")
	}
	row, err := repository.queries.InsertInboundEvent(ctx, messagingsql.InsertInboundEventParams{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID,
		InstallationGeneration: input.InstallGeneration,
		ExternalWorkspaceID:    input.ExternalWorkspaceID,
		ExternalEventID:        input.ExternalEventID,
		EventType:              input.EventType,
		PayloadEncrypted:       input.PayloadEncrypted,
	})
	if err == nil {
		return inboundEventRecord(
			row.ID, row.WorkspaceID, row.InstallationGeneration, row.ExternalWorkspaceID,
			row.ExternalEventID, row.Status, row.AttemptCount, row.RecoveryGeneration,
			row.ProcessedAt, row.PayloadEncrypted,
		), true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return InboundEventRecord{}, false, fmt.Errorf("register messaging inbound event: %w", err)
	}
	existing, err := repository.queries.BackfillInboundEventPayload(ctx, messagingsql.BackfillInboundEventPayloadParams{
		PayloadEncrypted: input.PayloadEncrypted, Provider: input.Provider,
		ExternalWorkspaceID: input.ExternalWorkspaceID, ExternalEventID: input.ExternalEventID,
	})
	if err != nil {
		return InboundEventRecord{}, false, fmt.Errorf("read messaging inbound event: %w", err)
	}
	return inboundEventRecord(
		existing.ID, existing.WorkspaceID, existing.InstallationGeneration,
		existing.ExternalWorkspaceID, existing.ExternalEventID, existing.Status,
		existing.AttemptCount, existing.RecoveryGeneration, existing.ProcessedAt,
		existing.PayloadEncrypted,
	), false, nil
}

// ClaimRecoverableInboundEvents assigns a durable queue generation to inbox
// work that is absent, failed, or held by an expired processing lease.
func (repository *Repository) ClaimRecoverableInboundEvents(
	ctx context.Context,
	provider string,
	limit int,
) ([]InboundEventRecord, error) {
	if !repository.configured() {
		return nil, errors.New("messaging repository is not configured")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := repository.queries.ClaimRecoverableInboundEvents(ctx, messagingsql.ClaimRecoverableInboundEventsParams{
		FilterProvider: provider, LeaseSeconds: int64(messagingLeaseDuration / time.Second),
		RecoveryBaseSeconds: int32(inboundRecoveryBaseDelay / time.Second),
		RecoveryMaxShift:    inboundRecoveryMaxShift, RowLimit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("claim recoverable messaging inbound events: %w", err)
	}
	records := make([]InboundEventRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, inboundEventRecord(
			row.ID, row.WorkspaceID, row.InstallationGeneration, row.ExternalWorkspaceID,
			row.ExternalEventID, row.Status, row.AttemptCount, row.RecoveryGeneration,
			row.ProcessedAt, row.PayloadEncrypted,
		))
	}
	return records, nil
}

func (repository *Repository) MarkInboundEventQueued(ctx context.Context, id uuid.UUID) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	affected, err := repository.queries.MarkInboundEventQueued(ctx, messagingsql.MarkInboundEventQueuedParams{ID: id})
	if err != nil {
		return fmt.Errorf("mark messaging inbound event queued: %w", err)
	}
	return requireAffectedRows(affected, "mark messaging inbound event queued")
}

func (repository *Repository) ReleaseInboundEventRecovery(ctx context.Context, id uuid.UUID, generation int) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	databaseGeneration, err := safecast.Int32(generation)
	if err != nil {
		return fmt.Errorf("validate messaging inbound recovery generation: %w", err)
	}
	if err := repository.queries.ReleaseInboundEventRecovery(ctx, messagingsql.ReleaseInboundEventRecoveryParams{
		ID: id, RecoveryGeneration: databaseGeneration,
	}); err != nil {
		return fmt.Errorf("release messaging inbound event recovery: %w", err)
	}
	return nil
}

func (repository *Repository) GetInboundEvent(
	ctx context.Context,
	provider, externalWorkspaceID, externalEventID string,
) (InboundEventRecord, error) {
	if !repository.configured() {
		return InboundEventRecord{}, errors.New("messaging repository is not configured")
	}
	row, err := repository.queries.GetInboundEvent(ctx, messagingsql.GetInboundEventParams{
		Provider: provider, ExternalWorkspaceID: externalWorkspaceID, ExternalEventID: externalEventID,
	})
	if err != nil {
		return InboundEventRecord{}, fmt.Errorf("get messaging inbound event: %w", err)
	}
	return inboundEventRecord(
		row.ID, row.WorkspaceID, row.InstallationGeneration, row.ExternalWorkspaceID,
		row.ExternalEventID, row.Status, row.AttemptCount, row.RecoveryGeneration,
		row.ProcessedAt, row.PayloadEncrypted,
	), nil
}

func (repository *Repository) StartInboundEvent(
	ctx context.Context,
	provider, externalWorkspaceID, externalEventID string,
) (InboundEventRecord, bool, error) {
	if !repository.configured() {
		return InboundEventRecord{}, false, errors.New("messaging repository is not configured")
	}
	row, err := repository.queries.ClaimInboundEvent(ctx, messagingsql.ClaimInboundEventParams{
		Provider: provider, ExternalWorkspaceID: externalWorkspaceID,
		ExternalEventID: externalEventID, LeaseSeconds: int64(messagingLeaseDuration / time.Second),
	})
	if err == nil {
		return inboundEventRecord(
			row.ID, row.WorkspaceID, row.InstallationGeneration, row.ExternalWorkspaceID,
			row.ExternalEventID, row.Status, row.AttemptCount, row.RecoveryGeneration,
			row.ProcessedAt, row.PayloadEncrypted,
		), true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return InboundEventRecord{}, false, fmt.Errorf("start messaging inbound event: %w", err)
	}
	record, err := repository.GetInboundEvent(ctx, provider, externalWorkspaceID, externalEventID)
	if err != nil {
		return InboundEventRecord{}, false, fmt.Errorf("read completed messaging inbound event: %w", err)
	}
	switch record.Status {
	case "completed", "ignored", "cancelled":
		return record, false, nil
	case "processing":
		return record, false, newLeaseBusyError("messaging inbound event")
	default:
		return record, false, fmt.Errorf("messaging inbound event is unexpectedly unclaimed in status %q", record.Status)
	}
}

func (repository *Repository) CompleteInboundEvent(ctx context.Context, id uuid.UUID, status, message string) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	if status != "completed" && status != "ignored" && status != "failed" {
		return fmt.Errorf("invalid messaging event status %q", status)
	}
	var processedAt *time.Time
	if status == "completed" || status == "ignored" {
		now := time.Now().UTC()
		processedAt = &now
	}
	if err := repository.queries.CompleteInboundEvent(ctx, messagingsql.CompleteInboundEventParams{
		Status: status, LastError: message, ProcessedAt: processedAt, ID: id,
	}); err != nil {
		return fmt.Errorf("complete messaging inbound event: %w", err)
	}
	return nil
}

func inboundEventRecord(
	id uuid.UUID,
	workspaceID, installGeneration *uuid.UUID,
	externalWorkspaceID, externalEventID, status string,
	attemptCount, recoveryGeneration int32,
	processedAt *time.Time,
	payloadEncrypted *string,
) InboundEventRecord {
	return InboundEventRecord{
		ID: id, WorkspaceID: workspaceID, InstallGeneration: installGeneration,
		ExternalWorkspaceID: externalWorkspaceID, ExternalEventID: externalEventID,
		Status: status, AttemptCount: int(attemptCount), RecoveryGeneration: int(recoveryGeneration),
		ProcessedAt: processedAt, PayloadEncrypted: payloadEncrypted,
	}
}

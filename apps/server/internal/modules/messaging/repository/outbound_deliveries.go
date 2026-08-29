package messagingrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OutboundDeliveryInput struct {
	Provider                string
	WorkspaceID             uuid.UUID
	UserID                  *uuid.UUID
	InstallGeneration       *uuid.UUID
	ExternalWorkspaceID     string
	ExternalRecipientUserID string
	InboundEventID          *uuid.UUID
	IdempotencyKey          string
	ExternalChannelID       string
	ExternalThreadID        string
	Content                 string
	ProviderPayload         []byte
	Purpose                 string
	ExpiresAt               *time.Time
}

type OutboundDeliveryRecord struct {
	ID                      uuid.UUID
	WorkspaceID             uuid.UUID
	UserID                  *uuid.UUID
	InstallGeneration       *uuid.UUID
	ExternalWorkspaceID     string
	ExternalRecipientUserID *string
	InboundEventID          *uuid.UUID
	IdempotencyKey          string
	ExternalChannelID       string
	ExternalThreadID        *string
	ExternalMessageID       *string
	Content                 *string
	ProviderPayload         []byte
	Status                  string
	AttemptCount            int
	Purpose                 string
	ExpiresAt               *time.Time
}

func (repository *Repository) ListRecoverableOutboundDeliveries(
	ctx context.Context,
	provider string,
	limit int,
) ([]OutboundDeliveryRecord, error) {
	if !repository.configured() {
		return nil, errors.New("messaging repository is not configured")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := repository.queries.ListRecoverableOutboundDeliveries(ctx, messagingsql.ListRecoverableOutboundDeliveriesParams{
		Provider: provider, LeaseSeconds: int64(messagingLeaseDuration / time.Second), RowLimit: int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list recoverable messaging outbound deliveries: %w", err)
	}
	records := make([]OutboundDeliveryRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, outboundDeliveryRecord(
			row.ID, row.WorkspaceID, row.UserID, row.InstallationGeneration,
			row.ExternalWorkspaceID, row.ExternalRecipientUserID, row.InboundEventID,
			row.IdempotencyKey, row.ExternalChannelID, row.ExternalThreadID,
			row.ExternalMessageID, row.Content, row.ProviderPayload, row.Purpose,
			row.ExpiresAt, row.Status, row.AttemptCount,
		))
	}
	return records, nil
}

func (repository *Repository) StartOutboundDelivery(
	ctx context.Context,
	input OutboundDeliveryInput,
) (OutboundDeliveryRecord, bool, error) {
	externalWorkspaceID := strings.TrimSpace(input.ExternalWorkspaceID)
	if externalWorkspaceID == "" {
		return OutboundDeliveryRecord{}, false, errors.New("messaging outbound delivery external workspace id is required")
	}
	purpose := strings.TrimSpace(input.Purpose)
	if purpose == "" {
		purpose = "provider_message"
	}
	providerPayload := strings.TrimSpace(string(input.ProviderPayload))
	if providerPayload != "" && !json.Valid([]byte(providerPayload)) {
		return OutboundDeliveryRecord{}, false, errors.New("messaging outbound delivery provider payload must be valid JSON")
	}
	if !repository.configured() {
		return OutboundDeliveryRecord{}, false, errors.New("messaging repository is not configured")
	}
	row, err := repository.queries.ClaimOutboundDelivery(ctx, messagingsql.ClaimOutboundDeliveryParams{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		InstallationGeneration: input.InstallGeneration, ExternalWorkspaceID: externalWorkspaceID,
		ExternalRecipientUserID: input.ExternalRecipientUserID, InboundEventID: input.InboundEventID,
		IdempotencyKey: input.IdempotencyKey, ExternalChannelID: input.ExternalChannelID,
		ExternalThreadID: input.ExternalThreadID, Content: input.Content,
		ProviderPayload: providerPayload, Purpose: purpose, ExpiresAt: input.ExpiresAt,
		LeaseSeconds: int64(messagingLeaseDuration / time.Second),
	})
	if err == nil {
		return outboundDeliveryRecord(
			row.ID, row.WorkspaceID, row.UserID, row.InstallationGeneration,
			row.ExternalWorkspaceID, row.ExternalRecipientUserID, row.InboundEventID,
			row.IdempotencyKey, row.ExternalChannelID, row.ExternalThreadID,
			row.ExternalMessageID, row.Content, row.ProviderPayload, row.Purpose,
			row.ExpiresAt, row.Status, row.AttemptCount,
		), true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return OutboundDeliveryRecord{}, false, fmt.Errorf("start messaging outbound delivery: %w", err)
	}
	existing, err := repository.queries.GetOutboundDelivery(ctx, messagingsql.GetOutboundDeliveryParams{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID, IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		return OutboundDeliveryRecord{}, false, fmt.Errorf("read messaging outbound delivery: %w", err)
	}
	record := outboundDeliveryRecord(
		existing.ID, existing.WorkspaceID, existing.UserID, existing.InstallationGeneration,
		existing.ExternalWorkspaceID, existing.ExternalRecipientUserID, existing.InboundEventID,
		existing.IdempotencyKey, existing.ExternalChannelID, existing.ExternalThreadID,
		existing.ExternalMessageID, existing.Content, existing.ProviderPayload, existing.Purpose,
		existing.ExpiresAt, existing.Status, existing.AttemptCount,
	)
	if strings.TrimSpace(record.ExternalWorkspaceID) != externalWorkspaceID {
		return record, false, fmt.Errorf(
			"messaging outbound delivery external workspace mismatch: persisted %q, requested %q",
			record.ExternalWorkspaceID,
			externalWorkspaceID,
		)
	}
	if record.Purpose != purpose || !equalOptionalUUID(record.UserID, input.UserID) ||
		!equalOptionalUUID(record.InstallGeneration, input.InstallGeneration) ||
		strings.TrimSpace(valueOrEmptyString(record.ExternalRecipientUserID)) != strings.TrimSpace(input.ExternalRecipientUserID) {
		return record, false, errors.New("messaging outbound delivery actor or installation binding mismatch")
	}
	switch record.Status {
	case "delivered", "cancelled":
		return record, false, nil
	case "delivering":
		return record, false, newLeaseBusyError("messaging outbound delivery")
	default:
		return record, false, fmt.Errorf("messaging outbound delivery is unexpectedly unclaimed in status %q", record.Status)
	}
}

func equalOptionalUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func valueOrEmptyString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (repository *Repository) SetOutboundDeliveryContent(ctx context.Context, id uuid.UUID, content string) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	affected, err := repository.queries.SetOutboundDeliveryContent(ctx, messagingsql.SetOutboundDeliveryContentParams{
		Content: &content, ID: id,
	})
	if err != nil {
		return fmt.Errorf("set messaging outbound delivery content: %w", err)
	}
	return requireAffectedRows(affected, "set messaging outbound delivery content")
}

func (repository *Repository) SetOutboundDeliveryContentAndProviderPayload(
	ctx context.Context,
	id uuid.UUID,
	content string,
	providerPayload []byte,
) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("messaging outbound delivery content is required")
	}
	payload := strings.TrimSpace(string(providerPayload))
	if payload == "" || !json.Valid([]byte(payload)) {
		return errors.New("messaging outbound delivery provider payload must be valid JSON")
	}
	affected, err := repository.queries.SetOutboundDeliveryContentAndProviderPayload(ctx, messagingsql.SetOutboundDeliveryContentAndProviderPayloadParams{
		Content: &content, ProviderPayload: []byte(payload), ID: id,
	})
	if err != nil {
		return fmt.Errorf("set messaging outbound delivery content and provider payload: %w", err)
	}
	return requireAffectedRows(affected, "set messaging outbound delivery content and provider payload")
}

func (repository *Repository) SetOutboundDeliveryContentAndDestination(
	ctx context.Context,
	id uuid.UUID,
	content, externalChannelID, externalThreadID string,
) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	externalChannelID = strings.TrimSpace(externalChannelID)
	if externalChannelID == "" {
		return errors.New("messaging outbound delivery channel id is required")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("messaging outbound delivery content is required")
	}
	affected, err := repository.queries.SetOutboundDeliveryContentAndDestination(ctx, messagingsql.SetOutboundDeliveryContentAndDestinationParams{
		Content: &content, ExternalChannelID: externalChannelID,
		ExternalThreadID: strings.TrimSpace(externalThreadID), ID: id,
	})
	if err != nil {
		return fmt.Errorf("set messaging outbound delivery content and destination: %w", err)
	}
	return requireAffectedRows(affected, "set messaging outbound delivery content and destination")
}

func (repository *Repository) CompleteOutboundDelivery(ctx context.Context, id uuid.UUID, externalMessageID string) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	affected, err := repository.queries.CompleteOutboundDelivery(ctx, messagingsql.CompleteOutboundDeliveryParams{
		ExternalMessageID: externalMessageID, ID: id,
	})
	if err != nil {
		return fmt.Errorf("complete messaging outbound delivery: %w", err)
	}
	return requireAffectedRows(affected, "complete messaging outbound delivery")
}

func (repository *Repository) FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	if err := repository.queries.FailOutboundDelivery(ctx, messagingsql.FailOutboundDeliveryParams{
		LastError: &message, ID: id,
	}); err != nil {
		return fmt.Errorf("fail messaging outbound delivery: %w", err)
	}
	return nil
}

func (repository *Repository) CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error {
	if !repository.configured() {
		return errors.New("messaging repository is not configured")
	}
	if err := repository.queries.CancelOutboundDelivery(ctx, messagingsql.CancelOutboundDeliveryParams{
		LastError: &message, ID: id,
	}); err != nil {
		return fmt.Errorf("cancel messaging outbound delivery: %w", err)
	}
	return nil
}

func outboundDeliveryRecord(
	id, workspaceID uuid.UUID,
	userID, installationGeneration *uuid.UUID,
	externalWorkspaceID string,
	externalRecipientUserID *string,
	inboundEventID *uuid.UUID,
	idempotencyKey, externalChannelID string,
	externalThreadID, externalMessageID, content *string,
	providerPayload []byte,
	purpose string,
	expiresAt *time.Time,
	status string,
	attemptCount int32,
) OutboundDeliveryRecord {
	return OutboundDeliveryRecord{
		ID: id, WorkspaceID: workspaceID, UserID: userID,
		InstallGeneration: installationGeneration, ExternalWorkspaceID: externalWorkspaceID,
		ExternalRecipientUserID: externalRecipientUserID, InboundEventID: inboundEventID,
		IdempotencyKey: idempotencyKey, ExternalChannelID: externalChannelID,
		ExternalThreadID: externalThreadID, ExternalMessageID: externalMessageID,
		Content: content, ProviderPayload: append([]byte(nil), providerPayload...),
		Status: status, AttemptCount: int(attemptCount), Purpose: purpose, ExpiresAt: expiresAt,
	}
}

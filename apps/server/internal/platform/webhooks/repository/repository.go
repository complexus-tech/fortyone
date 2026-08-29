package webhooksrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksql "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries webhooksql.Querier
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	return &Repository{queries: webhooksql.New(pool)}
}

func (repository *Repository) Register(
	ctx context.Context,
	envelope webhooks.Envelope,
	encryptedPayload string,
	expiresAt time.Time,
) (webhooks.Record, bool, error) {
	if repository == nil || repository.queries == nil {
		return webhooks.Record{}, false, webhooks.ErrNotConfigured
	}
	if err := webhooks.ValidateEnvelope(envelope); err != nil {
		return webhooks.Record{}, false, err
	}
	if encryptedPayload == "" {
		return webhooks.Record{}, false, webhooks.ErrInvalidDelivery
	}
	if err := webhooks.ValidatePayloadRetention(envelope.ReceivedAt, expiresAt); err != nil {
		return webhooks.Record{}, false, err
	}
	workspaceID := envelope.WorkspaceID
	installationID := envelope.InstallationID
	installationGeneration := envelope.InstallationGeneration
	traceID := optionalString(envelope.TraceID)
	payloadExpiresAt := expiresAt.UTC()

	row, err := repository.queries.InsertDelivery(ctx, webhooksql.InsertDeliveryParams{
		EnvelopeVersion:        envelope.Version,
		Provider:               string(envelope.Provider),
		WorkspaceID:            &workspaceID,
		InstallationID:         &installationID,
		InstallationGeneration: &installationGeneration,
		ExternalAccountID:      envelope.ExternalAccountID,
		DeliveryID:             envelope.DeliveryID,
		EventType:              envelope.EventType,
		TraceID:                traceID,
		PayloadEncrypted:       encryptedPayload,
		PayloadExpiresAt:       &payloadExpiresAt,
		ReceivedAt:             envelope.ReceivedAt.UTC(),
	})
	if err == nil {
		record, mapErr := toRecord(row)
		return record, true, mapErr
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return webhooks.Record{}, false, fmt.Errorf("insert webhook delivery: %w", err)
	}

	row, err = repository.queries.ReadDuplicateDelivery(ctx, webhooksql.ReadDuplicateDeliveryParams{
		PayloadEncrypted:  encryptedPayload,
		PayloadExpiresAt:  &payloadExpiresAt,
		Provider:          string(envelope.Provider),
		ExternalAccountID: envelope.ExternalAccountID,
		DeliveryID:        envelope.DeliveryID,
	})
	if err != nil {
		return webhooks.Record{}, false, mapReadError("read duplicate webhook delivery", err)
	}
	record, err := toRecord(row)
	return record, false, err
}

func (repository *Repository) MarkQueued(ctx context.Context, id uuid.UUID, queuedAt time.Time) error {
	if repository == nil || repository.queries == nil {
		return webhooks.ErrNotConfigured
	}
	rows, err := repository.queries.MarkDeliveryQueued(ctx, webhooksql.MarkDeliveryQueuedParams{
		QueuedAt: queuedAt.UTC(),
		ID:       id,
	})
	return requireAffected("mark webhook delivery queued", rows, err)
}

func (repository *Repository) GetByID(ctx context.Context, id uuid.UUID) (webhooks.Record, error) {
	if repository == nil || repository.queries == nil {
		return webhooks.Record{}, webhooks.ErrNotConfigured
	}
	row, err := repository.queries.GetDeliveryByID(ctx, webhooksql.GetDeliveryByIDParams{ID: id})
	if err != nil {
		return webhooks.Record{}, mapReadError("get webhook delivery by ID", err)
	}
	return toRecord(row)
}

func (repository *Repository) GetByExternalKey(
	ctx context.Context,
	provider integrations.ProviderKey,
	externalAccountID, deliveryID string,
) (webhooks.Record, error) {
	if repository == nil || repository.queries == nil {
		return webhooks.Record{}, webhooks.ErrNotConfigured
	}
	row, err := repository.queries.GetDeliveryByExternalKey(ctx, webhooksql.GetDeliveryByExternalKeyParams{
		Provider:          string(provider),
		ExternalAccountID: externalAccountID,
		DeliveryID:        deliveryID,
	})
	if err != nil {
		return webhooks.Record{}, mapReadError("get webhook delivery by provider identity", err)
	}
	return toRecord(row)
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func mapReadError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return webhooks.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func requireAffected(operation string, rows int64, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	if rows == 0 {
		return webhooks.ErrNotFound
	}
	return nil
}

var _ webhooks.Inbox = (*Repository)(nil)

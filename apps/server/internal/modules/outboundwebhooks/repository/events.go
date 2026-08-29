package outboundwebhooksrepository

import (
	"context"
	"errors"
	"fmt"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhookssql "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) PublishEvent(
	ctx context.Context,
	event outboundwebhooksdomain.Event,
	body []byte,
) ([]uuid.UUID, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	if len(body) < 2 || len(body) > 256<<10 {
		return nil, outboundwebhooksdomain.ErrInvalidPayload
	}
	payloadVersion, err := safecast.Int32(event.PayloadVersion)
	if err != nil {
		return nil, outboundwebhooksdomain.ErrInvalidPayload
	}
	var deliveryIDs []uuid.UUID
	err = repository.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		queries := outboundwebhookssql.New(tx)
		_, err := queries.CreateOutboundWebhookEvent(ctx, outboundwebhookssql.CreateOutboundWebhookEventParams{
			EventID: event.ID, WorkspaceID: event.WorkspaceID, EventType: string(event.Type),
			PayloadVersion: payloadVersion, SubjectType: string(event.SubjectType), SubjectID: event.SubjectID,
			ActorKind: string(event.Actor.Kind), ActorID: event.Actor.PrincipalID,
			ActorCredentialID: optionalUUID(event.Actor.CredentialID), Payload: event.Payload,
			OccurredAt: event.OccurredAt.UTC(), CreatedAt: event.CreatedAt.UTC(),
		})
		if err != nil {
			return fmt.Errorf("create outbound webhook event: %w", err)
		}
		deliveryIDs, err = queries.CreateOutboundWebhookDeliveries(ctx, outboundwebhookssql.CreateOutboundWebhookDeliveriesParams{
			EventID: event.ID, PayloadBody: body, CreatedAt: event.CreatedAt.UTC(),
			WorkspaceID: event.WorkspaceID, EventType: string(event.Type),
		})
		if err != nil {
			return fmt.Errorf("create outbound webhook deliveries: %w", err)
		}
		return nil
	})
	if err == nil {
		return deliveryIDs, nil
	}
	if platformdatabase.Classify(err) != platformdatabase.ErrorClassUniqueViolation {
		return nil, err
	}

	// A caller-supplied event ID is the idempotency identity. Reusing it with
	// different tenant, type, subject, actor, timestamp, or payload is a hard
	// conflict and must never silently attach to the original deliveries.
	existing, readErr := repository.queries.GetOutboundWebhookEvent(ctx, outboundwebhookssql.GetOutboundWebhookEventParams{
		EventID: event.ID, ExpectedPayload: event.Payload,
	})
	if readErr != nil {
		return nil, fmt.Errorf("read duplicate outbound webhook event: %w", readErr)
	}
	if existing.WorkspaceID != event.WorkspaceID || existing.EventType != string(event.Type) || existing.PayloadVersion != payloadVersion ||
		existing.SubjectType != string(event.SubjectType) || existing.SubjectID != event.SubjectID ||
		existing.ActorKind != string(event.Actor.Kind) || existing.ActorID != event.Actor.PrincipalID ||
		!equalOptionalUUID(existing.ActorCredentialID, optionalUUID(event.Actor.CredentialID)) ||
		!existing.OccurredAt.Equal(event.OccurredAt.UTC()) || !existing.PayloadMatches {
		return nil, errors.Join(outboundwebhooksdomain.ErrInvalidPayload, outboundwebhooksdomain.ErrEndpointConflict)
	}
	return []uuid.UUID{}, nil
}

func equalOptionalUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

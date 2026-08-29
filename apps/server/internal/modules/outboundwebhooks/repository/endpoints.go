package outboundwebhooksrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhookssql "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/repository/sqlc"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) CreateEndpoint(
	ctx context.Context,
	input outboundwebhooksdomain.CreateEndpoint,
	secretEnvelope string,
) (outboundwebhooksdomain.Endpoint, error) {
	if err := repository.configured(); err != nil {
		return outboundwebhooksdomain.Endpoint{}, err
	}
	if err := input.Validate(); err != nil || secretEnvelope == "" {
		return outboundwebhooksdomain.Endpoint{}, errors.Join(outboundwebhooksdomain.ErrInvalidEndpoint, err)
	}
	createdBy, err := input.Actor.UserID()
	if err != nil {
		return outboundwebhooksdomain.Endpoint{}, err
	}
	var endpoint outboundwebhooksdomain.Endpoint
	err = repository.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		queries := outboundwebhookssql.New(tx)
		if _, err := queries.EnsureOutboundWebhookOwnerActive(ctx, outboundwebhookssql.EnsureOutboundWebhookOwnerActiveParams{
			PrincipalID: input.OwnerPrincipalID,
			WorkspaceID: input.WorkspaceID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return outboundwebhooksdomain.ErrEndpointOwnerInactive
			}
			return fmt.Errorf("ensure outbound webhook owner active: %w", err)
		}
		row, err := queries.CreateOutboundWebhookEndpoint(ctx, outboundwebhookssql.CreateOutboundWebhookEndpointParams{
			EndpointID: input.ID, WorkspaceID: input.WorkspaceID, OwnerPrincipalID: input.OwnerPrincipalID,
			Name: input.Name, EndpointURL: input.URL, SigningSecretEnvelope: secretEnvelope,
			CreatedByUserID: &createdBy, CreatedAt: input.CreatedAt.UTC(),
		})
		if err != nil {
			return fmt.Errorf("create outbound webhook endpoint: %w", mapWriteError(err))
		}
		for _, eventType := range input.Subscriptions {
			if err := queries.AddOutboundWebhookSubscription(ctx, outboundwebhookssql.AddOutboundWebhookSubscriptionParams{
				EndpointID: input.ID, WorkspaceID: input.WorkspaceID,
				EventType: string(eventType), CreatedAt: input.CreatedAt.UTC(),
			}); err != nil {
				return fmt.Errorf("add outbound webhook subscription: %w", err)
			}
		}
		endpoint, err = endpointFromCreate(row, input.Subscriptions)
		if err != nil {
			return err
		}
		return recordAudit(ctx, queries, auditInput{
			ID: input.AuditID, WorkspaceID: input.WorkspaceID, Actor: input.Actor,
			Operation: "endpoint.created", EndpointID: &input.ID,
			Result: "succeeded", RequestID: input.RequestID,
			Metadata: map[string]any{"subscription_count": len(input.Subscriptions)}, CreatedAt: input.CreatedAt.UTC(),
		})
	})
	if err != nil {
		return outboundwebhooksdomain.Endpoint{}, err
	}
	return endpoint, nil
}

func (repository *Repository) GetEndpoint(ctx context.Context, workspaceID, endpointID uuid.UUID) (outboundwebhooksdomain.Endpoint, error) {
	if err := repository.configured(); err != nil {
		return outboundwebhooksdomain.Endpoint{}, err
	}
	row, err := repository.queries.GetOutboundWebhookEndpoint(ctx, outboundwebhookssql.GetOutboundWebhookEndpointParams{
		EndpointID: endpointID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return outboundwebhooksdomain.Endpoint{}, fmt.Errorf("get outbound webhook endpoint: %w", mapReadError(err, outboundwebhooksdomain.ErrEndpointNotFound))
	}
	subscriptions, err := mapSubscriptions(row.Subscriptions)
	if err != nil {
		return outboundwebhooksdomain.Endpoint{}, err
	}
	return endpointFromGet(row, subscriptions)
}

func (repository *Repository) ListEndpoints(
	ctx context.Context,
	workspaceID uuid.UUID,
	cursor *outboundwebhooksdomain.EndpointCursor,
	pageSize int,
) ([]outboundwebhooksdomain.Endpoint, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	if workspaceID == uuid.Nil || pageSize < 1 || pageSize > 101 {
		return nil, outboundwebhooksdomain.ErrInvalidEndpoint
	}
	var cursorCreatedAt *time.Time
	var cursorEndpointID *uuid.UUID
	if cursor != nil {
		if err := cursor.Validate(); err != nil {
			return nil, err
		}
		createdAt := cursor.CreatedAt.UTC()
		endpointID := cursor.ID
		cursorCreatedAt = &createdAt
		cursorEndpointID = &endpointID
	}
	rows, err := repository.queries.ListOutboundWebhookEndpoints(ctx, outboundwebhookssql.ListOutboundWebhookEndpointsParams{
		WorkspaceID: workspaceID, CursorCreatedAt: cursorCreatedAt,
		CursorEndpointID: cursorEndpointID, PageSize: int32(pageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("list outbound webhook endpoints: %w", err)
	}
	endpoints := make([]outboundwebhooksdomain.Endpoint, 0, len(rows))
	for _, row := range rows {
		subscriptions, mapErr := mapSubscriptions(row.Subscriptions)
		if mapErr != nil {
			return nil, mapErr
		}
		endpoint, mapErr := endpointFromList(row, subscriptions)
		if mapErr != nil {
			return nil, mapErr
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (repository *Repository) ReplaceSubscriptions(
	ctx context.Context,
	actor platformauth.Actor,
	workspaceID, endpointID uuid.UUID,
	auditID uuid.UUID,
	subscriptions []outboundwebhooksdomain.EventType,
	now time.Time,
	requestID string,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	if auditID == uuid.Nil {
		return outboundwebhooksdomain.ErrInvalidEndpoint
	}
	if err := outboundwebhooksdomain.ValidateSubscriptions(subscriptions); err != nil {
		return err
	}
	return repository.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		queries := outboundwebhookssql.New(tx)
		if _, err := queries.LockOutboundWebhookEndpoint(ctx, outboundwebhookssql.LockOutboundWebhookEndpointParams{
			EndpointID: endpointID, WorkspaceID: workspaceID,
		}); err != nil {
			return mapReadError(err, outboundwebhooksdomain.ErrEndpointNotFound)
		}
		if err := queries.DeleteOutboundWebhookSubscriptions(ctx, outboundwebhookssql.DeleteOutboundWebhookSubscriptionsParams{
			EndpointID: endpointID, WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("delete outbound webhook subscriptions: %w", err)
		}
		for _, eventType := range subscriptions {
			if err := queries.AddOutboundWebhookSubscription(ctx, outboundwebhookssql.AddOutboundWebhookSubscriptionParams{
				EndpointID: endpointID, WorkspaceID: workspaceID,
				EventType: string(eventType), CreatedAt: now.UTC(),
			}); err != nil {
				return fmt.Errorf("replace outbound webhook subscription: %w", err)
			}
		}
		if _, err := queries.ReplaceOutboundWebhookSubscriptionGeneration(ctx, outboundwebhookssql.ReplaceOutboundWebhookSubscriptionGenerationParams{
			UpdatedAt: now.UTC(), EndpointID: endpointID, WorkspaceID: workspaceID,
		}); err != nil {
			return mapReadError(err, outboundwebhooksdomain.ErrEndpointNotFound)
		}
		errorCode := "subscription_changed"
		cancelledAt := now.UTC()
		if _, err := queries.CancelPendingOutboundWebhookDeliveries(ctx, outboundwebhookssql.CancelPendingOutboundWebhookDeliveriesParams{
			ErrorCode: &errorCode, CancelledAt: &cancelledAt, EndpointID: endpointID, WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("cancel stale outbound webhook deliveries: %w", err)
		}
		return recordAudit(ctx, queries, auditInput{
			ID: auditID, WorkspaceID: workspaceID, Actor: actor,
			Operation: "endpoint.subscriptions_replaced", EndpointID: &endpointID,
			Result: "succeeded", RequestID: requestID,
			Metadata: map[string]any{"subscription_count": len(subscriptions)}, CreatedAt: now.UTC(),
		})
	})
}

func (repository *Repository) DisableEndpoint(
	ctx context.Context,
	actor platformauth.Actor,
	workspaceID, endpointID uuid.UUID,
	auditID uuid.UUID,
	reason, requestID string,
	now time.Time,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	if auditID == uuid.Nil {
		return outboundwebhooksdomain.ErrInvalidEndpoint
	}
	return repository.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		queries := outboundwebhookssql.New(tx)
		disabledAt := now.UTC()
		if _, err := queries.DisableOutboundWebhookEndpoint(ctx, outboundwebhookssql.DisableOutboundWebhookEndpointParams{
			DisabledAt: &disabledAt, DisabledReason: &reason, EndpointID: endpointID, WorkspaceID: workspaceID,
		}); err != nil {
			return mapReadError(err, outboundwebhooksdomain.ErrEndpointNotFound)
		}
		errorCode := "endpoint_disabled"
		if _, err := queries.CancelPendingOutboundWebhookDeliveries(ctx, outboundwebhookssql.CancelPendingOutboundWebhookDeliveriesParams{
			ErrorCode: &errorCode, CancelledAt: &disabledAt, EndpointID: endpointID, WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("cancel disabled endpoint deliveries: %w", err)
		}
		return recordAudit(ctx, queries, auditInput{
			ID: auditID, WorkspaceID: workspaceID, Actor: actor, Operation: "endpoint.disabled",
			EndpointID: &endpointID, Result: "succeeded", RequestID: requestID,
			Metadata: map[string]any{"reason": reason}, CreatedAt: disabledAt,
		})
	})
}

func (repository *Repository) RotateEndpointSecret(
	ctx context.Context,
	actor platformauth.Actor,
	workspaceID, endpointID, auditID uuid.UUID,
	expectedGeneration int,
	secretEnvelope string,
	previousSecretExpiresAt, rotatedAt time.Time,
	requestID string,
) (int, error) {
	if err := repository.configured(); err != nil {
		return 0, err
	}
	if workspaceID == uuid.Nil || endpointID == uuid.Nil || auditID == uuid.Nil || expectedGeneration <= 0 ||
		secretEnvelope == "" || !previousSecretExpiresAt.After(rotatedAt) {
		return 0, outboundwebhooksdomain.ErrInvalidEndpoint
	}
	expectedDatabaseGeneration, err := safecast.Int32(expectedGeneration)
	if err != nil {
		return 0, outboundwebhooksdomain.ErrInvalidEndpoint
	}
	var generation int32
	err = repository.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		queries := outboundwebhookssql.New(tx)
		expiresAt := previousSecretExpiresAt.UTC()
		var err error
		generation, err = queries.RotateOutboundWebhookSigningSecret(ctx, outboundwebhookssql.RotateOutboundWebhookSigningSecretParams{
			PreviousSecretExpiresAt: &expiresAt, SigningSecretEnvelope: secretEnvelope,
			RotatedAt: rotatedAt.UTC(), EndpointID: endpointID, WorkspaceID: workspaceID,
			ExpectedSecretGeneration: expectedDatabaseGeneration,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return outboundwebhooksdomain.ErrEndpointConflict
			}
			return fmt.Errorf("rotate outbound webhook signing secret: %w", err)
		}
		return recordAudit(ctx, queries, auditInput{
			ID: auditID, WorkspaceID: workspaceID, Actor: actor, Operation: "endpoint.secret_rotated",
			EndpointID: &endpointID, Result: "succeeded", RequestID: requestID,
			Metadata: map[string]any{
				"previous_generation": expectedGeneration,
				"new_generation":      generation,
				"overlap_seconds":     int(previousSecretExpiresAt.Sub(rotatedAt).Seconds()),
			},
			CreatedAt: rotatedAt.UTC(),
		})
	})
	if err != nil {
		return 0, err
	}
	return int(generation), nil
}

type auditInput struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Actor       platformauth.Actor
	Operation   string
	EndpointID  *uuid.UUID
	DeliveryID  *uuid.UUID
	Result      string
	ReasonCode  string
	RequestID   string
	Metadata    map[string]any
	CreatedAt   time.Time
}

func recordAudit(ctx context.Context, queries *outboundwebhookssql.Queries, input auditInput) error {
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return fmt.Errorf("encode outbound webhook audit metadata: %w", err)
	}
	credentialID := optionalUUID(input.Actor.CredentialID)
	reasonCode := optionalString(input.ReasonCode)
	requestID := optionalString(input.RequestID)
	if err := queries.RecordOutboundWebhookAuditEvent(ctx, outboundwebhookssql.RecordOutboundWebhookAuditEventParams{
		AuditEventID: input.ID, WorkspaceID: input.WorkspaceID,
		ActorKind: string(input.Actor.Kind), ActorID: input.Actor.PrincipalID, ActorCredentialID: credentialID,
		Operation: input.Operation, EndpointID: input.EndpointID, DeliveryID: input.DeliveryID,
		Result: input.Result, ReasonCode: reasonCode, RequestID: requestID,
		Metadata: metadata, CreatedAt: input.CreatedAt,
	}); err != nil {
		return fmt.Errorf("record outbound webhook audit event: %w", err)
	}
	return nil
}

func optionalUUID(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	return &value
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

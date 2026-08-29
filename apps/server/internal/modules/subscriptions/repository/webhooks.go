package subscriptionsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	subscriptionssql "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (repository *Repository) ClaimWebhookEvent(ctx context.Context, eventID, eventType string, attemptedAt, leaseExpiresAt time.Time) (subscriptionsdomain.WebhookClaim, error) {
	ctx, span := otel.Tracer("subscriptions.repository").Start(ctx, "subscriptions.ClaimWebhookEvent")
	defer span.End()
	if err := repository.configured(); err != nil {
		return subscriptionsdomain.WebhookClaim{}, err
	}
	if strings.TrimSpace(eventID) == "" || strings.TrimSpace(eventType) == "" || len(eventID) > 255 || len(eventType) > 255 || !leaseExpiresAt.After(attemptedAt) {
		return subscriptionsdomain.WebhookClaim{}, subscriptionsdomain.ErrInvalidStripeEventIdentity
	}
	leaseToken := uuid.New()
	row, err := repository.queries.ClaimStripeWebhookEvent(ctx, subscriptionssql.ClaimStripeWebhookEventParams{
		EventID: eventID, EventType: eventType, AttemptedAt: attemptedAt,
		LeaseExpiresAt: &leaseExpiresAt, LeaseToken: &leaseToken,
	})
	if err == nil {
		if row.LeaseToken == nil {
			return subscriptionsdomain.WebhookClaim{}, errors.New("claim Stripe webhook event returned no lease token")
		}
		return subscriptionsdomain.WebhookClaim{Disposition: subscriptionsdomain.WebhookClaimAcquired, LeaseToken: *row.LeaseToken, Attempt: int(row.Attempts)}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return subscriptionsdomain.WebhookClaim{}, fmt.Errorf("claim Stripe webhook event: %w", err)
	}
	state, err := repository.queries.GetStripeWebhookClaimState(ctx, subscriptionssql.GetStripeWebhookClaimStateParams{EventID: eventID})
	if err != nil {
		return subscriptionsdomain.WebhookClaim{}, fmt.Errorf("read Stripe webhook claim state: %w", err)
	}
	if state.EventType != eventType {
		return subscriptionsdomain.WebhookClaim{}, subscriptionsdomain.ErrWebhookEventTypeConflict
	}
	switch state.ProcessingState {
	case "processed":
		return subscriptionsdomain.WebhookClaim{Disposition: subscriptionsdomain.WebhookClaimAlreadyProcessed, Attempt: int(state.Attempts)}, nil
	case "processing":
		return subscriptionsdomain.WebhookClaim{Disposition: subscriptionsdomain.WebhookClaimInProgress, Attempt: int(state.Attempts)}, nil
	default:
		return subscriptionsdomain.WebhookClaim{}, fmt.Errorf("unexpected Stripe webhook processing state %q", state.ProcessingState)
	}
}

func (repository *Repository) MarkWebhookEventProcessed(ctx context.Context, eventID string, leaseToken uuid.UUID, outcome subscriptionsdomain.WebhookOutcome, processedAt time.Time) error {
	if err := repository.configured(); err != nil {
		return err
	}
	result := string(outcome.Result)
	count, err := repository.queries.CompleteStripeWebhookEvent(ctx, subscriptionssql.CompleteStripeWebhookEventParams{
		ProcessingResult: &result, WorkspaceID: outcome.WorkspaceID, ProcessedAt: &processedAt,
		EventID: eventID, LeaseToken: &leaseToken,
	})
	return requireWebhookClaimUpdate(count, err, "complete")
}

func (repository *Repository) MarkWebhookEventFailed(ctx context.Context, eventID string, leaseToken uuid.UUID, failureCode subscriptionsdomain.WebhookFailureCode, failedAt time.Time) error {
	if err := repository.configured(); err != nil {
		return err
	}
	code := string(failureCode)
	count, err := repository.queries.FailStripeWebhookEvent(ctx, subscriptionssql.FailStripeWebhookEventParams{
		FailedAt: &failedAt, LastErrorCode: &code, EventID: eventID, LeaseToken: &leaseToken,
	})
	return requireWebhookClaimUpdate(count, err, "fail")
}

func requireWebhookClaimUpdate(count int64, err error, operation string) error {
	if err != nil {
		return fmt.Errorf("%s Stripe webhook event: %w", operation, err)
	}
	if count != 1 {
		return subscriptionsdomain.ErrWebhookEventClaimLost
	}
	return nil
}

package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"go.opentelemetry.io/otel"
)

const defaultWebhookEventLease = 5 * time.Minute

var (
	ErrInvalidWebhookSignature        = errors.New("invalid Stripe webhook signature")
	ErrWebhookEventProcessingFailed   = errors.New("stripe webhook event processing failed")
	ErrWebhookEventPersistenceFailed  = errors.New("stripe webhook event persistence failed")
	ErrWebhookEventClaimLost          = subscriptionsdomain.ErrWebhookEventClaimLost
	ErrInvalidWebhookClaimDisposition = errors.New("invalid Stripe webhook claim disposition")
)

type WebhookClaimDisposition = subscriptionsdomain.WebhookClaimDisposition

const (
	WebhookClaimAcquired         = subscriptionsdomain.WebhookClaimAcquired
	WebhookClaimAlreadyProcessed = subscriptionsdomain.WebhookClaimAlreadyProcessed
	WebhookClaimInProgress       = subscriptionsdomain.WebhookClaimInProgress
)

type WebhookProcessingResult = subscriptionsdomain.WebhookProcessingResult

const (
	WebhookResultHandled = subscriptionsdomain.WebhookResultHandled
	WebhookResultIgnored = subscriptionsdomain.WebhookResultIgnored
)

type WebhookFailureCode = subscriptionsdomain.WebhookFailureCode

const WebhookFailureHandler = subscriptionsdomain.WebhookFailureHandler

type WebhookClaim = subscriptionsdomain.WebhookClaim
type WebhookOutcome = subscriptionsdomain.WebhookOutcome

// WebhookRepository coordinates durable ownership of Stripe deliveries. The
// lease token prevents a stale worker from overwriting a newer worker's result.
type WebhookRepository interface {
	ClaimWebhookEvent(
		ctx context.Context,
		eventID string,
		eventType string,
		attemptedAt time.Time,
		leaseExpiresAt time.Time,
	) (WebhookClaim, error)
	MarkWebhookEventProcessed(
		ctx context.Context,
		eventID string,
		leaseToken uuid.UUID,
		outcome WebhookOutcome,
		processedAt time.Time,
	) error
	MarkWebhookEventFailed(
		ctx context.Context,
		eventID string,
		leaseToken uuid.UUID,
		failureCode WebhookFailureCode,
		failedAt time.Time,
	) error
}

type webhookEventProcessor interface {
	ProcessWebhookEvent(ctx context.Context, event stripe.Event) (WebhookOutcome, error)
}

type serviceWebhookEventProcessor struct {
	service *Service
}

func (p serviceWebhookEventProcessor) ProcessWebhookEvent(
	ctx context.Context,
	event stripe.Event,
) (WebhookOutcome, error) {
	switch event.Type {
	case "checkout.session.completed":
		return p.service.handleCheckoutSessionCompleted(ctx, event)
	case "customer.subscription.updated":
		return p.service.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.created":
		return p.service.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.deleted":
		return p.service.handleSubscriptionDeleted(ctx, event)
	case "invoice.paid":
		return p.service.handleInvoicePaid(ctx, event)
	default:
		return WebhookOutcome{Result: WebhookResultIgnored}, nil
	}
}

// HandleWebhookEvent verifies, claims, processes, and durably completes one
// Stripe delivery. Any uncertain non-terminal result is returned as an error so
// Stripe retries the original signed delivery.
func (s *Service) HandleWebhookEvent(ctx context.Context, payload []byte, signature string) error {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.HandleWebhookEvent")
	defer span.End()

	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		span.RecordError(ErrInvalidWebhookSignature)
		s.log.Warn(ctx, "Rejected Stripe webhook with an invalid signature")
		return ErrInvalidWebhookSignature
	}

	now := s.currentWebhookTime()
	claim, err := s.webhookRepo.ClaimWebhookEvent(
		ctx,
		event.ID,
		string(event.Type),
		now,
		now.Add(s.currentWebhookLease()),
	)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to claim Stripe webhook event", "event_id", event.ID, "event_type", event.Type)
		return fmt.Errorf("%w: claim event: %w", ErrWebhookEventPersistenceFailed, err)
	}

	switch claim.Disposition {
	case WebhookClaimAlreadyProcessed:
		s.log.Info(ctx, "Stripe webhook event was already processed", "event_id", event.ID, "event_type", event.Type)
		return nil
	case WebhookClaimInProgress:
		s.log.Info(ctx, "Stripe webhook event is already being processed", "event_id", event.ID, "event_type", event.Type)
		return ErrAlreadyProcessingEvent
	case WebhookClaimAcquired:
		if claim.LeaseToken == uuid.Nil {
			span.RecordError(ErrInvalidWebhookClaimDisposition)
			return fmt.Errorf("%w: acquired claim has no lease token", ErrInvalidWebhookClaimDisposition)
		}
	default:
		span.RecordError(ErrInvalidWebhookClaimDisposition)
		return fmt.Errorf("%w: %q", ErrInvalidWebhookClaimDisposition, claim.Disposition)
	}

	outcome, processingErr := s.webhookEvents.ProcessWebhookEvent(ctx, event)
	if processingErr != nil {
		failedAt := s.currentWebhookTime()
		if err := s.webhookRepo.MarkWebhookEventFailed(
			ctx,
			event.ID,
			claim.LeaseToken,
			WebhookFailureHandler,
			failedAt,
		); err != nil {
			span.RecordError(err)
			s.log.Error(ctx, "Failed to record Stripe webhook failure", "event_id", event.ID, "event_type", event.Type)
			return errors.Join(
				fmt.Errorf("%w: mark event failed: %w", ErrWebhookEventPersistenceFailed, err),
				ErrWebhookEventProcessingFailed,
			)
		}

		span.RecordError(ErrWebhookEventProcessingFailed)
		s.log.Error(ctx, "Stripe webhook event processing failed", "event_id", event.ID, "event_type", event.Type)
		return fmt.Errorf("%w: %v", ErrWebhookEventProcessingFailed, processingErr)
	}

	if err := s.webhookRepo.MarkWebhookEventProcessed(
		ctx,
		event.ID,
		claim.LeaseToken,
		outcome,
		s.currentWebhookTime(),
	); err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to complete Stripe webhook event", "event_id", event.ID, "event_type", event.Type)
		return fmt.Errorf("%w: complete event: %w", ErrWebhookEventPersistenceFailed, err)
	}

	s.log.Info(
		ctx,
		"Stripe webhook event reached a durable terminal state",
		"event_id", event.ID,
		"event_type", event.Type,
		"result", outcome.Result,
		"attempt", claim.Attempt,
	)
	return nil
}

func (s *Service) currentWebhookTime() time.Time {
	if s.webhookClock == nil {
		return time.Now().UTC()
	}
	return s.webhookClock().UTC()
}

func (s *Service) currentWebhookLease() time.Duration {
	if s.webhookLease <= 0 {
		return defaultWebhookEventLease
	}
	return s.webhookLease
}

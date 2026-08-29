package subscriptions

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"go.opentelemetry.io/otel"
)

// UpdateSubscriptionSeats reconciles the provider quantity to the current
// workspace membership and requests immediate proration.
func (s *Service) UpdateSubscriptionSeats(ctx context.Context, workspaceID uuid.UUID) error {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.UpdateSeats")
	defer span.End()

	hasActiveSubscription, err := s.repo.HasActiveSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("check active workspace subscription: %w", err)
	}
	if !hasActiveSubscription {
		return ErrNoActiveSubscriptionToChange
	}

	subscription, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, ErrSubscriptionNotFound) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("failed to retrieve subscription data: %w", err)
	}
	if subscription.StripeSubscriptionID == nil || *subscription.StripeSubscriptionID == "" {
		return ErrNoActiveSubscriptionToChange
	}
	if subscription.StripeSubscriptionItemID == nil || *subscription.StripeSubscriptionItemID == "" {
		return ErrSubscriptionItemNotFound
	}

	workspaceUserCount, err := s.repo.GetWorkspaceUserCount(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to determine target seat count: %w", err)
	}
	if workspaceUserCount < 1 || workspaceUserCount == subscription.SeatCount {
		return nil
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{{
			ID:       stripe.String(*subscription.StripeSubscriptionItemID),
			Quantity: stripe.Int64(int64(workspaceUserCount)),
		}},
		ProrationBehavior: stripe.String("always_invoice"),
		CancelAtPeriodEnd: stripe.Bool(false),
	}
	params.Context = ctx
	params.IdempotencyKey = stripe.String(fmt.Sprintf(
		"subscription-seats:%s:%d",
		*subscription.StripeSubscriptionID,
		workspaceUserCount,
	))

	if _, err := s.stripeClient.Subscriptions.Update(*subscription.StripeSubscriptionID, params); err != nil {
		s.log.Error(ctx, "Failed to update Stripe subscription seat quantity", "error", err, "workspace_id", workspaceID)
		return fmt.Errorf("%w: updating subscription item quantity: %v", ErrStripeOperationFailed, err)
	}

	// The signed webhook is the authority for the local seat-count update.
	return nil
}

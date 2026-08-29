package subscriptions

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"go.opentelemetry.io/otel"
)

// ChangeSubscriptionPlan changes the price on the existing subscription item.
// The local snapshot remains webhook-driven.
func (s *Service) ChangeSubscriptionPlan(ctx context.Context, workspaceID uuid.UUID, lookupKey string) error {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.ChangePlan")
	defer span.End()

	if lookupKey == "free" {
		return s.CancelSubscription(ctx, workspaceID)
	}
	if !supportedPaidLookupKey(lookupKey) {
		return ErrInvalidPriceLookupKey
	}

	subscription, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, ErrSubscriptionNotFound) {
			return ErrNoActiveSubscriptionToChange
		}
		return fmt.Errorf("failed to retrieve current subscription: %w", err)
	}
	if !changeableSubscription(subscription) {
		return ErrNoActiveSubscriptionToChange
	}

	newPriceID, err := s.lookupStripePriceID(ctx, lookupKey)
	if err != nil {
		return err
	}

	stripeSubscriptionID := *subscription.StripeSubscriptionID
	params := &stripe.SubscriptionParams{Expand: []*string{stripe.String("items.data.price")}}
	params.Context = ctx
	current, err := s.stripeClient.Subscriptions.Get(stripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("%w: fetching current stripe subscription: %v", ErrStripeOperationFailed, err)
	}
	if current == nil || len(current.Items.Data) == 0 || current.Items.Data[0] == nil || current.Items.Data[0].Price == nil {
		return ErrSubscriptionItemNotFound
	}

	item := current.Items.Data[0]
	if item.Price.ID == newPriceID {
		return ErrAlreadySubscribedToThisPlan
	}

	update := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{{
			ID:       stripe.String(item.ID),
			Price:    stripe.String(newPriceID),
			Quantity: stripe.Int64(int64(subscription.SeatCount)),
		}},
		ProrationBehavior: stripe.String("always_invoice"),
		CancelAtPeriodEnd: stripe.Bool(false),
	}
	update.Context = ctx
	update.IdempotencyKey = stripe.String(fmt.Sprintf("subscription-plan:%s:%s", stripeSubscriptionID, newPriceID))
	if _, err := s.stripeClient.Subscriptions.Update(stripeSubscriptionID, update); err != nil {
		s.log.Error(ctx, "Failed to change Stripe subscription plan", "error", err, "workspace_id", workspaceID)
		return fmt.Errorf("%w: updating subscription: %v", ErrStripeOperationFailed, err)
	}

	return nil
}

func changeableSubscription(subscription CoreWorkspaceSubscription) bool {
	if subscription.SubscriptionStatus == nil || subscription.StripeSubscriptionID == nil || *subscription.StripeSubscriptionID == "" {
		return false
	}
	return *subscription.SubscriptionStatus == StatusActive || *subscription.SubscriptionStatus == StatusTrialing
}

// CancelSubscription schedules provider cancellation at period end. The signed
// provider webhook is responsible for changing local state.
func (s *Service) CancelSubscription(ctx context.Context, workspaceID uuid.UUID) error {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.Cancel")
	defer span.End()

	subscription, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, ErrSubscriptionNotFound) {
			return ErrNoActiveSubscriptionToChange
		}
		span.RecordError(err)
		return fmt.Errorf("failed to retrieve current subscription: %w", err)
	}
	if subscription.StripeSubscriptionID == nil || *subscription.StripeSubscriptionID == "" {
		return ErrNoActiveSubscriptionToChange
	}

	stripeSubscriptionID := *subscription.StripeSubscriptionID
	getParams := &stripe.SubscriptionParams{}
	getParams.Context = ctx
	current, err := s.stripeClient.Subscriptions.Get(stripeSubscriptionID, getParams)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: fetching stripe subscription: %v", ErrStripeOperationFailed, err)
	}
	if current == nil {
		return ErrInvalidSubscription
	}
	if current.CancelAtPeriodEnd {
		return ErrSubscriptionAlreadyCanceled
	}

	params := &stripe.SubscriptionParams{CancelAtPeriodEnd: stripe.Bool(true)}
	params.Context = ctx
	params.IdempotencyKey = stripe.String("subscription-cancel-at-period-end:" + stripeSubscriptionID)
	if _, err := s.stripeClient.Subscriptions.Update(stripeSubscriptionID, params); err != nil {
		s.log.Error(ctx, "Failed to schedule Stripe subscription cancellation", "error", err, "workspace_id", workspaceID)
		span.RecordError(err)
		return fmt.Errorf("%w: updating subscription for cancellation: %v", ErrStripeOperationFailed, err)
	}

	return nil
}

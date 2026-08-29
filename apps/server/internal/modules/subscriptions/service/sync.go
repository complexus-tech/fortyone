package subscriptions

import (
	"context"
	"fmt"
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"go.opentelemetry.io/otel"
)

func (s *Service) SyncSubscription(ctx context.Context, workspaceID uuid.UUID) error {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.SyncSubscription")
	defer span.End()

	currentSubscription, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get current subscription: %w", err)
	}
	if currentSubscription.StripeSubscriptionID == nil || *currentSubscription.StripeSubscriptionID == "" {
		return ErrSubscriptionNotFound
	}

	params := &stripe.SubscriptionParams{Expand: []*string{
		stripe.String("items.data.price"),
		stripe.String("customer"),
	}}
	params.Context = ctx
	stripeSubscription, err := s.stripeClient.Subscriptions.Get(*currentSubscription.StripeSubscriptionID, params)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("%w: fetching stripe subscription: %v", ErrStripeOperationFailed, err)
	}
	if stripeSubscription == nil {
		return ErrInvalidSubscription
	}

	customerID := currentSubscription.StripeCustomerID
	if stripeSubscription.Customer != nil && stripeSubscription.Customer.ID != "" {
		customerID = stripeSubscription.Customer.ID
	}

	itemID, seatCount, tier, interval, billingEndsAt, err := stripeSubscriptionDetails(stripeSubscription)
	if err != nil {
		span.RecordError(err)
		return err
	}

	var trialEnd *time.Time
	if stripeSubscription.TrialEnd > 0 {
		value := time.Unix(stripeSubscription.TrialEnd, 0)
		trialEnd = &value
	}

	if err := s.repo.UpdateWorkspaceSubscription(ctx, workspaceID, subscriptionsdomain.SubscriptionSnapshot{
		StripeCustomerID:         customerID,
		StripeSubscriptionID:     stripeSubscription.ID,
		StripeSubscriptionItemID: &itemID,
		Status:                   SubscriptionStatus(stripeSubscription.Status),
		SeatCount:                seatCount,
		TrialEnd:                 trialEnd,
		Tier:                     tier,
		BillingInterval:          interval,
		BillingEndsAt:            billingEndsAt,
	}); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update subscription details: %w", err)
	}

	return nil
}

func stripeSubscriptionDetails(subscription *stripe.Subscription) (string, int, SubscriptionTier, *BillingInterval, *time.Time, error) {
	if subscription == nil || len(subscription.Items.Data) == 0 || subscription.Items.Data[0] == nil {
		return "", 0, TierFree, nil, nil, ErrInvalidSubscription
	}

	item := subscription.Items.Data[0]
	if item.ID == "" || item.Price == nil {
		return "", 0, TierFree, nil, nil, ErrInvalidSubscription
	}
	tier, ok := subscriptionTierForLookupKey(item.Price.LookupKey)
	if !ok {
		return "", 0, TierFree, nil, nil, fmt.Errorf("%w: unsupported Stripe price lookup key %q", ErrInvalidSubscription, item.Price.LookupKey)
	}
	var interval *BillingInterval
	if item.Price != nil && item.Price.Recurring != nil {
		value := BillingInterval(item.Price.Recurring.Interval)
		interval = &value
	}

	var billingEndsAt *time.Time
	if item.CurrentPeriodEnd > 0 {
		value := time.Unix(item.CurrentPeriodEnd, 0)
		billingEndsAt = &value
	}

	return item.ID, int(item.Quantity), tier, interval, billingEndsAt, nil
}

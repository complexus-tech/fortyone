package subscriptions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"go.opentelemetry.io/otel"
)

func (s *Service) GetInvoices(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error) {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.GetInvoices")
	defer span.End()

	invoices, err := s.repo.GetInvoicesByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to get invoices", "error", err, "workspace_id", workspaceID)
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	return invoices, nil
}

func (s *Service) GetSubscription(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error) {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.GetSubscription")
	defer span.End()

	subscription, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to get subscription", "error", err, "workspace_id", workspaceID)
		return CoreWorkspaceSubscription{}, fmt.Errorf("failed to get subscription: %w", err)
	}
	return subscription, nil
}

// SubscriptionPrice is read from the existing subscription item, never the
// public lookup key, so grandfathered rates remain visible after catalog changes.
type SubscriptionPrice struct {
	UnitAmount    int64
	Currency      string
	Interval      string
	IntervalCount int64
}

func (s *Service) GetSubscriptionPrice(ctx context.Context, subscription CoreWorkspaceSubscription) (*SubscriptionPrice, error) {
	if subscription.StripeSubscriptionID == nil || *subscription.StripeSubscriptionID == "" {
		return nil, nil
	}
	params := &stripe.SubscriptionParams{}
	params.Context = ctx
	params.AddExpand("items.data.price")
	current, err := s.stripeClient.Subscriptions.Get(*subscription.StripeSubscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription price: %w", err)
	}
	if current == nil || current.Items == nil || subscription.StripeSubscriptionItemID == nil {
		return nil, ErrSubscriptionItemNotFound
	}
	for _, item := range current.Items.Data {
		if item == nil || item.ID != *subscription.StripeSubscriptionItemID {
			continue
		}
		price := item.Price
		if price == nil || price.Recurring == nil || price.BillingScheme != stripe.PriceBillingSchemePerUnit || price.TransformQuantity != nil {
			return nil, ErrInvalidSubscription
		}
		return &SubscriptionPrice{UnitAmount: price.UnitAmount, Currency: string(price.Currency), Interval: string(price.Recurring.Interval), IntervalCount: price.Recurring.IntervalCount}, nil
	}
	return nil, ErrSubscriptionItemNotFound
}

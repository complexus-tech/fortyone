package subscriptions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"go.opentelemetry.io/otel"
)

// CreateCustomerPortalSession creates a short-lived provider portal URL for
// the workspace's bound Stripe customer.
func (s *Service) CreateCustomerPortalSession(ctx context.Context, workspaceID uuid.UUID, returnURL string) (string, error) {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.CreateCustomerPortalSession")
	defer span.End()

	returnRedirect, err := s.billingRedirect(returnURL)
	if err != nil {
		return "", err
	}

	subscription, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("get customer subscription: %w", err)
	}
	if subscription.StripeCustomerID == "" {
		return "", ErrSubscriptionNotFound
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(subscription.StripeCustomerID),
		ReturnURL: stripe.String(returnRedirect.String()),
	}
	params.Context = ctx
	session, err := s.stripeClient.BillingPortalSessions.New(params)
	if err != nil {
		s.log.Error(ctx, "Failed to create Stripe portal session", "error", err, "workspace_id", workspaceID)
		return "", fmt.Errorf("%w: creating portal session: %v", ErrStripeOperationFailed, err)
	}
	if session == nil || session.URL == "" {
		return "", fmt.Errorf("%w: Stripe returned an empty portal session", ErrStripeOperationFailed)
	}

	// Portal URLs grant account access and must never be logged.
	return session.URL, nil
}

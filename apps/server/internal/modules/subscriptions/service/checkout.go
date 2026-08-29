package subscriptions

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// CreateCheckoutSession starts a checkout for a workspace without an active
// subscription. Existing inactive customer identities are reused.
func (s *Service) CreateCheckoutSession(
	ctx context.Context,
	workspaceID uuid.UUID,
	lookupKey string,
	userEmail string,
	workspaceName string,
	successURL string,
	cancelURL string,
) (string, error) {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.CreateCheckoutSession")
	defer span.End()

	if !supportedPaidLookupKey(lookupKey) {
		return "", ErrInvalidPriceLookupKey
	}
	successRedirect, err := s.checkoutSuccessRedirect(successURL)
	if err != nil {
		return "", err
	}
	cancelRedirect, err := s.billingRedirect(cancelURL)
	if err != nil {
		return "", err
	}

	hasActiveSubscription, err := s.repo.HasActiveSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("check active workspace subscription: %w", err)
	}
	if hasActiveSubscription {
		return "", ErrWorkspaceHasActiveSub
	}

	priceID, err := s.lookupStripePriceID(ctx, lookupKey)
	if err != nil {
		return "", err
	}
	span.SetAttributes(attribute.String("stripe.requested_price_id", priceID))

	customerID, err := s.checkoutCustomerID(ctx, workspaceID, userEmail, workspaceName)
	if err != nil {
		return "", err
	}
	seatCount := s.checkoutSeatCount(ctx, workspaceID)

	params := &stripe.CheckoutSessionParams{
		Customer:            stripe.String(customerID),
		Mode:                stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ClientReferenceID:   stripe.String(workspaceID.String()),
		AllowPromotionCodes: stripe.Bool(true),
		LineItems: []*stripe.CheckoutSessionLineItemParams{{
			Price:    stripe.String(priceID),
			Quantity: stripe.Int64(int64(seatCount)),
		}},
		SuccessURL: stripe.String(successRedirect),
		CancelURL:  stripe.String(cancelRedirect.String()),
		Expand: []*string{
			stripe.String("subscription"),
			stripe.String("customer"),
		},
	}
	params.Context = ctx

	session, err := s.stripeClient.CheckoutSessions.New(params)
	if err != nil {
		s.log.Error(ctx, "Failed to create Stripe checkout session", "error", err, "workspace_id", workspaceID)
		return "", fmt.Errorf("%w: creating checkout session: %v", ErrStripeOperationFailed, err)
	}
	if session == nil || session.URL == "" {
		return "", fmt.Errorf("%w: Stripe returned an empty checkout session", ErrStripeOperationFailed)
	}

	// Checkout URLs grant access to the provider session and must never be logged.
	s.log.Info(ctx, "Stripe checkout session created", "session_id", session.ID, "workspace_id", workspaceID)
	return session.URL, nil
}

func (s *Service) checkoutCustomerID(ctx context.Context, workspaceID uuid.UUID, userEmail, workspaceName string) (string, error) {
	currentSubscription, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err == nil && currentSubscription.StripeCustomerID != "" {
		return currentSubscription.StripeCustomerID, nil
	}
	if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
		return "", fmt.Errorf("failed to retrieve workspace data for new subscription: %w", err)
	}

	params := &stripe.CustomerParams{
		Email:    stripe.String(userEmail),
		Name:     stripe.String(workspaceName),
		Metadata: map[string]string{"workspace_id": workspaceID.String()},
	}
	params.Context = ctx
	params.IdempotencyKey = stripe.String("workspace-billing-customer:" + workspaceID.String())
	customer, err := s.stripeClient.Customers.New(params)
	if err != nil {
		s.log.Error(ctx, "Failed to create Stripe customer", "error", err, "workspace_id", workspaceID)
		return "", fmt.Errorf("%w: creating customer: %v", ErrStripeOperationFailed, err)
	}
	if customer == nil || customer.ID == "" {
		return "", fmt.Errorf("%w: Stripe returned an empty customer", ErrStripeOperationFailed)
	}
	return customer.ID, nil
}

func (s *Service) checkoutSeatCount(ctx context.Context, workspaceID uuid.UUID) int {
	seatCount, err := s.repo.GetWorkspaceUserCount(ctx, workspaceID)
	if err != nil {
		s.log.Error(ctx, "Failed to get workspace user count for checkout", "error", err, "workspace_id", workspaceID)
		return 1
	}
	if seatCount < 1 {
		return 1
	}
	return seatCount
}

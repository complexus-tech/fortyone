package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
	"github.com/stripe/stripe-go/v82/webhook"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrSubscriptionNotFound       = errors.New("subscription not found")
	ErrInvoiceNotFound            = errors.New("invoice not found")
	ErrInvalidSubscription        = errors.New("invalid subscription data")
	ErrInvalidInvoice             = errors.New("invalid invoice data")
	ErrStripeOperationFailed      = errors.New("stripe operation failed")
	ErrAlreadyProcessingEvent     = errors.New("already processing event")
	ErrSubscriptionItemNotFound   = errors.New("stripe subscription item ID not found for workspace")
	ErrWorkspaceHasActiveSub      = errors.New("workspace already has an active subscription")
	ErrFailedToCreateSubscription = errors.New("failed to create subscription")
)

// Repository defines methods for accessing subscription data
type Repository interface {
	GetSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error)
	HasActiveSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) bool
	GetInvoicesByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error)
	GetWorkspaceUserCount(ctx context.Context, workspaceID uuid.UUID) (int, error)
	SaveStripeCustomerID(ctx context.Context, workspaceID uuid.UUID, customerID string) error
	UpdateSubscriptionDetails(ctx context.Context, subID, custID, itemID string, status SubscriptionStatus, seatCount int, trialEnd *time.Time, tier SubscriptionTier) error
	UpdateSubscriptionStatus(ctx context.Context, subID string, status SubscriptionStatus) error
	CreateInvoice(ctx context.Context, invoice CoreSubscriptionInvoice) error
	CreateSubscription(ctx context.Context, workspaceID uuid.UUID, stripeCustomerID string, subscriptionID string, subscriptionItemID string, status SubscriptionStatus, seatCount int, trialEnd *time.Time, tier SubscriptionTier) error
	HasEventBeenProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventAsProcessed(ctx context.Context, eventID string, eventType string, workspaceID *uuid.UUID, payload []byte) error
}

// Service provides subscription operations, now including Stripe interactions
type Service struct {
	repo               Repository
	log                *logger.Logger
	stripeClient       *client.API
	checkoutSuccessURL string
	checkoutCancelURL  string
	webhookSecret      string
}

// New creates a new subscription service with Stripe support
func New(log *logger.Logger, repo Repository, stripeClient *client.API, successURL, cancelURL, webhookSecret string) *Service {
	if stripeClient == nil {
		panic("Stripe client cannot be nil")
	}
	return &Service{
		repo:               repo,
		log:                log,
		stripeClient:       stripeClient,
		checkoutSuccessURL: successURL,
		checkoutCancelURL:  cancelURL,
		webhookSecret:      webhookSecret,
	}
}

// GetInvoices returns all invoices for a workspace
func (s *Service) GetInvoices(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error) {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.GetInvoices")
	defer span.End()

	invoices, err := s.repo.GetInvoicesByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to get invoices", "error", err, "workspace_id", workspaceID)
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	return invoices, nil
}

// GetSubscription returns the subscription for a workspace
func (s *Service) GetSubscription(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error) {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.GetSubscription")
	defer span.End()

	sub, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to get subscription", "error", err, "workspace_id", workspaceID)
		return CoreWorkspaceSubscription{}, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

// CreateCheckoutSession initiates the Stripe Checkout process for a new subscription
func (s *Service) CreateCheckoutSession(ctx context.Context, workspaceID uuid.UUID, lookupKey string, userEmail string, workspaceName string) (string, error) {
	s.log.Info(ctx, "Initiating checkout session", "workspace_id", workspaceID, "lookup_key", lookupKey)
	ctx, span := web.AddSpan(ctx, "business.subscriptions.CreateCheckoutSession")
	defer span.End()

	hasActiveSub := s.repo.HasActiveSubscriptionByWorkspaceID(ctx, workspaceID)
	if hasActiveSub {
		s.log.Info(ctx, "Workspace already has an active subscription", "workspace_id", workspaceID)
		return "", ErrWorkspaceHasActiveSub
	}

	// 1. Get current subscription/customer info
	sub, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	customerID := ""
	if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
		s.log.Error(ctx, "Failed to get subscription info", "error", err, "workspace_id", workspaceID)
		return "", fmt.Errorf("failed to retrieve workspace data: %w", err)
	}
	if err == nil { // Subscription exists
		customerID = sub.StripeCustomerID
		span.SetAttributes(attribute.String("customer_id", customerID))
		s.log.Info(ctx, "Existing Stripe customer found", "customer_id", customerID, "workspace_id", workspaceID)
	}

	// 2. Create Stripe Customer if needed
	if customerID == "" {
		s.log.Info(ctx, "Creating new Stripe customer", "workspace_id", workspaceID, "email", userEmail)
		span.AddEvent("Creating new Stripe customer", trace.WithAttributes(attribute.String("workspace_id", workspaceID.String()), attribute.String("email", userEmail)))
		custParams := &stripe.CustomerParams{
			Email: stripe.String(userEmail),
			Name:  stripe.String(workspaceName),
			Metadata: map[string]string{
				"workspace_id": workspaceID.String(),
			},
		}
		newCust, err := s.stripeClient.Customers.New(custParams)
		if err != nil {
			span.RecordError(err)
			s.log.Error(ctx, "Failed to create Stripe customer", "error", err, "workspace_id", workspaceID)
			return "", fmt.Errorf("%w: creating customer: %v", ErrStripeOperationFailed, err)
		}
		customerID = newCust.ID
		s.log.Info(ctx, "Stripe customer created successfully", "customer_id", customerID, "workspace_id", workspaceID)
	}

	// Get current workspace user count for initial seat quantity
	userCount, err := s.repo.GetWorkspaceUserCount(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to get workspace user count", "error", err, "workspace_id", workspaceID)
		userCount = 1
	}
	// Ensure we have at least 1 seat
	if userCount < 1 {
		userCount = 1
	}

	s.log.Info(ctx, "Setting initial seat count for subscription", "workspace_id", workspaceID, "seat_count", userCount)

	// lookup the price
	priceListParams := &stripe.PriceListParams{}
	priceListParams.Filters.AddFilter("lookup_keys[]", "", lookupKey)
	priceListParams.Filters.AddFilter("limit", "", "1")
	i := s.stripeClient.Prices.List(priceListParams) // use your initialized Stripe client here
	var priceID string
	for i.Next() {
		p := i.Price()
		priceID = p.ID
		break
	}

	if priceID == "" {
		s.log.Error(ctx, "No price found for lookup key", "lookup_key", lookupKey)
		return "", fmt.Errorf("no price found for lookup key: %s", lookupKey)
	}

	// 3. Create Stripe Checkout Session
	checkoutParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(int64(userCount)),
			},
		},
		SuccessURL:          stripe.String(s.checkoutSuccessURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:           stripe.String(s.checkoutCancelURL),
		CustomerEmail:       stripe.String(userEmail),
		AllowPromotionCodes: stripe.Bool(true),
		Expand: []*string{
			stripe.String("subscription"),
			stripe.String("customer"),
		},
	}

	s.log.Info(ctx, "Creating Stripe checkout session", "customer_id", customerID, "lookup_key", lookupKey)
	session, err := s.stripeClient.CheckoutSessions.New(checkoutParams)
	if err != nil {
		s.log.Error(ctx, "Failed to create Stripe checkout session", "error", err, "customer_id", customerID)
		return "", fmt.Errorf("%w: creating checkout session: %v", ErrStripeOperationFailed, err)
	}

	s.log.Info(ctx, "Stripe checkout session created successfully", "session_id", session.ID, "customer_id", customerID)
	return session.URL, nil
}

// AddSeatToSubscription increases the seat count and triggers immediate prorated billing
func (s *Service) AddSeatToSubscription(ctx context.Context, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "Attempting to add seat to subscription", "workspace_id", workspaceID)
	ctx, span := web.AddSpan(ctx, "business.subscriptions.AddSeatToSubscription")
	defer span.End()

	// 1. Get current subscription details from DB
	subData, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to get subscription for adding seat", "error", err, "workspace_id", workspaceID)
		if errors.Is(err, ErrSubscriptionNotFound) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("failed to retrieve subscription data: %w", err)
	}

	if subData.StripeSubscriptionID == nil || *subData.StripeSubscriptionID == "" {
		s.log.Error(ctx, "Cannot add seat: workspace has no active Stripe subscription ID", "workspace_id", workspaceID)
		return fmt.Errorf("workspace %s does not have an active Stripe subscription", workspaceID)
	}
	if subData.StripeSubscriptionItemID == nil || *subData.StripeSubscriptionItemID == "" {
		s.log.Error(ctx, "Cannot add seat: Stripe subscription item ID is missing", "workspace_id", workspaceID)
		return ErrSubscriptionItemNotFound
	}

	// 2. Determine the new seat count
	newUserCount, err := s.repo.GetWorkspaceUserCount(ctx, workspaceID)
	if err != nil {
		s.log.Error(ctx, "Failed to get workspace user count for adding seat", "error", err, "workspace_id", workspaceID)
		return fmt.Errorf("failed to determine target seat count: %w", err)
	}
	newQuantity := int64(newUserCount)

	// Check if quantity actually needs increasing
	if newQuantity <= int64(subData.SeatCount) {
		s.log.Info(ctx, "Seat count does not need increasing", "workspace_id", workspaceID,
			"current_seats", subData.SeatCount, "target_seats", newQuantity)
		return nil
	}

	s.log.Info(ctx, "Updating Stripe subscription item quantity",
		"workspace_id", workspaceID,
		"subscription_id", *subData.StripeSubscriptionID,
		"item_id", *subData.StripeSubscriptionItemID,
		"new_quantity", newQuantity,
	)

	// 3. Update Stripe Subscription Item
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(*subData.StripeSubscriptionItemID),
				Quantity: stripe.Int64(newQuantity),
			},
		},
		// Trigger immediate prorated invoice
		ProrationBehavior: stripe.String("always_invoice"),
	}

	// Use idempotency key
	idempotencyKey := fmt.Sprintf("add-seat-%s-%d", workspaceID.String(), time.Now().UnixNano())
	params.IdempotencyKey = stripe.String(idempotencyKey)

	_, err = s.stripeClient.Subscriptions.Update(*subData.StripeSubscriptionID, params)
	if err != nil {
		s.log.Error(ctx, "Failed to update Stripe subscription item quantity", "error", err, "workspace_id", workspaceID)
		return fmt.Errorf("%w: updating subscription item quantity: %v", ErrStripeOperationFailed, err)
	}

	s.log.Info(ctx, "Stripe subscription item quantity update request successful", "workspace_id", workspaceID)
	// The DB seat count update happens via webhook confirmation
	return nil
}

// CreateCustomerPortalSession generates a URL for the Stripe Customer Portal
func (s *Service) CreateCustomerPortalSession(ctx context.Context, workspaceID uuid.UUID, returnURL string) (string, error) {
	s.log.Info(ctx, "Creating customer portal session", "workspace_id", workspaceID)
	ctx, span := web.AddSpan(ctx, "business.subscriptions.CreateCustomerPortalSession")
	defer span.End()

	// Get subscription
	lastSub, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil || lastSub.StripeCustomerID == "" {
		s.log.Error(ctx, "Failed to get Stripe customer ID", "error", err, "workspace_id", workspaceID)
		return "", fmt.Errorf("customer not found: %w", err)
	}

	// Create portal session
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(lastSub.StripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	ps, err := s.stripeClient.BillingPortalSessions.New(params)
	if err != nil {
		s.log.Error(ctx, "Failed to create portal session", "error", err, "customer_id", lastSub.StripeCustomerID)
		return "", fmt.Errorf("%w: creating portal session: %v", ErrStripeOperationFailed, err)
	}

	s.log.Info(ctx, "Portal session created", "customer_id", lastSub.StripeCustomerID)
	return ps.URL, nil
}

// mapPriceToTier converts a Stripe Price to the internal SubscriptionTier
func (s *Service) mapPriceToTier(ctx context.Context, price *stripe.Price) SubscriptionTier {
	if price == nil {
		return TierFree
	}
	switch price.LookupKey {
	case "pro_monthly", "pro_yearly":
		return TierPro
	case "business_monthly", "business_yearly":
		return TierBusiness
	case "enterprise_monthly", "enterprise_yearly":
		return TierEnterprise
	default:
		s.log.Error(ctx, "Unknown price lookup key", "lookup_key", price.LookupKey, "price_id", price.ID)
		return TierFree
	}
}

// HandleWebhookEvent processes incoming Stripe webhook events
func (s *Service) HandleWebhookEvent(ctx context.Context, payload []byte, signature string) error {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.HandleWebhookEvent")
	defer span.End()

	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to verify webhook signature", "error", err)
		return fmt.Errorf("invalid webhook signature: %w", err)
	}
	s.log.Info(ctx, "Handling Stripe webhook event", "event_id", event.ID, "event_type", event.Type)

	// Check for idempotency
	// processed, err := s.repo.HasEventBeenProcessed(ctx, event.ID)
	// if err != nil {
	// 	span.RecordError(err)
	// 	s.log.Error(ctx, "Failed to check event processing status", "error", err, "event_id", event.ID)
	// 	return fmt.Errorf("failed to check event idempotency: %w", err)
	// }
	// if processed {
	// 	s.log.Warn(ctx, "Webhook event already processed", "event_id", event.ID, "event_type", event.Type)
	// 	return nil // Already processed
	// }

	// Extract workspace ID for logging
	var workspaceID *uuid.UUID
	// Process based on event type
	var processingError error
	switch event.Type {
	case "checkout.session.completed":
		processingError = s.handleCheckoutSessionCompleted(ctx, event)
	case "customer.subscription.updated":
		processingError = s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.created":
		processingError = s.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.deleted":
		processingError = s.handleSubscriptionDeleted(ctx, event)
	case "invoice.paid":
		processingError = s.handleInvoicePaid(ctx, event)
	default:
		s.log.Info(ctx, "Unhandled Stripe webhook event type", "event_type", event.Type)
	}

	// Mark as processed even if error occurred
	err = s.repo.MarkEventAsProcessed(ctx, event.ID, string(event.Type), workspaceID, payload)
	if err != nil {
		s.log.Error(ctx, "CRITICAL: Failed to mark event as processed", "error", err, "event_id", event.ID, "event_type", event.Type)
		return nil
	}

	if processingError != nil {
		s.log.Error(ctx, "Error processing webhook event", "error", processingError, "event_id", event.ID, "event_type", event.Type)
		return processingError
	}
	s.log.Info(ctx, "Successfully processed webhook event", "event_id", event.ID, "event_type", event.Type)
	return nil
}

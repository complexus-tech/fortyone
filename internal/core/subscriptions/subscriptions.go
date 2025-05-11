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
)

var (
	ErrSubscriptionNotFound         = errors.New("subscription not found")
	ErrInvoiceNotFound              = errors.New("invoice not found")
	ErrInvalidSubscription          = errors.New("invalid subscription data")
	ErrInvalidInvoice               = errors.New("invalid invoice data")
	ErrStripeOperationFailed        = errors.New("stripe operation failed")
	ErrAlreadyProcessingEvent       = errors.New("already processing event")
	ErrSubscriptionItemNotFound     = errors.New("stripe subscription item ID not found for workspace")
	ErrWorkspaceHasActiveSub        = errors.New("workspace already has an active subscription, use change plan flow")
	ErrFailedToCreateSubscription   = errors.New("failed to create subscription")
	ErrAlreadySubscribedToThisPlan  = errors.New("already subscribed to this specific plan")
	ErrNoActiveSubscriptionToChange = errors.New("no active subscription found to change")
	ErrSubscriptionAlreadyCanceled  = errors.New("subscription is already canceled or pending cancellation")
)

// Repository defines methods for accessing subscription data
type Repository interface {
	GetSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error)
	HasActiveSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) bool
	GetInvoicesByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error)
	GetWorkspaceUserCount(ctx context.Context, workspaceID uuid.UUID) (int, error)
	SaveStripeCustomerID(ctx context.Context, workspaceID uuid.UUID, customerID string) error
	UpdateSubscriptionDetails(ctx context.Context, subID, custID, itemID string, status SubscriptionStatus, seatCount int, trialEnd *time.Time, tier SubscriptionTier, billingInterval *BillingInterval, billingEndsAt *time.Time) error
	UpdateSubscriptionStatus(ctx context.Context, subID string, status SubscriptionStatus) error
	CreateInvoice(ctx context.Context, invoice CoreSubscriptionInvoice) error
	CreateSubscription(ctx context.Context, workspaceID uuid.UUID, stripeCustomerID string, subscriptionID string, subscriptionItemID string, status SubscriptionStatus, seatCount int, trialEnd *time.Time, tier SubscriptionTier, billingInterval *BillingInterval, billingEndsAt *time.Time) error
	HasEventBeenProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventAsProcessed(ctx context.Context, eventID string, eventType string, workspaceID *uuid.UUID, payload []byte) error
}

// Service provides subscription operations, now including Stripe interactions
type Service struct {
	repo          Repository
	log           *logger.Logger
	stripeClient  *client.API
	webhookSecret string
}

// New creates a new subscription service with Stripe support
func New(log *logger.Logger, repo Repository, stripeClient *client.API, webhookSecret string) *Service {
	if stripeClient == nil {
		panic("Stripe client cannot be nil")
	}
	return &Service{
		repo:          repo,
		log:           log,
		stripeClient:  stripeClient,
		webhookSecret: webhookSecret,
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

// CreateCheckoutSession initiates a Stripe Checkout process for a NEW subscription.
// If the workspace already has an active subscription, it returns an error.
func (s *Service) CreateCheckoutSession(ctx context.Context, workspaceID uuid.UUID, lookupKey string, userEmail string, workspaceName string, successURL string, cancelURL string) (string, error) {
	s.log.Info(ctx, "Initiating NEW checkout session request", "workspace_id", workspaceID, "lookup_key", lookupKey)
	ctx, span := web.AddSpan(ctx, "business.subscriptions.CreateCheckoutSession_NewOnly")
	defer span.End()

	// 1. Check if workspace already has an active subscription in our DB.
	// We use HasActiveSubscriptionByWorkspaceID for a quick check.
	if s.repo.HasActiveSubscriptionByWorkspaceID(ctx, workspaceID) {
		s.log.Info(ctx, "Workspace already has an active subscription. New checkout session denied.", "workspace_id", workspaceID)
		return "", ErrWorkspaceHasActiveSub // Instruct user to use change plan flow
	}

	// 2. Determine Price ID for the requested lookupKey (since it's a new sub, no need to compare with existing plan)
	requestedPriceID := ""
	priceListParams := &stripe.PriceListParams{}
	priceListParams.Filters.AddFilter("lookup_keys[]", "", lookupKey)
	priceListParams.Filters.AddFilter("limit", "", "1")
	iPrices := s.stripeClient.Prices.List(priceListParams)
	for iPrices.Next() {
		p := iPrices.Price()
		requestedPriceID = p.ID
		break
	}
	if requestedPriceID == "" {
		s.log.Error(ctx, "No Stripe price found for lookup key for new subscription", "lookup_key", lookupKey, "workspace_id", workspaceID)
		return "", fmt.Errorf("no price found for lookup key: %s", lookupKey)
	}
	span.SetAttributes(attribute.String("stripe.requested_price_id", requestedPriceID))

	// 3. Determine Stripe Customer ID
	// For a strictly new subscription flow, we expect no existing internal sub record,
	// or if one exists (e.g. a past cancelled one), we might still want to use its customer ID.
	customerID := ""
	currentInternalSub, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID) // Check if any record exists
	if err == nil && currentInternalSub.StripeCustomerID != "" {
		// A record exists (even if not active sub, could be cancelled) and has a customer ID
		customerID = currentInternalSub.StripeCustomerID
		s.log.Info(ctx, "Using existing Stripe Customer ID from DB record for new subscription flow", "customer_id", customerID, "workspace_id", workspaceID)
	} else if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
		// An actual error occurred trying to fetch from DB (not just 'not found')
		s.log.Error(ctx, "DB error when checking for existing customer ID for new subscription", "error", err, "workspace_id", workspaceID)
		return "", fmt.Errorf("failed to retrieve workspace data for new subscription: %w", err)
	}

	if customerID == "" { // No existing customer ID found, create a new Stripe Customer
		s.log.Info(ctx, "Creating new Stripe customer for new subscription.", "workspace_id", workspaceID, "email", userEmail)
		custParams := &stripe.CustomerParams{
			Email:    stripe.String(userEmail),
			Name:     stripe.String(workspaceName),
			Metadata: map[string]string{"workspace_id": workspaceID.String()},
		}
		newCust, stripeErr := s.stripeClient.Customers.New(custParams)
		if stripeErr != nil {
			s.log.Error(ctx, "Failed to create Stripe customer for new subscription", "error", stripeErr, "workspace_id", workspaceID)
			return "", fmt.Errorf("%w: creating customer: %v", ErrStripeOperationFailed, stripeErr)
		}
		customerID = newCust.ID
		s.log.Info(ctx, "Stripe customer created successfully for new subscription", "customer_id", customerID, "workspace_id", workspaceID)
	}

	// 4. Get current workspace user count for initial seat quantity
	userCount, countErr := s.repo.GetWorkspaceUserCount(ctx, workspaceID)
	if countErr != nil {
		s.log.Error(ctx, "Failed to get workspace user count for new subscription", "error", countErr, "workspace_id", workspaceID)
		userCount = 1
	}
	if userCount < 1 {
		userCount = 1
	}
	s.log.Info(ctx, "Setting seat count for new subscription checkout", "workspace_id", workspaceID, "seat_count", userCount)

	// 5. Create Stripe Checkout Session
	checkoutParams := &stripe.CheckoutSessionParams{
		Customer:          stripe.String(customerID),
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ClientReferenceID: stripe.String(workspaceID.String()),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(requestedPriceID),
				Quantity: stripe.Int64(int64(userCount)),
			},
		},
		SuccessURL: stripe.String(successURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(cancelURL),
		Expand: []*string{
			stripe.String("subscription"),
			stripe.String("customer"),
		},
	}
	if customerID != "" {
		checkoutParams.AllowPromotionCodes = stripe.Bool(true)
	}

	s.log.Info(ctx, "Creating Stripe checkout session for new subscription", "customer_id", customerID, "requested_price_id", requestedPriceID, "workspace_id", workspaceID)
	sessionStripe, err := s.stripeClient.CheckoutSessions.New(checkoutParams)
	if err != nil {
		s.log.Error(ctx, "Failed to create Stripe checkout session for new subscription", "error", err, "customer_id", customerID, "workspace_id", workspaceID)
		return "", fmt.Errorf("%w: creating checkout session: %v", ErrStripeOperationFailed, err)
	}

	s.log.Info(ctx, "Stripe checkout session for new subscription created successfully", "session_url", sessionStripe.URL, "session_id", sessionStripe.ID, "workspace_id", workspaceID)
	return sessionStripe.URL, nil
}

// UpdateSubscriptionSeats increases the seat count and triggers immediate prorated billing
func (s *Service) UpdateSubscriptionSeats(ctx context.Context, workspaceID uuid.UUID) error {
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
		ProrationBehavior: stripe.String("always_invoice"),
		CancelAtPeriodEnd: stripe.Bool(false), // Ensure the subscription doesn't get marked for cancellation
		// PaymentBehavior: stripe.String("error_if_incomplete"), // Optional: how to handle SCA failures
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
		s.log.Error(ctx, "Unknown price lookup key for tier mapping", "lookup_key", price.LookupKey, "price_id", price.ID)
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
	processed, err := s.repo.HasEventBeenProcessed(ctx, event.ID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to check event processing status", "error", err, "event_id", event.ID)
		return fmt.Errorf("failed to check event idempotency: %w", err)
	}
	if processed {
		s.log.Warn(ctx, "Webhook event already processed", "event_id", event.ID, "event_type", event.Type)
		return nil // Already processed
	}

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

// ChangeSubscriptionPlan allows a workspace to change their active subscription to a new plan or billing frequency.
func (s *Service) ChangeSubscriptionPlan(ctx context.Context, workspaceID uuid.UUID, newLookupKey string) error {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.ChangeSubscriptionPlan")
	defer span.End()
	s.log.Info(ctx, "Attempting to change subscription plan", "workspace_id", workspaceID, "new_lookup_key", newLookupKey)

	// 1. Get current active subscription from DB
	currentInternalSub, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, ErrSubscriptionNotFound) {
			s.log.Warn(ctx, "No subscription found for workspace to change plan", "workspace_id", workspaceID)
			return ErrNoActiveSubscriptionToChange
		}
		s.log.Error(ctx, "Failed to get current subscription from DB for plan change", "workspace_id", workspaceID, "error", err)
		return fmt.Errorf("failed to retrieve current subscription: %w", err)
	}

	// Check if the subscription is active or trialing
	if currentInternalSub.SubscriptionStatus == nil ||
		!(*currentInternalSub.SubscriptionStatus == StatusActive || *currentInternalSub.SubscriptionStatus == StatusTrialing) ||
		currentInternalSub.StripeSubscriptionID == nil || *currentInternalSub.StripeSubscriptionID == "" {
		s.log.Warn(ctx, "Subscription found but not in an active/trialing state or StripeSubscriptionID missing", "workspace_id", workspaceID, "status", currentInternalSub.SubscriptionStatus, "stripe_id_present", currentInternalSub.StripeSubscriptionID != nil)
		return ErrNoActiveSubscriptionToChange
	}
	currentStripeSubscriptionID := *currentInternalSub.StripeSubscriptionID

	// 2. Determine Price ID for the newLookupKey
	newPriceID := ""
	priceListParams := &stripe.PriceListParams{}
	priceListParams.Filters.AddFilter("lookup_keys[]", "", newLookupKey)
	priceListParams.Filters.AddFilter("limit", "", "1")
	iPrices := s.stripeClient.Prices.List(priceListParams)
	for iPrices.Next() {
		p := iPrices.Price()
		newPriceID = p.ID
		break
	}
	if newPriceID == "" {
		s.log.Error(ctx, "No Stripe price found for new lookup key during plan change", "new_lookup_key", newLookupKey, "workspace_id", workspaceID)
		return fmt.Errorf("no price found for new lookup key: %s", newLookupKey)
	}

	// 3. Fetch the current Stripe subscription to get its active price ID and item ID
	stripeSubParams := &stripe.SubscriptionParams{}
	stripeSubParams.AddExpand("items.data.price")
	currentStripeSub, stripeErr := s.stripeClient.Subscriptions.Get(currentStripeSubscriptionID, stripeSubParams)
	if stripeErr != nil {
		s.log.Error(ctx, "Failed to fetch current Stripe subscription for plan change", "stripe_subscription_id", currentStripeSubscriptionID, "workspace_id", workspaceID, "error", stripeErr)
		return fmt.Errorf("%w: fetching current stripe subscription: %v", ErrStripeOperationFailed, stripeErr)
	}

	if len(currentStripeSub.Items.Data) == 0 || currentStripeSub.Items.Data[0] == nil || currentStripeSub.Items.Data[0].Price == nil {
		s.log.Error(ctx, "Current Stripe subscription has no items or price data", "stripe_subscription_id", currentStripeSubscriptionID, "workspace_id", workspaceID)
		return ErrSubscriptionItemNotFound // Or a more general error
	}
	currentSubscriptionItemID := currentStripeSub.Items.Data[0].ID
	currentActivePriceID := currentStripeSub.Items.Data[0].Price.ID

	// 4. Check if already subscribed to this exact plan
	if currentActivePriceID == newPriceID {
		s.log.Info(ctx, "Workspace already subscribed to this exact plan (change plan request).", "workspace_id", workspaceID, "price_id", newPriceID, "new_lookup_key", newLookupKey)
		return ErrAlreadySubscribedToThisPlan
	}

	// 5. Construct SubscriptionParams for the update
	updateParams := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(currentSubscriptionItemID), // ID of the subscription item to update
				Price: stripe.String(newPriceID),                // ID of the new price to switch to
				// Quantity will be inherited from the existing item unless specified.
				// If you need to change quantity (seats) simultaneously, set it here:
				Quantity: stripe.Int64(int64(currentInternalSub.SeatCount)), // Explicitly set seat count
			},
		},
		ProrationBehavior: stripe.String("always_invoice"),
		CancelAtPeriodEnd: stripe.Bool(false), // Ensure the subscription doesn't get marked for cancellation
		// PaymentBehavior: stripe.String("error_if_incomplete"), // Optional: how to handle SCA failures
	}

	s.log.Info(ctx, "Updating Stripe subscription via API for plan change", "workspace_id", workspaceID, "stripe_subscription_id", currentStripeSubscriptionID, "new_price_id", newPriceID)
	_, err = s.stripeClient.Subscriptions.Update(currentStripeSubscriptionID, updateParams)
	if err != nil {
		s.log.Error(ctx, "Stripe API failed to update subscription for plan change", "stripe_subscription_id", currentStripeSubscriptionID, "workspace_id", workspaceID, "error", err)
		return fmt.Errorf("%w: updating subscription: %v", ErrStripeOperationFailed, err)
	}

	s.log.Info(ctx, "Successfully initiated subscription plan change via Stripe API. Webhooks will update DB.", "workspace_id", workspaceID, "stripe_subscription_id", currentStripeSubscriptionID, "new_price_id", newPriceID)
	return nil
}

// CancelSubscription schedules an active subscription to be canceled at the end of the current billing period.
func (s *Service) CancelSubscription(ctx context.Context, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.CancelSubscription")
	defer span.End()
	s.log.Info(ctx, "Attempting to cancel subscription at period end", "workspace_id", workspaceID)

	// 1. Get current subscription from DB
	currentInternalSub, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, ErrSubscriptionNotFound) {
			s.log.Warn(ctx, "No subscription found for workspace to cancel", "workspace_id", workspaceID)
			return ErrNoActiveSubscriptionToChange // No active subscription to cancel
		}
		s.log.Error(ctx, "Failed to get current subscription from DB for cancellation", "workspace_id", workspaceID, "error", err)
		span.RecordError(err)
		return fmt.Errorf("failed to retrieve current subscription: %w", err)
	}

	// 2. Check if StripeSubscriptionID exists
	if currentInternalSub.StripeSubscriptionID == nil || *currentInternalSub.StripeSubscriptionID == "" {
		s.log.Warn(ctx, "Subscription found but has no StripeSubscriptionID, cannot cancel via Stripe", "workspace_id", workspaceID)
		// This case might indicate an orphaned local record or a subscription not fully set up.
		// Depending on business logic, you might want a different error or handling here.
		return ErrNoActiveSubscriptionToChange // Or a more specific error like ErrInvalidSubscription
	}
	currentStripeSubscriptionID := *currentInternalSub.StripeSubscriptionID

	// 3. Fetch the subscription directly from Stripe to get its current state, especially CancelAtPeriodEnd
	stripeSub, stripeErr := s.stripeClient.Subscriptions.Get(currentStripeSubscriptionID, nil)
	if stripeErr != nil {
		s.log.Error(ctx, "Failed to fetch current Stripe subscription for cancellation check", "stripe_subscription_id", currentStripeSubscriptionID, "workspace_id", workspaceID, "error", stripeErr)
		span.RecordError(stripeErr)
		return fmt.Errorf("%w: fetching stripe subscription: %v", ErrStripeOperationFailed, stripeErr)
	}

	// 4. Check if already set to cancel at period end in Stripe
	if stripeSub.CancelAtPeriodEnd {
		s.log.Info(ctx, "Subscription is already scheduled for cancellation at period end in Stripe", "workspace_id", workspaceID, "stripe_subscription_id", stripeSub.ID)
		return ErrSubscriptionAlreadyCanceled
	}

	// 5. (Optional) Check internal status for consistency, though Stripe is the source of truth for CancelAtPeriodEnd
	if currentInternalSub.SubscriptionStatus != nil && *currentInternalSub.SubscriptionStatus == StatusCanceled {
		s.log.Warn(ctx, "Internal DB status is Canceled, but Stripe subscription is not (CancelAtPeriodEnd is false). This might indicate a sync issue.", "workspace_id", workspaceID, "stripe_subscription_id", currentStripeSubscriptionID)
		// Proceeding based on Stripe's state (CancelAtPeriodEnd=false)
	}

	// 6. Construct SubscriptionParams to set CancelAtPeriodEnd to true
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	s.log.Info(ctx, "Requesting Stripe to cancel subscription at period end", "workspace_id", workspaceID, "stripe_subscription_id", currentStripeSubscriptionID)
	updatedStripeSub, err := s.stripeClient.Subscriptions.Update(currentStripeSubscriptionID, params)
	if err != nil {
		s.log.Error(ctx, "Stripe API failed to update subscription for cancellation (CancelAtPeriodEnd=true)", "stripe_subscription_id", currentStripeSubscriptionID, "workspace_id", workspaceID, "error", err)
		span.RecordError(err)
		return fmt.Errorf("%w: updating subscription for cancellation: %v", ErrStripeOperationFailed, err)
	}

	s.log.Info(ctx, "Successfully requested Stripe to cancel subscription at period end.", "workspace_id", workspaceID, "stripe_subscription_id", updatedStripeSub.ID, "cancels_at", time.Unix(updatedStripeSub.CancelAt, 0).Format(time.RFC3339))
	// The local database will be updated via the customer.subscription.updated webhook.
	return nil
}

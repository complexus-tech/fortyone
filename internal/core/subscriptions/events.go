package subscriptions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
)

// handleCheckoutSessionCompleted handles the checkout.session.completed event triggered by the customer
// when they complete the checkout process.
func (s *Service) handleCheckoutSessionCompleted(ctx context.Context, event stripe.Event) error {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.handleCheckoutSessionCompleted")
	defer span.End()

	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("failed to unmarshal checkout session: %w", err)
	}

	// Get subscription details
	if session.Subscription == nil {
		s.log.Error(ctx, "Checkout session completed but no subscription ID found", "session_id", session.ID)
		return fmt.Errorf("checkout session completed but no subscription ID found")
	}

	// Fetch full subscription to get item ID
	subParams := &stripe.SubscriptionParams{
		Expand: []*string{
			stripe.String("items.data.price"),
			stripe.String("customer"),
		},
	}
	stripeSub, err := s.stripeClient.Subscriptions.Get(session.Subscription.ID, subParams)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription: %w", err)
	}
	if stripeSub == nil {
		return fmt.Errorf("fetched subscription is nil for session %s", session.ID)
	}

	// Get subscription item details
	var subItemID string
	var seatCount int
	var tier SubscriptionTier = TierFree
	var billingEndsAt *time.Time
	var billingInterval *BillingInterval

	if len(stripeSub.Items.Data) > 0 {
		item := stripeSub.Items.Data[0]
		subItemID = item.ID
		seatCount = int(item.Quantity)
		tier = s.mapPriceToTier(ctx, item.Price)
		if item.Price != nil && item.Price.Recurring != nil {
			intervalStr := BillingInterval(item.Price.Recurring.Interval)
			billingInterval = &intervalStr
		}
		if int64(int64(item.CurrentPeriodEnd)) > 0 {
			tempBillingEndsAt := time.Unix(int64(item.CurrentPeriodEnd), 0)
			billingEndsAt = &tempBillingEndsAt
		}
	} else {
		return fmt.Errorf("subscription %s created with no items", stripeSub.ID)
	}

	status := SubscriptionStatus(stripeSub.Status)
	var trialEnd *time.Time
	if stripeSub.TrialEnd > 0 {
		t := time.Unix(stripeSub.TrialEnd, 0)
		trialEnd = &t
	}

	// Update database
	err = s.repo.UpdateSubscriptionDetails(
		ctx, stripeSub.ID, stripeSub.Customer.ID, subItemID, status,
		seatCount, trialEnd, tier, billingInterval, billingEndsAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription details: %w", err)
	}

	s.log.Info(ctx, "Subscription details updated from event", "event_type", event.Type,
		"subscription_id", stripeSub.ID)
	return nil
}

// handleSubscriptionUpdated handles the subscription.updated event triggered by Stripe when the subscription
// is updated. This includes changes to the status, items, or other subscription details.
func (s *Service) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.handleSubscriptionUpdated")
	defer span.End()

	var stripeSub *stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	if stripeSub == nil {
		return fmt.Errorf("unmarshalled subscription is nil")
	}

	// Extract relevant data
	var subItemID string
	var seatCount int
	var tier SubscriptionTier = TierFree
	var billingEndsAt *time.Time
	var billingInterval *BillingInterval

	if len(stripeSub.Items.Data) > 0 {
		item := stripeSub.Items.Data[0]
		subItemID = item.ID
		seatCount = int(item.Quantity)
		tier = s.mapPriceToTier(ctx, item.Price)
		if item.Price != nil && item.Price.Recurring != nil {
			intervalStr := BillingInterval(item.Price.Recurring.Interval)
			billingInterval = &intervalStr
		}
		if int64(int64(item.CurrentPeriodEnd)) > 0 {
			tempBillingEndsAt := time.Unix(int64(item.CurrentPeriodEnd), 0)
			billingEndsAt = &tempBillingEndsAt
		}
	} else {
		s.log.Error(ctx, "Subscription update event received with no items", "subscription_id", stripeSub.ID)
		seatCount = 0
	}

	status := SubscriptionStatus(stripeSub.Status)
	var trialEnd *time.Time
	if stripeSub.TrialEnd > 0 {
		t := time.Unix(stripeSub.TrialEnd, 0)
		trialEnd = &t
	}

	// Update database
	if err := s.repo.UpdateSubscriptionDetails(
		ctx, stripeSub.ID, stripeSub.Customer.ID, subItemID, status,
		seatCount, trialEnd, tier, billingInterval, billingEndsAt,
	); err != nil {
		return fmt.Errorf("failed to update subscription details: %w", err)
	}

	s.log.Info(ctx, "Subscription details updated from subscription event",
		"subscription_id", stripeSub.ID,
		"status", status, "seat_count", seatCount)
	return nil
}

func (s *Service) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.handleSubscriptionDeleted")
	defer span.End()

	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription deletion: %w", err)
	}
	// Update status in DB to 'canceled'
	err := s.repo.UpdateSubscriptionStatus(ctx, stripeSub.ID, StatusCanceled)
	if err != nil {
		s.log.Error(ctx, "Failed to update subscription status to canceled in DB", "error", err,
			"subscription_id", stripeSub.ID)
	}

	s.log.Info(ctx, "Subscription marked as canceled", "subscription_id", stripeSub.ID)
	return nil
}

// handleSubscriptionCreated handles the customer.subscription.created event triggered by Stripe
// when a subscription is first created, either through checkout or directly via the API.
func (s *Service) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.handleSubscriptionCreated")
	defer span.End()

	var stripeSub *stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	if stripeSub == nil {
		return fmt.Errorf("unmarshalled subscription is nil")
	}

	// Get customer details to retrieve workspace_id
	cust, err := s.stripeClient.Customers.Get(stripeSub.Customer.ID, nil)
	if err != nil || cust.Metadata["workspace_id"] == "" {
		return fmt.Errorf("missing workspace_id metadata for subscription %s", stripeSub.ID)
	}

	workspaceIDStr := cust.Metadata["workspace_id"]
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return fmt.Errorf("invalid workspace_id format in metadata: %w", err)
	}

	// Extract relevant data
	var subItemID string
	var seatCount int
	var tier SubscriptionTier = TierFree
	var billingEndsAt *time.Time
	var billingInterval *BillingInterval

	if len(stripeSub.Items.Data) > 0 {
		item := stripeSub.Items.Data[0]
		subItemID = item.ID
		seatCount = int(item.Quantity)
		tier = s.mapPriceToTier(ctx, item.Price)
		if item.Price != nil && item.Price.Recurring != nil {
			intervalStr := BillingInterval(item.Price.Recurring.Interval)
			billingInterval = &intervalStr
		}
		if int64(int64(item.CurrentPeriodEnd)) > 0 {
			tempBillingEndsAt := time.Unix(int64(item.CurrentPeriodEnd), 0)
			billingEndsAt = &tempBillingEndsAt
		}
	} else {
		s.log.Error(ctx, "Subscription creation event received with no items", "subscription_id", stripeSub.ID)
		seatCount = 0
	}

	status := SubscriptionStatus(stripeSub.Status)
	var trialEnd *time.Time
	if stripeSub.TrialEnd > 0 {
		t := time.Unix(stripeSub.TrialEnd, 0)
		trialEnd = &t
	}

	// Create subscription in database
	if err := s.repo.CreateSubscription(ctx, workspaceID, cust.ID, stripeSub.ID, subItemID, status, seatCount, trialEnd, tier, billingInterval, billingEndsAt); err != nil {
		return ErrFailedToCreateSubscription
	}

	creatorEmail, err := s.repo.GetWorkspaceCreatorEmail(ctx, workspaceID)
	if err != nil {
		s.log.Warn(ctx, "Failed to get workspace creator email", "error", err, "workspace_id", workspaceID)
		// no need to return error this is not a critical operation
	}

	// enqueue workspace trial end task here
	if creatorEmail != "" {
		_, err = s.tasksService.EnqueueWorkspaceTrialEnd(tasks.WorkspaceTrialEndPayload{
			Email: creatorEmail,
		})
		if err != nil {
			s.log.Error(ctx, "Failed to enqueue workspace trial end task", "error", err, "workspace_id", workspaceID)
			// no need to return error this is not a critical operation
		}
	}

	s.log.Info(ctx, "New subscription created in database",
		"subscription_id", stripeSub.ID,
		"workspace_id", workspaceID,
		"status", status,
		"seat_count", seatCount)
	return nil
}

func (s *Service) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	ctx, span := web.AddSpan(ctx, "business.subscriptions.handleInvoicePaid")
	defer span.End()

	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	cust, err := s.stripeClient.Customers.Get(invoice.Customer.ID, nil)
	if err != nil || cust.Metadata["workspace_id"] == "" {
		return fmt.Errorf("missing workspace_id metadata for invoice %s", invoice.ID)
	}

	workspaceIDStr := cust.Metadata["workspace_id"]
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return fmt.Errorf("invalid workspace_id format in metadata: %w", err)
	}
	// Extract invoice details
	amountPaid := float64(invoice.AmountPaid) / 100.0
	invoiceDate := time.Unix(invoice.Created, 0)
	if invoice.StatusTransitions.PaidAt > 0 {
		invoiceDate = time.Unix(invoice.StatusTransitions.PaidAt, 0)
	}

	seatCount := 0
	if len(invoice.Lines.Data) > 0 && invoice.Lines.Data[0].Quantity > 0 {
		seatCount = int(invoice.Lines.Data[0].Quantity)
	}

	coreInvoice := CoreSubscriptionInvoice{
		WorkspaceID:     workspaceID,
		StripeInvoiceID: invoice.ID,
		AmountPaid:      amountPaid,
		InvoiceDate:     invoiceDate,
		Status:          string(invoice.Status),
		SeatsCount:      seatCount,
		CreatedAt:       time.Now(),
		HostedURL:       &invoice.HostedInvoiceURL,
		CustomerName:    &cust.Name,
	}

	err = s.repo.CreateInvoice(ctx, coreInvoice)
	if err != nil {
		return fmt.Errorf("failed to save invoice record: %w", err)
	}
	s.log.Info(ctx, "Paid invoice record saved", "invoice_id", invoice.ID)
	return nil
}

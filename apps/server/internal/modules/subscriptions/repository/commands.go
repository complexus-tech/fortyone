package subscriptionsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SaveStripeCustomerID stores or updates the Stripe customer ID for a workspace
func (r *repo) SaveStripeCustomerID(ctx context.Context, workspaceID uuid.UUID, customerID string) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.SaveStripeCustomerID")
	defer span.End()

	query := `
        INSERT INTO workspace_subscriptions (
            workspace_id,
            stripe_customer_id,
            subscription_tier,
            seat_count,
            created_at,
            updated_at
        ) VALUES (
            :workspace_id,
            :stripe_customer_id,
            'free',
            0,
            NOW(),
            NOW()
        )
        ON CONFLICT (workspace_id) 
        DO UPDATE SET 
            stripe_customer_id = :stripe_customer_id,
            updated_at = NOW()
    `

	params := map[string]any{
		"workspace_id":       workspaceID,
		"stripe_customer_id": customerID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to save Stripe customer ID: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to save Stripe customer ID"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// UpdateSubscriptionDetails updates all subscription fields in one operation
func (r *repo) UpdateSubscriptionDetails(ctx context.Context, subID, custID, itemID string, status subscriptions.SubscriptionStatus, seatCount int, trialEnd *time.Time, tier subscriptions.SubscriptionTier, billingInterval *subscriptions.BillingInterval, billingEndsAt *time.Time) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.UpdateSubscriptionDetails")
	defer span.End()

	query := `
        UPDATE workspace_subscriptions
        SET
            stripe_subscription_item_id = :stripe_subscription_item_id,
            stripe_customer_id = :stripe_customer_id,
            subscription_status = :subscription_status,
            subscription_tier = :subscription_tier,
            seat_count = :seat_count,
            trial_end_date = :trial_end_date,
            billing_interval = :billing_interval,
            billing_ends_at = :billing_ends_at,
            updated_at = NOW()
        WHERE
            stripe_subscription_id = :stripe_subscription_id
    `

	// Convert subscription status from enum to string
	statusStr := string(status)

	params := map[string]any{
		"stripe_subscription_id":      subID,
		"stripe_subscription_item_id": itemID,
		"subscription_status":         statusStr,
		"subscription_tier":           string(tier),
		"seat_count":                  seatCount,
		"trial_end_date":              trialEnd,
		"stripe_customer_id":          custID,
		"billing_interval":            nil,
		"billing_ends_at":             billingEndsAt,
	}

	if billingInterval != nil {
		params["billing_interval"] = string(*billingInterval)
	} else {
		params["billing_interval"] = nil
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to update subscription details: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update subscription details"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errMsg := fmt.Sprintf("failed to get rows affected: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get rows affected"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	if rowsAffected == 0 {
		return subscriptions.ErrSubscriptionNotFound
	}

	return nil
}

// UpdateSubscriptionStatus updates only the subscription status
func (r *repo) UpdateSubscriptionStatus(ctx context.Context, subID string, status subscriptions.SubscriptionStatus) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.UpdateSubscriptionStatus")
	defer span.End()

	query := `
        UPDATE workspace_subscriptions
        SET
            subscription_status = :subscription_status,
            updated_at = NOW()
        WHERE
          stripe_subscription_id = :stripe_subscription_id
    `

	statusStr := string(status)

	params := map[string]any{
		"stripe_subscription_id": subID,
		"subscription_status":    statusStr,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to update subscription status: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update subscription status"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errMsg := fmt.Sprintf("failed to get rows affected: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get rows affected"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	if rowsAffected == 0 {
		return subscriptions.ErrSubscriptionNotFound
	}

	return nil
}

// CreateInvoice creates a new invoice record
func (r *repo) CreateInvoice(ctx context.Context, invoice subscriptions.CoreSubscriptionInvoice) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.CreateInvoice")
	defer span.End()

	query := `
        INSERT INTO subscription_invoices (
            workspace_id,
            stripe_invoice_id,
            amount_paid,
            invoice_date,
            status,
            seats_count,
            hosted_url,
            customer_name,
            created_at
        ) VALUES (
            :workspace_id,
            :stripe_invoice_id,
            :amount_paid,
            :invoice_date,
            :status,
            :seats_count,
            :hosted_url,
            :customer_name,
            NOW()
        )
    `

	params := map[string]any{
		"workspace_id":      invoice.WorkspaceID,
		"stripe_invoice_id": invoice.StripeInvoiceID,
		"amount_paid":       invoice.AmountPaid,
		"invoice_date":      invoice.InvoiceDate,
		"status":            invoice.Status,
		"seats_count":       invoice.SeatsCount,
		"hosted_url":        invoice.HostedURL,
		"customer_name":     invoice.CustomerName,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to create invoice: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create invoice"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// CreateSubscription creates a new subscription record
func (r *repo) CreateSubscription(
	ctx context.Context,
	workspaceID uuid.UUID,
	stripeCustomerID string,
	subscriptionID string,
	subscriptionItemID string,
	status subscriptions.SubscriptionStatus,
	seatCount int,
	trialEnd *time.Time,
	tier subscriptions.SubscriptionTier,
	billingInterval *subscriptions.BillingInterval,
	billingEndsAt *time.Time,
) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.CreateSubscription")
	defer span.End()

	query := `
        INSERT INTO workspace_subscriptions (
            stripe_customer_id,
            workspace_id,
            stripe_subscription_id,
            stripe_subscription_item_id,
            subscription_status,
            subscription_tier,
            seat_count,
            trial_end_date,
            created_at,
            updated_at,
            billing_interval,
            billing_ends_at
        ) VALUES (
            :stripe_customer_id,
            :workspace_id,
            :stripe_subscription_id,
            :stripe_subscription_item_id,
            :subscription_status,
            :subscription_tier,
            :seat_count,
            :trial_end_date,
            NOW(),
            NOW(),
            :billing_interval,
            :billing_ends_at
        )
    `

	// Convert subscription status to string
	statusStr := string(status)

	params := map[string]any{
		"workspace_id":                workspaceID,
		"stripe_subscription_id":      subscriptionID,
		"stripe_subscription_item_id": subscriptionItemID,
		"subscription_status":         statusStr,
		"subscription_tier":           string(tier),
		"seat_count":                  seatCount,
		"trial_end_date":              trialEnd,
		"stripe_customer_id":          stripeCustomerID,
		"billing_interval":            nil,
		"billing_ends_at":             billingEndsAt,
	}

	if billingInterval != nil {
		params["billing_interval"] = string(*billingInterval)
	} else {
		params["billing_interval"] = nil
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to create subscription: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create subscription"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// MarkEventAsProcessed records that a webhook event has been processed
func (r *repo) MarkEventAsProcessed(ctx context.Context, eventID string, eventType string, workspaceID *uuid.UUID, payload []byte) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.MarkEventAsProcessed")
	defer span.End()

	query := `
        INSERT INTO stripe_webhook_events (
            event_id,
            event_type,
            workspace_id,
            processed_at,
            payload
        ) VALUES (
            :event_id,
            :event_type,
            :workspace_id,
            NOW(),
            :payload
        )
    `

	params := map[string]any{
		"event_id":     eventID,
		"event_type":   eventType,
		"workspace_id": workspaceID,
		"payload":      payload,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("failed to mark event as processed: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to mark event as processed"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

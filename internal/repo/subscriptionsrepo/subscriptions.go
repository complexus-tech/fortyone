package subscriptionsrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

// New creates a new subscriptions repository
func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// GetSubscriptionByWorkspaceID retrieves a workspace's subscription
func (r *repo) GetSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (subscriptions.CoreWorkspaceSubscription, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.GetSubscriptionByWorkspaceID")
	defer span.End()

	query := `
        SELECT
            workspace_id,
            stripe_customer_id,
            stripe_subscription_id,
            stripe_subscription_item_id,
            subscription_status,
            subscription_tier,
            seat_count,
            trial_end_date,
            created_at,
            updated_at
        FROM
            workspace_subscriptions
        WHERE
            workspace_id = :workspace_id
    `

	params := map[string]any{
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return subscriptions.CoreWorkspaceSubscription{}, err
	}
	defer stmt.Close()

	var sub dbWorkspaceSubscription
	if err := stmt.GetContext(ctx, &sub, params); err != nil {
		if err == sql.ErrNoRows {
			return subscriptions.CoreWorkspaceSubscription{}, subscriptions.ErrSubscriptionNotFound
		}
		errMsg := fmt.Sprintf("failed to get subscription: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get subscription"), trace.WithAttributes(attribute.String("error", errMsg)))
		return subscriptions.CoreWorkspaceSubscription{}, err
	}

	return toCoreSubscription(sub), nil
}

// GetInvoicesByWorkspaceID retrieves invoices for a workspace
func (r *repo) GetInvoicesByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]subscriptions.CoreSubscriptionInvoice, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.GetInvoicesByWorkspaceID")
	defer span.End()

	query := `
        SELECT
            invoice_id,
            workspace_id,
            stripe_invoice_id,
            amount_paid,
            invoice_date,
            status,
            seats_count,
            created_at
        FROM
            subscription_invoices
        WHERE
            workspace_id = :workspace_id
        ORDER BY
            invoice_date DESC
    `

	params := map[string]any{
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	var invoices []dbSubscriptionInvoice
	if err := stmt.SelectContext(ctx, &invoices, params); err != nil {
		errMsg := fmt.Sprintf("failed to get invoices: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get invoices"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return toCoreInvoices(invoices), nil
}

// GetWorkspaceUserCount counts the number of users in a workspace
func (r *repo) GetWorkspaceUserCount(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.GetWorkspaceUserCount")
	defer span.End()

	query := `
        SELECT COUNT(*) 
        FROM workspace_members
        WHERE workspace_id = :workspace_id
    `

	params := map[string]any{
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to count workspace users: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to count workspace users"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	return count, nil
}

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
func (r *repo) UpdateSubscriptionDetails(ctx context.Context, subID, itemID string, status subscriptions.SubscriptionStatus, seatCount int, trialEnd *time.Time, tier subscriptions.SubscriptionTier) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.UpdateSubscriptionDetails")
	defer span.End()

	query := `
        UPDATE workspace_subscriptions
        SET
            stripe_subscription_id = :stripe_subscription_id,
            stripe_subscription_item_id = :stripe_subscription_item_id,
            subscription_status = :subscription_status,
            subscription_tier = :subscription_tier,
            seat_count = :seat_count,
            trial_end_date = :trial_end_date,
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
            created_at
        ) VALUES (
            :workspace_id,
            :stripe_invoice_id,
            :amount_paid,
            :invoice_date,
            :status,
            :seats_count,
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

// HasEventBeenProcessed checks if a webhook event has already been processed
func (r *repo) HasEventBeenProcessed(ctx context.Context, eventID string) (bool, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.HasEventBeenProcessed")
	defer span.End()

	query := `
        SELECT EXISTS(
            SELECT 1 
            FROM stripe_webhook_events 
            WHERE event_id = :event_id
        )
    `

	params := map[string]any{
		"event_id": eventID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return false, err
	}
	defer stmt.Close()

	var exists bool
	if err := stmt.GetContext(ctx, &exists, params); err != nil {
		errMsg := fmt.Sprintf("failed to check event status: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to check event status"), trace.WithAttributes(attribute.String("error", errMsg)))
		return false, err
	}

	return exists, nil
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

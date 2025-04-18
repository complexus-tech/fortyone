package subscriptionsrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

// GetWorkspaceSubscription retrieves a workspace's subscription
func (r *repo) GetWorkspaceSubscription(ctx context.Context, workspaceID uuid.UUID) (subscriptions.CoreWorkspaceSubscription, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.GetWorkspaceSubscription")
	defer span.End()

	query := `
        SELECT
            workspace_id,
            stripe_customer_id,
            stripe_subscription_id,
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

	params := map[string]interface{}{
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

// CreateWorkspaceSubscription creates a new subscription record
func (r *repo) CreateWorkspaceSubscription(ctx context.Context, sub subscriptions.CoreWorkspaceSubscription) (subscriptions.CoreWorkspaceSubscription, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.CreateWorkspaceSubscription")
	defer span.End()

	query := `
        INSERT INTO workspace_subscriptions (
            workspace_id,
            stripe_customer_id,
            stripe_subscription_id,
            subscription_status,
            subscription_tier,
            seat_count,
            trial_end_date,
            created_at,
            updated_at
        ) VALUES (
            :workspace_id,
            :stripe_customer_id,
            :stripe_subscription_id,
            :subscription_status,
            :subscription_tier,
            :seat_count,
            :trial_end_date,
            :created_at,
            :updated_at
        )
        RETURNING
            workspace_id,
            stripe_customer_id,
            stripe_subscription_id,
            subscription_status,
            subscription_tier,
            seat_count,
            trial_end_date,
            created_at,
            updated_at
    `

	// Convert subscription status from enum to string if present
	var statusStr *string
	if sub.SubscriptionStatus != nil {
		s := string(*sub.SubscriptionStatus)
		statusStr = &s
	}

	params := map[string]interface{}{
		"workspace_id":           sub.WorkspaceID,
		"stripe_customer_id":     sub.StripeCustomerID,
		"stripe_subscription_id": sub.StripeSubscriptionID,
		"subscription_status":    statusStr,
		"subscription_tier":      string(sub.SubscriptionTier),
		"seat_count":             sub.SeatCount,
		"trial_end_date":         sub.TrialEndDate,
		"created_at":             sub.CreatedAt,
		"updated_at":             sub.UpdatedAt,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return subscriptions.CoreWorkspaceSubscription{}, err
	}
	defer stmt.Close()

	var dbSub dbWorkspaceSubscription
	if err := stmt.GetContext(ctx, &dbSub, params); err != nil {
		errMsg := fmt.Sprintf("failed to create subscription: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create subscription"), trace.WithAttributes(attribute.String("error", errMsg)))
		return subscriptions.CoreWorkspaceSubscription{}, err
	}

	return toCoreSubscription(dbSub), nil
}

// UpdateWorkspaceSubscription updates an existing subscription
func (r *repo) UpdateWorkspaceSubscription(ctx context.Context, sub subscriptions.CoreWorkspaceSubscription) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.UpdateWorkspaceSubscription")
	defer span.End()

	query := `
        UPDATE workspace_subscriptions
        SET
            stripe_customer_id = :stripe_customer_id,
            stripe_subscription_id = :stripe_subscription_id,
            subscription_status = :subscription_status,
            subscription_tier = :subscription_tier,
            seat_count = :seat_count,
            trial_end_date = :trial_end_date,
            updated_at = :updated_at
        WHERE
            workspace_id = :workspace_id
    `

	// Convert subscription status from enum to string if present
	var statusStr *string
	if sub.SubscriptionStatus != nil {
		s := string(*sub.SubscriptionStatus)
		statusStr = &s
	}

	params := map[string]interface{}{
		"workspace_id":           sub.WorkspaceID,
		"stripe_customer_id":     sub.StripeCustomerID,
		"stripe_subscription_id": sub.StripeSubscriptionID,
		"subscription_status":    statusStr,
		"subscription_tier":      string(sub.SubscriptionTier),
		"seat_count":             sub.SeatCount,
		"trial_end_date":         sub.TrialEndDate,
		"updated_at":             sub.UpdatedAt,
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
		errMsg := fmt.Sprintf("failed to update subscription: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update subscription"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// CreateSubscriptionInvoice creates a record of an invoice
func (r *repo) CreateSubscriptionInvoice(ctx context.Context, invoice subscriptions.CoreSubscriptionInvoice) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.CreateSubscriptionInvoice")
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
            :created_at
        )
    `

	params := map[string]interface{}{
		"workspace_id":      invoice.WorkspaceID,
		"stripe_invoice_id": invoice.StripeInvoiceID,
		"amount_paid":       invoice.AmountPaid,
		"invoice_date":      invoice.InvoiceDate,
		"status":            invoice.Status,
		"seats_count":       invoice.SeatsCount,
		"created_at":        invoice.CreatedAt,
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
		errMsg := fmt.Sprintf("failed to create invoice record: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create invoice record"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	return nil
}

// ListSubscriptionInvoices lists invoices for a workspace
func (r *repo) ListSubscriptionInvoices(ctx context.Context, workspaceID uuid.UUID) ([]subscriptions.CoreSubscriptionInvoice, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.ListSubscriptionInvoices")
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

	params := map[string]interface{}{
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

	var dbInvoices []dbSubscriptionInvoice
	if err := stmt.SelectContext(ctx, &dbInvoices, params); err != nil {
		errMsg := fmt.Sprintf("failed to list invoices: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to list invoices"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return toCoreInvoices(dbInvoices), nil
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

	params := map[string]interface{}{
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

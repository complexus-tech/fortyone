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

// GetSubscriptionByWorkspaceID retrieves a workspace's subscription
func (r *repo) GetSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (subscriptions.CoreWorkspaceSubscription, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.GetSubscriptionByWorkspaceID")
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

// CreateSubscription creates a new subscription record
func (r *repo) CreateSubscription(ctx context.Context, sub subscriptions.CoreWorkspaceSubscription) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.CreateSubscription")
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
    `

	// Convert subscription status from enum to string if present
	var statusStr *string
	if sub.SubscriptionStatus != nil {
		s := string(*sub.SubscriptionStatus)
		statusStr = &s
	}

	params := map[string]any{
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

// UpdateSubscription updates an existing subscription
func (r *repo) UpdateSubscription(ctx context.Context, sub subscriptions.CoreWorkspaceSubscription) error {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.UpdateSubscription")
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

	params := map[string]any{
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

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to update subscription: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update subscription"), trace.WithAttributes(attribute.String("error", errMsg)))
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

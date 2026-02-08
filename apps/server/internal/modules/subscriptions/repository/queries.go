package subscriptionsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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
            updated_at,
            billing_interval,
            billing_ends_at
        FROM
            workspace_subscriptions
        WHERE
            workspace_id = :workspace_id
        ORDER BY
            created_at DESC
        LIMIT 1
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
            hosted_url,
            customer_name,
            created_at
        FROM
            subscription_invoices
        WHERE
            workspace_id = :workspace_id
        ORDER BY
            invoice_date DESC
				LIMIT 5
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
				INNER JOIN users u ON workspace_members.user_id = u.user_id
        WHERE workspace_id = :workspace_id
        AND role IN ('admin', 'member')
        AND u.is_active = TRUE
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

// HasActiveSubscriptionByWorkspaceID checks if a currently active (or trialing/past_due) subscription exists for a workspace.
// It returns false if no active subscription exists or if an error occurs.
func (r *repo) HasActiveSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) bool {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.HasActiveSubscriptionByWorkspaceID")
	defer span.End()

	query := `
        SELECT
            1
        FROM
            workspace_subscriptions
        WHERE
            workspace_id = :workspace_id
        AND (
            subscription_status = :status_active OR
            subscription_status = :status_trialing OR
            subscription_status = :status_past_due
        )
        LIMIT 1
    `

	params := map[string]any{
		"workspace_id":    workspaceID,
		"status_active":   string(subscriptions.StatusActive),
		"status_trialing": string(subscriptions.StatusTrialing),
		"status_past_due": string(subscriptions.StatusPastDue),
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement for active subscription check: %s", err)
		r.log.Error(ctx, errMsg, "workspace_id", workspaceID)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return false
	}
	defer stmt.Close()

	var exists int
	if err := stmt.GetContext(ctx, &exists, params); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			span.SetAttributes(attribute.Bool("active_subscription_found", false))
			return false
		}
		errMsg := fmt.Sprintf("failed to execute query for active subscription check: %s", err)
		r.log.Error(ctx, errMsg, "workspace_id", workspaceID)
		span.RecordError(errors.New("failed to get active subscription check"), trace.WithAttributes(attribute.String("error", errMsg)))
		return false
	}

	span.SetAttributes(attribute.Bool("active_subscription_found", true))
	return true
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

// Get workspace creator email
func (r *repo) GetWorkspaceCreatorEmail(ctx context.Context, workspaceID uuid.UUID) (string, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.subscriptions.GetWorkspaceCreatorEmail")
	defer span.End()

	query := `
		SELECT u.email
		FROM users u
		INNER JOIN workspaces w ON u.user_id = w.created_by
		WHERE w.workspace_id = :workspace_id
		AND u.is_active = TRUE
	`

	params := map[string]any{
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return "", err
	}
	defer stmt.Close()

	var email string
	if err := stmt.GetContext(ctx, &email, params); err != nil {
		errMsg := fmt.Sprintf("failed to get workspace creator email: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get workspace creator email"), trace.WithAttributes(attribute.String("error", errMsg)))
		return "", err
	}
	return email, nil
}

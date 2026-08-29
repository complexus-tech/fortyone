package subscriptionsrepository

import (
	"context"
	"errors"
	"fmt"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	subscriptionssql "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

const invoiceHistoryLimit = 5

func (repository *Repository) GetSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (subscriptionsdomain.WorkspaceSubscription, error) {
	ctx, span := otel.Tracer("subscriptions.repository").Start(ctx, "subscriptions.GetByWorkspace")
	defer span.End()
	if err := repository.configured(); err != nil {
		return subscriptionsdomain.WorkspaceSubscription{}, err
	}
	row, err := repository.queries.GetSubscriptionByWorkspaceID(ctx, subscriptionssql.GetSubscriptionByWorkspaceIDParams{WorkspaceID: workspaceID})
	if errors.Is(err, pgx.ErrNoRows) {
		return subscriptionsdomain.WorkspaceSubscription{}, subscriptionsdomain.ErrSubscriptionNotFound
	}
	if err != nil {
		return subscriptionsdomain.WorkspaceSubscription{}, fmt.Errorf("get workspace subscription: %w", err)
	}
	return toDomainSubscription(row), nil
}

func (repository *Repository) GetInvoicesByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]subscriptionsdomain.SubscriptionInvoice, error) {
	ctx, span := otel.Tracer("subscriptions.repository").Start(ctx, "subscriptions.ListInvoices")
	defer span.End()
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListWorkspaceInvoices(ctx, subscriptionssql.ListWorkspaceInvoicesParams{WorkspaceID: workspaceID, ResultLimit: invoiceHistoryLimit})
	if err != nil {
		return nil, fmt.Errorf("list workspace invoices: %w", err)
	}
	result := make([]subscriptionsdomain.SubscriptionInvoice, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainInvoice(row))
	}
	return result, nil
}

func (repository *Repository) GetWorkspaceUserCount(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	if err := repository.configured(); err != nil {
		return 0, err
	}
	count, err := repository.queries.CountBillableWorkspaceUsers(ctx, subscriptionssql.CountBillableWorkspaceUsersParams{WorkspaceID: workspaceID})
	if err != nil {
		return 0, fmt.Errorf("count billable workspace users: %w", err)
	}
	return int(count), nil
}

func (repository *Repository) HasActiveSubscriptionByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	if err := repository.configured(); err != nil {
		return false, err
	}
	exists, err := repository.queries.HasActiveWorkspaceSubscription(ctx, subscriptionssql.HasActiveWorkspaceSubscriptionParams{WorkspaceID: workspaceID})
	if err != nil {
		return false, fmt.Errorf("check active workspace subscription: %w", err)
	}
	return exists, nil
}

func (repository *Repository) GetWorkspaceCreatorEmail(ctx context.Context, workspaceID uuid.UUID) (string, error) {
	if err := repository.configured(); err != nil {
		return "", err
	}
	email, err := repository.queries.GetWorkspaceCreatorEmail(ctx, subscriptionssql.GetWorkspaceCreatorEmailParams{WorkspaceID: workspaceID})
	if err != nil {
		return "", fmt.Errorf("get workspace creator email: %w", err)
	}
	return email, nil
}

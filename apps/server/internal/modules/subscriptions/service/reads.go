package subscriptions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

func (s *Service) GetInvoices(ctx context.Context, workspaceID uuid.UUID) ([]CoreSubscriptionInvoice, error) {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.GetInvoices")
	defer span.End()

	invoices, err := s.repo.GetInvoicesByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to get invoices", "error", err, "workspace_id", workspaceID)
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	return invoices, nil
}

func (s *Service) GetSubscription(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSubscription, error) {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.GetSubscription")
	defer span.End()

	subscription, err := s.repo.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Failed to get subscription", "error", err, "workspace_id", workspaceID)
		return CoreWorkspaceSubscription{}, fmt.Errorf("failed to get subscription: %w", err)
	}
	return subscription, nil
}

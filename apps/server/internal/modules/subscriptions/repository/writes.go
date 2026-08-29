package subscriptionsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	subscriptionssql "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) UpdateWorkspaceSubscription(ctx context.Context, workspaceID uuid.UUID, snapshot subscriptionsdomain.SubscriptionSnapshot) error {
	if err := repository.configured(); err != nil {
		return err
	}
	seatCount, err := subscriptionSeatCount(snapshot.SeatCount)
	if err != nil {
		return err
	}
	status, tier, interval := subscriptionParams(snapshot)
	count, err := repository.queries.UpdateWorkspaceSubscriptionSnapshot(ctx, subscriptionssql.UpdateWorkspaceSubscriptionSnapshotParams{
		StripeSubscriptionItemID: snapshot.StripeSubscriptionItemID, StripeCustomerID: snapshot.StripeCustomerID,
		SubscriptionStatus: status, SubscriptionTier: tier, SeatCount: seatCount,
		TrialEndDate: snapshot.TrialEnd, BillingInterval: interval, BillingEndsAt: snapshot.BillingEndsAt,
		WorkspaceID: workspaceID, StripeSubscriptionID: snapshot.StripeSubscriptionID,
	})
	if err != nil {
		return fmt.Errorf("update workspace subscription: %w", err)
	}
	if count != 1 {
		conflict, lookupErr := repository.hasCustomerBindingOutsideWorkspace(ctx, snapshot.StripeCustomerID, workspaceID)
		if lookupErr != nil {
			return lookupErr
		}
		if conflict {
			return subscriptionsdomain.ErrProviderIdentityConflict
		}
		return subscriptionsdomain.ErrSubscriptionNotFound
	}
	return nil
}

func (repository *Repository) ApplyStripeSubscriptionSnapshot(ctx context.Context, snapshot subscriptionsdomain.SubscriptionSnapshot, cursor subscriptionsdomain.StripeEventCursor) (subscriptionsdomain.SubscriptionMutation, error) {
	if err := repository.configured(); err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, err
	}
	if err := validateCursor(cursor); err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, err
	}
	seatCount, err := subscriptionSeatCount(snapshot.SeatCount)
	if err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, err
	}
	status, tier, interval := subscriptionParams(snapshot)
	row, err := repository.queries.ApplyStripeSubscriptionSnapshot(ctx, subscriptionssql.ApplyStripeSubscriptionSnapshotParams{
		StripeSubscriptionItemID: snapshot.StripeSubscriptionItemID, StripeCustomerID: snapshot.StripeCustomerID,
		SubscriptionStatus: status, SubscriptionTier: tier, SeatCount: seatCount,
		TrialEndDate: snapshot.TrialEnd, BillingInterval: interval, BillingEndsAt: snapshot.BillingEndsAt,
		EventCreatedAt: cursor.CreatedAt, EventPriority: cursor.Priority, EventID: cursor.EventID,
		StripeSubscriptionID: snapshot.StripeSubscriptionID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return subscriptionsdomain.SubscriptionMutation{}, subscriptionsdomain.ErrSubscriptionNotFound
	}
	if err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, fmt.Errorf("apply Stripe subscription snapshot: %w", err)
	}
	if row.IdentityConflict {
		return subscriptionsdomain.SubscriptionMutation{}, subscriptionsdomain.ErrProviderIdentityConflict
	}
	return subscriptionsdomain.SubscriptionMutation{WorkspaceID: row.WorkspaceID, Applied: row.Applied}, nil
}

func (repository *Repository) UpsertStripeSubscription(ctx context.Context, workspaceID uuid.UUID, snapshot subscriptionsdomain.SubscriptionSnapshot, cursor subscriptionsdomain.StripeEventCursor) (subscriptionsdomain.SubscriptionMutation, error) {
	if err := repository.configured(); err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, err
	}
	if err := validateCursor(cursor); err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, err
	}
	seatCount, err := subscriptionSeatCount(snapshot.SeatCount)
	if err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, err
	}
	status, tier, interval := subscriptionParams(snapshot)
	row, err := repository.queries.UpsertStripeSubscriptionSnapshot(ctx, subscriptionssql.UpsertStripeSubscriptionSnapshotParams{
		WorkspaceID: workspaceID, StripeCustomerID: snapshot.StripeCustomerID,
		StripeSubscriptionID: snapshot.StripeSubscriptionID, StripeSubscriptionItemID: snapshot.StripeSubscriptionItemID,
		SubscriptionStatus: status, SubscriptionTier: tier, SeatCount: seatCount,
		TrialEndDate: snapshot.TrialEnd, BillingInterval: interval, BillingEndsAt: snapshot.BillingEndsAt,
		EventCreatedAt: cursor.CreatedAt, EventPriority: cursor.Priority, EventID: cursor.EventID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		conflict, lookupErr := repository.hasCustomerBindingOutsideWorkspace(ctx, snapshot.StripeCustomerID, workspaceID)
		if lookupErr != nil {
			return subscriptionsdomain.SubscriptionMutation{}, lookupErr
		}
		if conflict {
			return subscriptionsdomain.SubscriptionMutation{}, subscriptionsdomain.ErrProviderIdentityConflict
		}
	}
	if err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, fmt.Errorf("upsert Stripe subscription: %w", err)
	}
	if row.WorkspaceID != workspaceID {
		return subscriptionsdomain.SubscriptionMutation{}, subscriptionsdomain.ErrProviderIdentityConflict
	}
	if row.IdentityConflict {
		return subscriptionsdomain.SubscriptionMutation{}, subscriptionsdomain.ErrProviderIdentityConflict
	}
	return subscriptionsdomain.SubscriptionMutation{WorkspaceID: row.WorkspaceID, Applied: row.Applied}, nil
}

func (repository *Repository) ApplyStripeSubscriptionDeletion(ctx context.Context, subscriptionID string, cursor subscriptionsdomain.StripeEventCursor) (subscriptionsdomain.SubscriptionMutation, error) {
	if err := repository.configured(); err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, err
	}
	if err := validateCursor(cursor); err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, err
	}
	row, err := repository.queries.ApplyStripeSubscriptionDeletion(ctx, subscriptionssql.ApplyStripeSubscriptionDeletionParams{
		EventCreatedAt: cursor.CreatedAt, EventPriority: cursor.Priority, EventID: cursor.EventID,
		StripeSubscriptionID: subscriptionID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return subscriptionsdomain.SubscriptionMutation{}, subscriptionsdomain.ErrSubscriptionNotFound
	}
	if err != nil {
		return subscriptionsdomain.SubscriptionMutation{}, fmt.Errorf("apply Stripe subscription deletion: %w", err)
	}
	return subscriptionsdomain.SubscriptionMutation{WorkspaceID: row.WorkspaceID, Applied: row.Applied}, nil
}

func (repository *Repository) UpsertStripeInvoice(ctx context.Context, customerID string, invoice subscriptionsdomain.SubscriptionInvoice) error {
	if err := repository.configured(); err != nil {
		return err
	}
	seatsCount, err := subscriptionSeatCount(invoice.SeatsCount)
	if err != nil {
		return err
	}
	_, err = repository.queries.UpsertWorkspaceInvoice(ctx, subscriptionssql.UpsertWorkspaceInvoiceParams{
		WorkspaceID: invoice.WorkspaceID, StripeInvoiceID: invoice.StripeInvoiceID, AmountPaid: invoice.AmountPaid,
		InvoiceDate: invoice.InvoiceDate, Status: invoice.Status, SeatsCount: seatsCount,
		HostedURL: invoice.HostedURL, CustomerName: invoice.CustomerName, StripeCustomerID: customerID,
	})
	if !errors.Is(err, pgx.ErrNoRows) {
		if err != nil {
			return fmt.Errorf("upsert Stripe invoice: %w", err)
		}
		return nil
	}
	existingWorkspace, lookupErr := repository.queries.GetInvoiceWorkspaceByProviderID(ctx, subscriptionssql.GetInvoiceWorkspaceByProviderIDParams{StripeInvoiceID: invoice.StripeInvoiceID})
	switch {
	case lookupErr == nil && existingWorkspace != invoice.WorkspaceID:
		return subscriptionsdomain.ErrProviderIdentityConflict
	case lookupErr == nil:
		return errors.New("upsert Stripe invoice returned no row")
	case errors.Is(lookupErr, pgx.ErrNoRows):
		return subscriptionsdomain.ErrSubscriptionNotFound
	default:
		return fmt.Errorf("resolve Stripe invoice conflict: %w", lookupErr)
	}
}

func subscriptionSeatCount(value int) (int32, error) {
	if value < 0 {
		return 0, subscriptionsdomain.ErrInvalidSeatCount
	}
	seatCount, err := safecast.Int32(value)
	if err != nil {
		return 0, subscriptionsdomain.ErrInvalidSeatCount
	}
	return seatCount, nil
}

func validateCursor(cursor subscriptionsdomain.StripeEventCursor) error {
	if cursor.CreatedAt.IsZero() || cursor.Priority < 0 || strings.TrimSpace(cursor.EventID) == "" || len(cursor.EventID) > 255 {
		return subscriptionsdomain.ErrInvalidStripeEventIdentity
	}
	return nil
}

func (repository *Repository) hasCustomerBindingOutsideWorkspace(ctx context.Context, customerID string, workspaceID uuid.UUID) (bool, error) {
	conflict, err := repository.queries.HasCustomerBindingOutsideWorkspace(ctx, subscriptionssql.HasCustomerBindingOutsideWorkspaceParams{
		StripeCustomerID: customerID,
		WorkspaceID:      workspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("resolve Stripe customer binding: %w", err)
	}
	return conflict, nil
}

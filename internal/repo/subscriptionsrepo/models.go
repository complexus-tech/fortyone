package subscriptionsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/google/uuid"
)

type dbWorkspaceSubscription struct {
	WorkspaceID          uuid.UUID  `db:"workspace_id"`
	StripeCustomerID     string     `db:"stripe_customer_id"`
	StripeSubscriptionID *string    `db:"stripe_subscription_id"`
	SubscriptionStatus   *string    `db:"subscription_status"`
	SubscriptionTier     string     `db:"subscription_tier"` // Enum type
	SeatCount            int        `db:"seat_count"`
	TrialEndDate         *time.Time `db:"trial_end_date"`
	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
}

type dbSubscriptionInvoice struct {
	InvoiceID       int64     `db:"invoice_id"`
	WorkspaceID     uuid.UUID `db:"workspace_id"`
	StripeInvoiceID string    `db:"stripe_invoice_id"`
	AmountPaid      float64   `db:"amount_paid"`
	InvoiceDate     time.Time `db:"invoice_date"`
	Status          string    `db:"status"`
	SeatsCount      int       `db:"seats_count"`
	CreatedAt       time.Time `db:"created_at"`
}

func toCoreSubscription(db dbWorkspaceSubscription) subscriptions.CoreWorkspaceSubscription {
	var subscription subscriptions.CoreWorkspaceSubscription

	subscription.WorkspaceID = db.WorkspaceID
	subscription.StripeCustomerID = db.StripeCustomerID
	subscription.StripeSubscriptionID = db.StripeSubscriptionID
	subscription.SeatCount = db.SeatCount
	subscription.TrialEndDate = db.TrialEndDate
	subscription.CreatedAt = db.CreatedAt
	subscription.UpdatedAt = db.UpdatedAt

	// Convert status string to enum if present
	if db.SubscriptionStatus != nil {
		status := subscriptions.SubscriptionStatus(*db.SubscriptionStatus)
		subscription.SubscriptionStatus = &status
	}

	subscription.SubscriptionTier = subscriptions.SubscriptionTier(db.SubscriptionTier)

	return subscription
}

func toCoreInvoice(db dbSubscriptionInvoice) subscriptions.CoreSubscriptionInvoice {
	return subscriptions.CoreSubscriptionInvoice{
		InvoiceID:       db.InvoiceID,
		WorkspaceID:     db.WorkspaceID,
		StripeInvoiceID: db.StripeInvoiceID,
		AmountPaid:      db.AmountPaid,
		InvoiceDate:     db.InvoiceDate,
		Status:          db.Status,
		SeatsCount:      db.SeatsCount,
		CreatedAt:       db.CreatedAt,
	}
}

func toCoreInvoices(dbs []dbSubscriptionInvoice) []subscriptions.CoreSubscriptionInvoice {
	result := make([]subscriptions.CoreSubscriptionInvoice, len(dbs))
	for i, db := range dbs {
		result[i] = toCoreInvoice(db)
	}
	return result
}

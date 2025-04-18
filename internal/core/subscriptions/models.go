package subscriptions

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	StatusActive   SubscriptionStatus = "active"
	StatusInactive SubscriptionStatus = "inactive"
	StatusTrialing SubscriptionStatus = "trialing"
)

// SubscriptionTier represents the tier of a subscription
type SubscriptionTier string

const (
	TierFree       SubscriptionTier = "free"
	TierStarter    SubscriptionTier = "starter"
	TierPro        SubscriptionTier = "pro"
	TierEnterprise SubscriptionTier = "enterprise"
)

// CoreWorkspaceSubscription represents a workspace subscription
type CoreWorkspaceSubscription struct {
	WorkspaceID          uuid.UUID
	StripeCustomerID     string
	StripeSubscriptionID *string
	SubscriptionStatus   *SubscriptionStatus
	SubscriptionTier     SubscriptionTier
	SeatCount            int
	TrialEndDate         *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// CoreSubscriptionInvoice represents a subscription invoice
type CoreSubscriptionInvoice struct {
	InvoiceID       int64
	WorkspaceID     uuid.UUID
	StripeInvoiceID string
	AmountPaid      float64
	InvoiceDate     time.Time
	Status          string
	SeatsCount      int
	CreatedAt       time.Time
}

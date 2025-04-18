package subscriptions

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	StatusActive            SubscriptionStatus = "active"
	StatusInactive          SubscriptionStatus = "inactive"
	StatusTrialing          SubscriptionStatus = "trialing"
	StatusPastDue           SubscriptionStatus = "past_due"
	StatusIncomplete        SubscriptionStatus = "incomplete"
	StatusIncompleteExpired SubscriptionStatus = "incomplete_expired"
	StatusUnpaid            SubscriptionStatus = "unpaid"
	StatusCanceled          SubscriptionStatus = "canceled"
)

// SubscriptionTier represents the tier of a subscription
type SubscriptionTier string

const (
	TierFree       SubscriptionTier = "free"
	TierTeam       SubscriptionTier = "team"
	TierBusiness   SubscriptionTier = "business"
	TierEnterprise SubscriptionTier = "enterprise"
)

// CoreWorkspaceSubscription represents a workspace subscription
type CoreWorkspaceSubscription struct {
	WorkspaceID              uuid.UUID
	StripeCustomerID         string
	StripeSubscriptionID     *string
	StripeSubscriptionItemID *string
	SubscriptionStatus       *SubscriptionStatus
	SubscriptionTier         SubscriptionTier
	SeatCount                int
	TrialEndDate             *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
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

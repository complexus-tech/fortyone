package subscriptionsdomain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	StatusActive            SubscriptionStatus = "active"
	StatusIncomplete        SubscriptionStatus = "incomplete"
	StatusIncompleteExpired SubscriptionStatus = "incomplete_expired"
	StatusTrialing          SubscriptionStatus = "trialing"
	StatusPastDue           SubscriptionStatus = "past_due"
	StatusUnpaid            SubscriptionStatus = "unpaid"
	StatusCanceled          SubscriptionStatus = "canceled"
	StatusPaused            SubscriptionStatus = "paused"
)

type BillingInterval string

const (
	IntervalDay   BillingInterval = "day"
	IntervalWeek  BillingInterval = "week"
	IntervalMonth BillingInterval = "month"
	IntervalYear  BillingInterval = "year"
)

type SubscriptionTier string

const (
	TierFree       SubscriptionTier = "free"
	TierPro        SubscriptionTier = "pro"
	TierBusiness   SubscriptionTier = "business"
	TierEnterprise SubscriptionTier = "enterprise"
)

type WorkspaceSubscription struct {
	WorkspaceID              uuid.UUID
	StripeCustomerID         string
	StripeSubscriptionID     *string
	StripeSubscriptionItemID *string
	SubscriptionStatus       *SubscriptionStatus
	SubscriptionTier         SubscriptionTier
	SeatCount                int
	TrialEndDate             *time.Time
	BillingInterval          *BillingInterval
	BillingEndsAt            *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type SubscriptionInvoice struct {
	InvoiceID       int64
	WorkspaceID     uuid.UUID
	StripeInvoiceID string
	AmountPaid      float64
	InvoiceDate     time.Time
	Status          string
	SeatsCount      int
	CreatedAt       time.Time
	HostedURL       *string
	CustomerName    *string
}

// StripeEventCursor establishes deterministic arbitration for provider state
// changes. Priority resolves different event semantics created in the same
// Stripe timestamp second. EventID is only a stable final tie-breaker; Stripe
// event IDs do not encode chronological order.
type StripeEventCursor struct {
	CreatedAt time.Time
	Priority  int16
	EventID   string
}

const (
	StripeEventPriorityCreated  int16 = 10
	StripeEventPrioritySnapshot int16 = 20
	StripeEventPriorityDeleted  int16 = 30
)

type SubscriptionMutation struct {
	WorkspaceID uuid.UUID
	Applied     bool
}

type SubscriptionSnapshot struct {
	StripeCustomerID         string
	StripeSubscriptionID     string
	StripeSubscriptionItemID *string
	Status                   SubscriptionStatus
	Tier                     SubscriptionTier
	SeatCount                int
	TrialEnd                 *time.Time
	BillingInterval          *BillingInterval
	BillingEndsAt            *time.Time
}

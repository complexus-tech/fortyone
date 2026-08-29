package subscriptions

import subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"

type SubscriptionStatus = subscriptionsdomain.SubscriptionStatus

const (
	StatusActive            = subscriptionsdomain.StatusActive
	StatusIncomplete        = subscriptionsdomain.StatusIncomplete
	StatusIncompleteExpired = subscriptionsdomain.StatusIncompleteExpired
	StatusTrialing          = subscriptionsdomain.StatusTrialing
	StatusPastDue           = subscriptionsdomain.StatusPastDue
	StatusUnpaid            = subscriptionsdomain.StatusUnpaid
	StatusCanceled          = subscriptionsdomain.StatusCanceled
	StatusPaused            = subscriptionsdomain.StatusPaused
)

type BillingInterval = subscriptionsdomain.BillingInterval

const (
	IntervalDay   = subscriptionsdomain.IntervalDay
	IntervalWeek  = subscriptionsdomain.IntervalWeek
	IntervalMonth = subscriptionsdomain.IntervalMonth
	IntervalYear  = subscriptionsdomain.IntervalYear
)

type SubscriptionTier = subscriptionsdomain.SubscriptionTier

const (
	TierFree       = subscriptionsdomain.TierFree
	TierPro        = subscriptionsdomain.TierPro
	TierBusiness   = subscriptionsdomain.TierBusiness
	TierEnterprise = subscriptionsdomain.TierEnterprise
)

type CoreWorkspaceSubscription = subscriptionsdomain.WorkspaceSubscription
type CoreSubscriptionInvoice = subscriptionsdomain.SubscriptionInvoice

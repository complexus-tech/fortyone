package subscriptionsrepository

import (
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	subscriptionssql "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository/sqlc"
)

func toDomainSubscription(row subscriptionssql.GetSubscriptionByWorkspaceIDRow) subscriptionsdomain.WorkspaceSubscription {
	subscriptionID := row.StripeSubscriptionID
	result := subscriptionsdomain.WorkspaceSubscription{
		WorkspaceID: row.WorkspaceID, StripeCustomerID: row.StripeCustomerID,
		StripeSubscriptionID: &subscriptionID, StripeSubscriptionItemID: row.StripeSubscriptionItemID,
		SeatCount: int(row.SeatCount), TrialEndDate: row.TrialEndDate, BillingEndsAt: row.BillingEndsAt,
		CreatedAt: valueOrZero(row.CreatedAt), UpdatedAt: valueOrZero(row.UpdatedAt),
	}
	if row.SubscriptionStatus != nil {
		status := subscriptionsdomain.SubscriptionStatus(*row.SubscriptionStatus)
		result.SubscriptionStatus = &status
	}
	if row.SubscriptionTier != nil {
		result.SubscriptionTier = subscriptionsdomain.SubscriptionTier(*row.SubscriptionTier)
	}
	if row.BillingInterval != nil {
		interval := subscriptionsdomain.BillingInterval(*row.BillingInterval)
		result.BillingInterval = &interval
	}
	return result
}

func toDomainInvoice(row subscriptionssql.ListWorkspaceInvoicesRow) subscriptionsdomain.SubscriptionInvoice {
	return subscriptionsdomain.SubscriptionInvoice{
		InvoiceID: int64(row.InvoiceID), WorkspaceID: row.WorkspaceID, StripeInvoiceID: row.StripeInvoiceID,
		AmountPaid: row.AmountPaid, InvoiceDate: row.InvoiceDate, Status: row.Status,
		SeatsCount: int(row.SeatsCount), CreatedAt: row.CreatedAt,
		HostedURL: row.HostedURL, CustomerName: row.CustomerName,
	}
}

func valueOrZero(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func subscriptionParams(snapshot subscriptionsdomain.SubscriptionSnapshot) (*string, *subscriptionssql.SubscriptionTierEnum, *subscriptionssql.BillingIntervalEnum) {
	status := string(snapshot.Status)
	tier := subscriptionssql.SubscriptionTierEnum(snapshot.Tier)
	var interval *subscriptionssql.BillingIntervalEnum
	if snapshot.BillingInterval != nil {
		value := subscriptionssql.BillingIntervalEnum(*snapshot.BillingInterval)
		interval = &value
	}
	return &status, &tier, interval
}

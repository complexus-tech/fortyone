package subscriptionshttp

import (
	"time"

	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	"github.com/google/uuid"
)

// Request/response types
type AppCheckoutRequest struct {
	PriceLookupKey string `json:"priceLookupKey" validate:"required"`
	SuccessURL     string `json:"successUrl" validate:"required"`
	CancelURL      string `json:"cancelUrl" validate:"required"`
}

type AppCheckoutResponse struct {
	URL string `json:"url"`
}

type AppCustomerPortalRequest struct {
	ReturnURL string `json:"returnUrl" validate:"required"`
}

type AppCustomerPortalResponse struct {
	URL string `json:"url"`
}

type AppChangeSubscriptionPlanRequest struct {
	NewLookupKey string `json:"newLookupKey" validate:"required"`
}

// App representation of a subscription
type AppSubscription struct {
	WorkspaceID          uuid.UUID  `json:"workspaceId"`
	StripeCustomerID     string     `json:"stripeCustomerId"`
	StripeSubscriptionID *string    `json:"stripeSubscriptionId"`
	Status               *string    `json:"status"`
	Tier                 string     `json:"tier"`
	SeatCount            int        `json:"seatCount"`
	TrialEndDate         *time.Time `json:"trialEndDate"`
	BillingInterval      *string    `json:"billingInterval"`
	BillingEndsAt        *time.Time `json:"billingEndsAt"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

// App representation of an invoice
type AppInvoice struct {
	InvoiceID       int64     `json:"invoiceId"`
	WorkspaceID     uuid.UUID `json:"workspaceId"`
	StripeInvoiceID string    `json:"stripeInvoiceId"`
	AmountPaid      float64   `json:"amountPaid"`
	InvoiceDate     time.Time `json:"invoiceDate"`
	Status          string    `json:"status"`
	SeatsCount      int       `json:"seatsCount"`
	CreatedAt       time.Time `json:"createdAt"`
	HostedURL       *string   `json:"hostedUrl"`
	CustomerName    *string   `json:"customerName"`
}

// Conversion functions
func toAppSubscription(core subscriptions.CoreWorkspaceSubscription) AppSubscription {
	appSub := AppSubscription{
		WorkspaceID:          core.WorkspaceID,
		StripeCustomerID:     core.StripeCustomerID,
		StripeSubscriptionID: core.StripeSubscriptionID,
		Tier:                 string(core.SubscriptionTier),
		SeatCount:            core.SeatCount,
		TrialEndDate:         core.TrialEndDate,
		CreatedAt:            core.CreatedAt,
		UpdatedAt:            core.UpdatedAt,
		BillingEndsAt:        core.BillingEndsAt,
	}

	if core.SubscriptionStatus != nil {
		statusStr := string(*core.SubscriptionStatus)
		appSub.Status = &statusStr
	}

	if core.BillingInterval != nil {
		intervalStr := string(*core.BillingInterval)
		appSub.BillingInterval = &intervalStr
	}

	return appSub
}

func toAppInvoices(coreInvoices []subscriptions.CoreSubscriptionInvoice) []AppInvoice {
	if len(coreInvoices) == 0 {
		return []AppInvoice{}
	}

	result := make([]AppInvoice, len(coreInvoices))
	for i, invoice := range coreInvoices {
		result[i] = AppInvoice{
			InvoiceID:       invoice.InvoiceID,
			WorkspaceID:     invoice.WorkspaceID,
			StripeInvoiceID: invoice.StripeInvoiceID,
			AmountPaid:      invoice.AmountPaid,
			InvoiceDate:     invoice.InvoiceDate,
			Status:          invoice.Status,
			SeatsCount:      invoice.SeatsCount,
			CreatedAt:       invoice.CreatedAt,
			HostedURL:       invoice.HostedURL,
			CustomerName:    invoice.CustomerName,
		}
	}
	return result
}

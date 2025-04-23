package subscriptionsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/google/uuid"
)

// Request/response types
type AppCheckoutRequest struct {
	PriceLookupKey string `json:"priceLookupKey"`
}

type AppCheckoutResponse struct {
	URL string `json:"url"`
}

type AppCustomerPortalRequest struct {
	ReturnURL string `json:"returnUrl"`
}

type AppCustomerPortalResponse struct {
	URL string `json:"url"`
}

// App representation of a subscription
type AppSubscription struct {
	WorkspaceID          uuid.UUID  `json:"workspaceId"`
	StripeCustomerID     string     `json:"stripeCustomerId"`
	StripeSubscriptionID *string    `json:"stripeSubscriptionId,omitempty"`
	Status               *string    `json:"status,omitempty"`
	Tier                 string     `json:"tier"`
	SeatCount            int        `json:"seatCount"`
	TrialEndDate         *time.Time `json:"trialEndDate,omitempty"`
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
}

// Conversion functions
func toAppSubscription(core subscriptions.CoreWorkspaceSubscription) AppSubscription {
	var status *string
	if core.SubscriptionStatus != nil {
		s := string(*core.SubscriptionStatus)
		status = &s
	}

	return AppSubscription{
		WorkspaceID:          core.WorkspaceID,
		StripeCustomerID:     core.StripeCustomerID,
		StripeSubscriptionID: core.StripeSubscriptionID,
		Status:               status,
		Tier:                 string(core.SubscriptionTier),
		SeatCount:            core.SeatCount,
		TrialEndDate:         core.TrialEndDate,
		CreatedAt:            core.CreatedAt,
		UpdatedAt:            core.UpdatedAt,
	}
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
		}
	}
	return result
}

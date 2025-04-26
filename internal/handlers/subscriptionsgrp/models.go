package subscriptionsgrp

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/go-playground/validator/v10"
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
	HostedURL       *string   `json:"hostedUrl"`
	CustomerName    *string   `json:"customerName"`
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
			HostedURL:       invoice.HostedURL,
			CustomerName:    invoice.CustomerName,
		}
	}
	return result
}

func (a AppCheckoutRequest) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(a)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range validationErrors {
				fieldName := getJSONTagName(reflect.TypeOf(a), e.Field())
				switch e.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("%s is required", fieldName))
				case "oneof":
					options := strings.Split(e.Param(), " ")
					formattedOptions := formatOptions(options)
					errorMessages = append(errorMessages, fmt.Sprintf("%s should be one of: %s", fieldName, formattedOptions))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation: %s", fieldName, e.Tag()))
				}
			}
			return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
		}
	}
	return err
}

func formatOptions(options []string) string {
	if len(options) == 0 {
		return ""
	}
	if len(options) == 1 {
		return options[0]
	}
	return fmt.Sprintf("%s or %s", strings.Join(options[:len(options)-1], ", "), options[len(options)-1])
}

func getJSONTagName(t reflect.Type, fieldName string) string {
	field, found := t.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return fieldName // Return original field name if no JSON tag
	}

	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return fieldName // Return original field name if JSON tag is "-"
	}

	return parts[0] // Return the JSON tag name
}

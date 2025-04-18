package subscriptionsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/complexus-tech/projects-api/internal/repo/subscriptionsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"github.com/stripe/stripe-go/v82/client"
)

type Config struct {
	DB                 *sqlx.DB
	Log                *logger.Logger
	SecretKey          string
	StripeClient       *client.API
	CheckoutSuccessURL string
	CheckoutCancelURL  string
	WebhookSecret      string
}

func Routes(cfg Config, app *web.App) {
	// Initialize repository and service
	subsRepo := subscriptionsrepo.New(cfg.Log, cfg.DB)
	subsService := subscriptions.New(
		cfg.Log,
		subsRepo,
		cfg.StripeClient,
		cfg.CheckoutSuccessURL,
		cfg.CheckoutCancelURL,
		cfg.WebhookSecret,
	)
	h := New(subsService, cfg.Log)

	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	app.Post("/workspaces/{workspaceId}/subscriptions/checkout", h.CreateCheckoutSession, auth)
	app.Post("/workspaces/{workspaceId}/subscriptions/portal", h.CreateCustomerPortal, auth)
	app.Get("/workspaces/{workspaceId}/subscription", h.GetSubscription, auth)
	app.Get("/workspaces/{workspaceId}/invoices", h.GetInvoices, auth)
	app.Post("/workspaces/{workspaceId}/subscriptions/add-seat", h.AddSeat, auth)

	// Public webhook endpoint
	app.Post("/webhooks/stripe", h.HandleWebhook)
}

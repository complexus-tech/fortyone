package subscriptionsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/repo/objectivestatusrepo"
	"github.com/complexus-tech/projects-api/internal/repo/statesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/storiesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/subscriptionsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/teamsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/usersrepo"
	"github.com/complexus-tech/projects-api/internal/repo/workspacesrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"github.com/stripe/stripe-go/v82/client"
)

type Config struct {
	DB            *sqlx.DB
	Log           *logger.Logger
	SecretKey     string
	StripeClient  *client.API
	WebhookSecret string
	Publisher     *publisher.Publisher
}

func Routes(cfg Config, app *web.App) {
	// Initialize repository and service
	subsRepo := subscriptionsrepo.New(cfg.Log, cfg.DB)
	subsService := subscriptions.New(
		cfg.Log,
		subsRepo,
		cfg.StripeClient,
		cfg.WebhookSecret,
	)

	teamsService := teams.New(cfg.Log, teamsrepo.New(cfg.Log, cfg.DB))
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB), cfg.Publisher)
	statusesService := states.New(cfg.Log, statesrepo.New(cfg.Log, cfg.DB))
	objectivestatusService := objectivestatus.New(cfg.Log, objectivestatusrepo.New(cfg.Log, cfg.DB))
	usersService := users.New(cfg.Log, usersrepo.New(cfg.Log, cfg.DB))
	workspacesService := workspaces.New(cfg.Log, workspacesrepo.New(cfg.Log, cfg.DB), cfg.DB, teamsService, storiesService, statusesService, usersService, objectivestatusService, subsService)

	h := New(subsService, usersService, workspacesService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Post("/workspaces/{workspaceId}/subscriptions/checkout", h.CreateCheckoutSession, auth)
	app.Post("/workspaces/{workspaceId}/subscriptions/portal", h.CreateCustomerPortal, auth)
	app.Get("/workspaces/{workspaceId}/subscription", h.GetSubscription, auth)
	app.Get("/workspaces/{workspaceId}/invoices", h.GetInvoices, auth)
	app.Post("/workspaces/{workspaceId}/subscriptions/add-seat", h.AddSeat, auth)
	app.Post("/workspaces/{workspaceId}/subscriptions/change-plan", h.ChangeSubscriptionPlan, auth)

	// Public webhook endpoint
	app.Post("/webhooks/stripe", h.HandleWebhook)
}

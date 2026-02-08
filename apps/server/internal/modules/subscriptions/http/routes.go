package subscriptionshttp

import (
	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
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
	TasksService  *tasks.Service
	SystemUserID  uuid.UUID
	Cache         *cache.Service
	Subscriptions *subscriptions.Service
	Users         *users.Service
	Workspaces    *workspaces.Service
}

func Routes(cfg Config, app *web.App) {
	subsService := cfg.Subscriptions
	usersService := cfg.Users
	workspacesService := cfg.Workspaces

	h := New(subsService, usersService, workspacesService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Post("/workspaces/{workspaceSlug}/subscriptions/checkout", h.CreateCheckoutSession, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/portal", h.CreateCustomerPortal, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/subscription", h.GetSubscription, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/invoices", h.GetInvoices, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/add-seat", h.AddSeat, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/change-plan", h.ChangeSubscriptionPlan, auth, workspace)

	// Public webhook endpoint
	app.Post("/webhooks/stripe", h.HandleWebhook)
}

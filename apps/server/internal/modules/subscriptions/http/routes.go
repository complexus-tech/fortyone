package subscriptionshttp

import (
	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Subscriptions     *subscriptions.Service
	Users             *users.Service
	Workspaces        *workspaces.Service
}

func Routes(cfg Config, app *web.App) {
	subsService := cfg.Subscriptions
	usersService := cfg.Users
	workspacesService := cfg.Workspaces

	h := New(subsService, usersService, workspacesService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	app.Post("/workspaces/{workspaceSlug}/subscriptions/checkout", h.CreateCheckoutSession, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/portal", h.CreateCustomerPortal, auth, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/subscription", h.GetSubscription, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/invoices", h.GetInvoices, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/add-seat", h.AddSeat, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/change-plan", h.ChangeSubscriptionPlan, auth, workspace, adminOnly)

	// Public webhook endpoint
	app.Post("/webhooks/stripe", h.HandleWebhook)
}

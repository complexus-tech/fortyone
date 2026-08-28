package invitationshttp

import (
	"time"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/stripe/stripe-go/v82/client"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	Publisher         *publisher.Publisher
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	StripeClient      *client.API
	StripeSecret      string
	TasksService      *tasks.Service
	Invitations       *invitations.Service
	UsersService      *users.Service
}

func Routes(cfg Config, app *web.App) {
	usersService := cfg.UsersService
	invitationsService := cfg.Invitations

	h := New(invitationsService, usersService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	createRateLimit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{
		Scope:  "workspace-invitation-create",
		Limit:  20,
		Window: time.Hour,
	})
	acceptRateLimit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{
		Scope:  "workspace-invitation-accept",
		Limit:  60,
		Window: time.Hour,
	})

	app.Post("/workspaces/{workspaceSlug}/invitations", h.CreateBulkInvitations, auth, createRateLimit, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/invitations", h.ListInvitations, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/invitations/{id}", h.RevokeInvitation, auth, workspace, adminOnly)
	app.Get("/invitations/{token}", h.GetInvitation)
	app.Get("/users/me/invitations", h.ListUserInvitations, auth)
	app.Post("/invitations/{token}/accept", h.AcceptInvitation, auth, acceptRateLimit)
}

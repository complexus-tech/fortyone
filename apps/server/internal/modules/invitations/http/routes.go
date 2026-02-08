package invitationshttp

import (
	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
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
	DB           *sqlx.DB
	Log          *logger.Logger
	SecretKey    string
	Publisher    *publisher.Publisher
	Cache        *cache.Service
	StripeClient *client.API
	StripeSecret string
	TasksService *tasks.Service
	SystemUserID uuid.UUID
	Invitations  *invitations.Service
	UsersService *users.Service
}

func Routes(cfg Config, app *web.App) {
	usersService := cfg.UsersService
	invitationsService := cfg.Invitations

	h := New(invitationsService, usersService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Post("/workspaces/{workspaceSlug}/invitations", h.CreateBulkInvitations, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/invitations", h.ListInvitations, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/invitations/{id}", h.RevokeInvitation, auth, workspace)
	app.Get("/invitations/{token}", h.GetInvitation)
	app.Get("/users/me/invitations", h.ListUserInvitations, auth)
	app.Post("/invitations/{token}/accept", h.AcceptInvitation, auth)
}

package invitationshttp

import (
	invitationsrepository "github.com/complexus-tech/projects-api/internal/modules/invitations/repository"
	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	subscriptionsrepository "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository"
	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
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
	DB           *sqlx.DB
	Log          *logger.Logger
	SecretKey    string
	Publisher    *publisher.Publisher
	Cache        *cache.Service
	StripeClient *client.API
	StripeSecret string
	TasksService *tasks.Service
	SystemUserID uuid.UUID
}

func Routes(cfg Config, app *web.App) {
	repo := invitationsrepository.New(cfg.Log, cfg.DB)
	usersService := users.New(cfg.Log, usersrepository.New(cfg.Log, cfg.DB), cfg.TasksService)
	subscriptionsService := subscriptions.New(cfg.Log, subscriptionsrepository.New(cfg.Log, cfg.DB), cfg.StripeClient, cfg.StripeSecret, cfg.TasksService)
	workspacesService := workspaces.New(cfg.Log, workspacesrepository.New(cfg.Log, cfg.DB), cfg.DB, nil, nil, nil, usersService, nil, subscriptionsService, nil, cfg.Cache, cfg.SystemUserID, cfg.Publisher, cfg.TasksService)
	teamsService := teams.New(cfg.Log, teamsrepository.New(cfg.Log, cfg.DB))
	invitationsService := invitations.New(repo, cfg.Log, cfg.Publisher, usersService, workspacesService, teamsService)
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

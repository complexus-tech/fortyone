package workspaceshttp

import (
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	mentionsrepository "github.com/complexus-tech/projects-api/internal/modules/mentions/repository"
	objectivestatusrepository "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/repository"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	statesrepository "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
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
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stripe/stripe-go/v82/client"
)

type Config struct {
	DB             *sqlx.DB
	Log            *logger.Logger
	SecretKey      string
	Publisher      *publisher.Publisher
	Cache          *cache.Service
	StripeClient   *client.API
	WebhookSecret  string
	TasksService   *tasks.Service
	SystemUserID   uuid.UUID
	StorageConfig  storage.Config
	StorageService storage.StorageService
}

func Routes(cfg Config, app *web.App) {
	teamsService := teams.New(cfg.Log, teamsrepository.New(cfg.Log, cfg.DB))
	mentionsRepo := mentionsrepository.New(cfg.Log, cfg.DB)
	storiesService := stories.New(cfg.Log, storiesrepository.New(cfg.Log, cfg.DB), mentionsRepo, cfg.Publisher)
	statusesService := states.New(cfg.Log, statesrepository.New(cfg.Log, cfg.DB))
	objectivestatusService := objectivestatus.New(cfg.Log, objectivestatusrepository.New(cfg.Log, cfg.DB))
	usersService := users.New(cfg.Log, usersrepository.New(cfg.Log, cfg.DB), cfg.TasksService)
	subscriptionsService := subscriptions.New(cfg.Log, subscriptionsrepository.New(cfg.Log, cfg.DB), cfg.StripeClient, cfg.WebhookSecret, cfg.TasksService)

	// Create attachments service for workspace logos
	attachmentsRepo := attachmentsrepository.New(cfg.Log, cfg.DB)
	attachmentsService := attachments.New(cfg.Log, attachmentsRepo, cfg.StorageService, cfg.StorageConfig)

	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	workspacesService := workspaces.New(cfg.Log, workspacesrepository.New(cfg.Log, cfg.DB), cfg.DB, teamsService, storiesService, statusesService, usersService, objectivestatusService, subscriptionsService, attachmentsService, cfg.Cache, cfg.SystemUserID, cfg.Publisher, cfg.TasksService)

	h := New(workspacesService, teamsService,
		storiesService, statusesService, usersService, objectivestatusService, subscriptionsService,
		cfg.Cache, cfg.Log, cfg.SecretKey, attachmentsService)

	app.Get("/workspaces/{workspaceSlug}", h.Get, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}", h.Update, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}", h.Delete, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/restore", h.Restore, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/members", h.AddMember, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/members/{userId}/role", h.UpdateMemberRole, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/members/{userId}", h.RemoveMember, auth, workspace, adminOnly)
	app.Post("/workspaces", h.Create, auth)
	app.Get("/workspaces", h.List, auth)
	app.Get("/workspaces/check-availability", h.CheckSlugAvailability)
	app.Get("/workspaces/{workspaceSlug}/settings", h.GetWorkspaceSettings, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/settings", h.UpdateWorkspaceSettings, auth, workspace, adminOnly)

	// Workspace logo endpoints
	app.Post("/workspaces/{workspaceSlug}/logo", h.UploadWorkspaceLogo, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/logo", h.DeleteWorkspaceLogo, auth, workspace, adminOnly)
}

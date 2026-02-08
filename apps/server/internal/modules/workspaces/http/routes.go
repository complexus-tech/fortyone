package workspaceshttp

import (
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
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
	DB              *sqlx.DB
	Log             *logger.Logger
	SecretKey       string
	Publisher       *publisher.Publisher
	Cache           *cache.Service
	StripeClient    *client.API
	WebhookSecret   string
	TasksService    *tasks.Service
	SystemUserID    uuid.UUID
	StorageConfig   storage.Config
	StorageService  storage.StorageService
	Workspaces      *workspaces.Service
	Teams           *teams.Service
	Stories         *stories.Service
	Statuses        *states.Service
	Users           *users.Service
	ObjectiveStatus *objectivestatus.Service
	Subscriptions   *subscriptions.Service
	Attachments     *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	teamsService := cfg.Teams
	storiesService := cfg.Stories
	statusesService := cfg.Statuses
	objectivestatusService := cfg.ObjectiveStatus
	usersService := cfg.Users
	subscriptionsService := cfg.Subscriptions
	attachmentsService := cfg.Attachments

	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	h := New(cfg.Workspaces, teamsService,
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

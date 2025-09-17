package workspacesgrp

import (
	"context"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/repo/attachmentsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/mentionsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/objectivestatusrepo"
	"github.com/complexus-tech/projects-api/internal/repo/statesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/storiesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/subscriptionsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/teamsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/usersrepo"
	"github.com/complexus-tech/projects-api/internal/repo/workspacesrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/azure"
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
	Publisher     *publisher.Publisher
	Cache         *cache.Service
	StripeClient  *client.API
	WebhookSecret string
	TasksService  *tasks.Service
	SystemUserID  uuid.UUID
	AzureConfig   azure.Config
}

func Routes(cfg Config, app *web.App) {
	teamsService := teams.New(cfg.Log, teamsrepo.New(cfg.Log, cfg.DB))
	mentionsRepo := mentionsrepo.New(cfg.Log, cfg.DB)
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB), mentionsRepo, cfg.Publisher)
	statusesService := states.New(cfg.Log, statesrepo.New(cfg.Log, cfg.DB))
	objectivestatusService := objectivestatus.New(cfg.Log, objectivestatusrepo.New(cfg.Log, cfg.DB))
	usersService := users.New(cfg.Log, usersrepo.New(cfg.Log, cfg.DB), cfg.TasksService)
	subscriptionsService := subscriptions.New(cfg.Log, subscriptionsrepo.New(cfg.Log, cfg.DB), cfg.StripeClient, cfg.WebhookSecret)

	// Create attachments service for workspace logos
	attachmentsRepo := attachmentsrepo.New(cfg.Log, cfg.DB)
	azureBlobService, err := azure.New(cfg.AzureConfig, cfg.Log)
	if err != nil {
		cfg.Log.Error(context.Background(), "failed to initialize Azure blob service", "error", err)
		return
	}
	attachmentsService := attachments.New(cfg.Log, attachmentsRepo, azureBlobService, cfg.AzureConfig)

	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	workspacesService := workspaces.New(cfg.Log, workspacesrepo.New(cfg.Log, cfg.DB), cfg.DB, teamsService, storiesService, statusesService, usersService, objectivestatusService, subscriptionsService, attachmentsService, cfg.Cache, cfg.SystemUserID)

	h := New(workspacesService, teamsService,
		storiesService, statusesService, usersService, objectivestatusService, subscriptionsService,
		cfg.Cache, cfg.Log, cfg.SecretKey, attachmentsService)

	app.Get("/workspaces/{workspaceSlug}", h.Get, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}", h.Update, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}", h.Delete, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/restore", h.Delete, auth, workspace, adminOnly)
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

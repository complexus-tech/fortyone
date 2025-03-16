package workspacesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/repo/objectivestatusrepo"
	"github.com/complexus-tech/projects-api/internal/repo/statesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/storiesrepo"
	"github.com/complexus-tech/projects-api/internal/repo/teamsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/usersrepo"
	"github.com/complexus-tech/projects-api/internal/repo/workspacesrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
	Publisher *publisher.Publisher
}

func Routes(cfg Config, app *web.App) {

	teamsService := teams.New(cfg.Log, teamsrepo.New(cfg.Log, cfg.DB))
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB), cfg.Publisher)
	statusesService := states.New(cfg.Log, statesrepo.New(cfg.Log, cfg.DB))
	objectivestatusService := objectivestatus.New(cfg.Log, objectivestatusrepo.New(cfg.Log, cfg.DB))
	usersService := users.New(cfg.Log, usersrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	workspacesService := workspaces.New(cfg.Log, workspacesrepo.New(cfg.Log, cfg.DB), cfg.DB, teamsService, storiesService, statusesService, usersService, objectivestatusService)

	h := New(workspacesService, teamsService,
		storiesService, statusesService, usersService, objectivestatusService,
		cfg.SecretKey)

	app.Get("/workspaces/{workspaceId}", h.Get, auth)
	app.Put("/workspaces/{workspaceId}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}", h.Delete, auth)
	app.Post("/workspaces/{workspaceId}/members", h.AddMember, auth)
	app.Put("/workspaces/{workspaceId}/members/{userId}/role", h.UpdateMemberRole, auth)
	app.Delete("/workspaces/{workspaceId}/members/{userId}", h.RemoveMember, auth)
	app.Post("/workspaces", h.Create, auth)
	app.Get("/workspaces", h.List, auth)
	app.Get("/workspaces/check-availability", h.CheckSlugAvailability)

	// Workspace settings endpoints (replaces terminology endpoints)
	app.Get("/workspaces/{workspaceId}/settings", h.GetWorkspaceSettings, auth)
	app.Put("/workspaces/{workspaceId}/settings", h.UpdateWorkspaceSettings, auth)
}

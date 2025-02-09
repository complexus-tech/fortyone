package workspacesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/objectivestatus/objectivestatusrepo"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/states/statesrepo"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/stories/storiesrepo"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/teams/teamsrepo"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/core/workspaces/workspacesrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
}

func Routes(cfg Config, app *web.App) {

	workspacesService := workspaces.New(cfg.Log, workspacesrepo.New(cfg.Log, cfg.DB))
	teamsService := teams.New(cfg.Log, teamsrepo.New(cfg.Log, cfg.DB))
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB))
	statusesService := states.New(cfg.Log, statesrepo.New(cfg.Log, cfg.DB))
	objectivestatusService := objectivestatus.New(cfg.Log, objectivestatusrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(workspacesService, teamsService,
		storiesService, statusesService, objectivestatusService,
		cfg.SecretKey)

	app.Get("/workspaces/{id}", h.Get, auth)
	app.Put("/workspaces/{id}", h.Update, auth)
	app.Delete("/workspaces/{id}", h.Delete, auth)
	app.Post("/workspaces/{id}/members", h.AddMember, auth)
	app.Delete("/workspaces/{id}/members/{userId}", h.RemoveMember, auth)
	app.Post("/workspaces", h.Create, auth)
	app.Get("/workspaces", h.List, auth)
}

package teamsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/repo/teamsrepo"
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
	teamsService := teams.New(cfg.Log, teamsrepo.New(cfg.Log, cfg.DB))
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(teamsService)

	app.Get("/workspaces/{workspaceId}/teams", h.List, auth)
	app.Get("/workspaces/{workspaceId}/teams/public", h.ListPublicTeams, auth)
	app.Post("/workspaces/{workspaceId}/teams", h.Create, auth)
	app.Put("/workspaces/{workspaceId}/teams/{id}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/teams/{id}", h.Delete, auth)
	app.Post("/workspaces/{workspaceId}/teams/{id}/members", h.AddMember, auth)
	app.Delete("/workspaces/{workspaceId}/teams/{id}/members/{userId}", h.RemoveMember, auth)
}

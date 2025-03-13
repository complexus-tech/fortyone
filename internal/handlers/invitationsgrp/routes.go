package invitationsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/invitations"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/repo/invitationsrepo"
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
	repo := invitationsrepo.New(cfg.Log, cfg.DB)
	usersService := users.New(cfg.Log, usersrepo.New(cfg.Log, cfg.DB))
	workspacesService := workspaces.New(cfg.Log, workspacesrepo.New(cfg.Log, cfg.DB), cfg.DB, nil, nil, nil, usersService, nil)
	teamsService := teams.New(cfg.Log, teamsrepo.New(cfg.Log, cfg.DB))
	invitationsService := invitations.New(repo, cfg.Log, cfg.Publisher, usersService, workspacesService, teamsService)
	h := New(invitationsService, usersService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Post("/workspaces/{workspaceId}/invitations", h.CreateBulkInvitations, auth)
	app.Get("/workspaces/{workspaceId}/invitations", h.ListInvitations, auth)
	app.Delete("/workspaces/{workspaceId}/invitations/{id}", h.RevokeInvitation, auth)
	app.Get("/invitations/{token}", h.GetInvitation)
	app.Get("/users/me/invitations", h.ListUserInvitations, auth)
	app.Post("/invitations/{token}/accept", h.AcceptInvitation, auth)
}

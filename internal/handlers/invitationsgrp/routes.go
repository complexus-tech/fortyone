package invitationsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/invitations"
	"github.com/complexus-tech/projects-api/internal/core/invitations/invitationsrepo"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/users/usersrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
	Publisher *events.Publisher
}

func Routes(cfg Config, app *web.App) {
	repo := invitationsrepo.New(cfg.Log, cfg.DB)
	usersService := users.New(cfg.Log, usersrepo.New(cfg.Log, cfg.DB))
	invitationsService := invitations.New(repo, cfg.Log, cfg.Publisher, usersService)
	h := New(invitationsService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Post("/workspaces/{workspaceId}/invitations", h.CreateBulkInvitations, auth)
	app.Get("/workspaces/{workspaceId}/invitations", h.ListInvitations, auth)
	app.Delete("/workspaces/{workspaceId}/invitations/{id}", h.RevokeInvitation, auth)
	app.Get("/invitations/{token}", h.GetInvitation)
}

package invitationsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/invitations"
	"github.com/complexus-tech/projects-api/internal/core/invitations/invitationsrepo"
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
	invitationsService := invitations.New(cfg.Log, invitationsrepo.New(cfg.Log, cfg.DB))
	h := New(invitationsService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	// Email invitations
	app.Post("/workspaces/{workspaceId}/invitations", h.CreateInvitation, auth)
	app.Get("/workspaces/{workspaceId}/invitations", h.ListInvitations, auth)
	app.Delete("/workspaces/{workspaceId}/invitations/{id}", h.RevokeInvitation, auth)
	app.Get("/invitations/{token}", h.GetInvitation)

	// Invitation links
	app.Post("/workspaces/{workspaceId}/invitation-links", h.CreateInvitationLink, auth)
	app.Get("/workspaces/{workspaceId}/invitation-links", h.ListInvitationLinks, auth)
	app.Delete("/workspaces/{workspaceId}/invitation-links/{id}", h.RevokeInvitationLink, auth)
	app.Get("/invitation-links/{token}", h.GetInvitationLink)
	app.Post("/invitation-links/{id}/use", h.UseInvitationLink, auth)
}

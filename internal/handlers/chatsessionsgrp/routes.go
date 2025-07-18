package chatsessionsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/chatsessions"
	"github.com/complexus-tech/projects-api/internal/repo/chatsessionsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
	Validate  *validator.Validate
}

func Routes(cfg Config, app *web.App) {
	chatsessionsRepo := chatsessionsrepo.New(cfg.Log, cfg.DB)
	chatsessionsService := chatsessions.New(cfg.Log, chatsessionsRepo)

	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)

	h := New(chatsessionsService, cfg.Log)

	// Chat sessions
	app.Post("/workspaces/{workspaceId}/chat-sessions", h.CreateSession, auth)
	app.Get("/workspaces/{workspaceId}/chat-sessions", h.ListSessions, auth, gzip)
	app.Get("/workspaces/{workspaceId}/chat-sessions/{sessionId}", h.GetSession, auth, gzip)
	app.Put("/workspaces/{workspaceId}/chat-sessions/{sessionId}", h.UpdateSession, auth)
	app.Delete("/workspaces/{workspaceId}/chat-sessions/{sessionId}", h.DeleteSession, auth)

	// Chat messages
	app.Post("/workspaces/{workspaceId}/chat-sessions/{sessionId}/messages", h.SaveMessages, auth)
	app.Get("/workspaces/{workspaceId}/chat-sessions/{sessionId}/messages", h.GetMessages, auth, gzip)
}

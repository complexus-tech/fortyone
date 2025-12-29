package chatsessionsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/chatsessions"
	"github.com/complexus-tech/projects-api/internal/repo/chatsessionsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
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
	Cache     *cache.Service
}

func Routes(cfg Config, app *web.App) {
	chatsessionsRepo := chatsessionsrepo.New(cfg.Log, cfg.DB)
	chatsessionsService := chatsessions.New(cfg.Log, chatsessionsRepo)

	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(chatsessionsService, cfg.Log)

	// Chat sessions
	app.Post("/workspaces/{workspaceSlug}/chat-sessions", h.CreateSession, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions", h.ListSessions, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}", h.GetSession, auth, workspace, gzip)
	app.Put("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}", h.UpdateSession, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}", h.DeleteSession, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions/messages/count", h.GetUserMessageCount, auth, workspace)

	// Chat messages
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/messages", h.SaveMessages, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/messages", h.GetMessages, auth, workspace, gzip)
}

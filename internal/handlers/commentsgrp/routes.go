package commentsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/repo/commentsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/mentionsrepo"
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
	mentionsRepo := mentionsrepo.New(cfg.Log, cfg.DB)
	commentsService := comments.New(cfg.Log, commentsrepo.New(cfg.Log, cfg.DB), mentionsRepo)
	h := New(cfg.Log, commentsService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Put("/workspaces/{workspaceId}/comments/{id}", h.UpdateComment, auth)
	app.Delete("/workspaces/{workspaceId}/comments/{id}", h.DeleteComment, auth)
}

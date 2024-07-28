package storiesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/stories/storiesrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB  *sqlx.DB
	Log *logger.Logger
}

func Routes(cfg Config, app *web.App) {
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB))

	h := New(storiesService, cfg.Log)

	app.Get("/workspaces/{workspaceId}/stories", h.List)
	app.Get("/workspaces/{workspaceId}/stories/{id}", h.Get)
	app.Get("/workspaces/{workspaceId}/my-stories", h.MyStories)
	app.Post("/workspaces/{workspaceId}/stories", h.Create)
	app.Post("/workspaces/{workspaceId}/stories/{id}/restore", h.Restore)
	app.Post("/workspaces/{workspaceId}/stories/restore", h.BulkRestore)
	app.Patch("/workspaces/{workspaceId}/stories/{id}", h.Update)
	app.Delete("/workspaces/{workspaceId}/stories/{id}", h.Delete)
	app.Delete("/workspaces/{workspaceId}/stories", h.BulkDelete)
}

package storiesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/stories/storiesrepo"
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
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB))
	h := New(storiesService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Get("/workspaces/{workspaceId}/stories", h.List, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}", h.Get, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/activities", h.GetActivities, auth)
	app.Post("/workspaces/{workspaceId}/stories/{id}/comments", h.CreateComment, auth)
	app.Put("/workspaces/{workspaceId}/stories/{id}/labels", h.UpdateLabels, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/links", h.GetStoryLinks, auth)
	app.Get("/workspaces/{workspaceId}/my-stories", h.MyStories, auth)
	app.Post("/workspaces/{workspaceId}/stories", h.Create, auth)
	app.Post("/workspaces/{workspaceId}/stories/{id}/restore", h.Restore, auth)
	app.Post("/workspaces/{workspaceId}/stories/restore", h.BulkRestore, auth)
	app.Patch("/workspaces/{workspaceId}/stories/{id}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/stories/{id}", h.Delete, auth)
	app.Delete("/workspaces/{workspaceId}/stories", h.BulkDelete, auth)
}

package storiesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/comments/commentsrepo"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/core/links/linksrepo"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/stories/storiesrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/email"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB           *sqlx.DB
	Log          *logger.Logger
	SecretKey    string
	Publisher    *events.Publisher
	EmailService email.Service
}

func Routes(cfg Config, app *web.App) {
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB), cfg.Publisher, cfg.EmailService)
	commentsService := comments.New(cfg.Log, commentsrepo.New(cfg.Log, cfg.DB))
	linksService := links.New(cfg.Log, linksrepo.New(cfg.Log, cfg.DB))

	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(storiesService, commentsService, linksService, cfg.Log)

	// Stories
	app.Get("/workspaces/{workspaceId}/stories", h.List, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}", h.Get, auth)
	app.Post("/workspaces/{workspaceId}/stories", h.Create, auth)
	app.Put("/workspaces/{workspaceId}/stories/{id}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/stories/{id}", h.Delete, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/activities", h.GetActivities, auth)

	// Comments
	app.Post("/workspaces/{workspaceId}/stories/{id}/comments", h.CreateComment, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/comments", h.GetComments, auth)
	app.Put("/workspaces/{workspaceId}/stories/{id}/labels", h.UpdateLabels, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/links", h.GetStoryLinks, auth)
	app.Get("/workspaces/{workspaceId}/my-stories", h.MyStories, auth)
}

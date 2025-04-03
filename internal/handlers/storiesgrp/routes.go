package storiesgrp

import (
	"context"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/repo/attachmentsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/commentsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/linksrepo"
	"github.com/complexus-tech/projects-api/internal/repo/storiesrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB          *sqlx.DB
	Log         *logger.Logger
	SecretKey   string
	Publisher   *publisher.Publisher
	Validate    *validator.Validate
	AzureConfig azure.Config
}

func Routes(cfg Config, app *web.App) {
	storiesService := stories.New(cfg.Log, storiesrepo.New(cfg.Log, cfg.DB), cfg.Publisher)
	commentsService := comments.New(cfg.Log, commentsrepo.New(cfg.Log, cfg.DB))
	linksService := links.New(cfg.Log, linksrepo.New(cfg.Log, cfg.DB))

	attachmentsRepo := attachmentsrepo.New(cfg.Log, cfg.DB)
	azureBlobService, err := azure.New(cfg.AzureConfig, cfg.Log)
	if err != nil {
		cfg.Log.Error(context.Background(), "failed to initialize Azure blob service", "error", err)
		return
	}
	attachmentsService := attachments.New(cfg.Log, attachmentsRepo, azureBlobService, cfg.AzureConfig)

	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	h := New(storiesService, commentsService, linksService, attachmentsService, cfg.Log)

	// Stories
	app.Get("/workspaces/{workspaceId}/stories", h.List, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}", h.Get, auth)
	app.Post("/workspaces/{workspaceId}/stories", h.Create, auth)
	app.Put("/workspaces/{workspaceId}/stories/{id}", h.Update, auth)
	app.Delete("/workspaces/{workspaceId}/stories/{id}", h.Delete, auth)
	app.Post("/workspaces/{workspaceId}/stories/{id}/restore", h.Restore, auth)
	app.Post("/workspaces/{workspaceId}/stories/restore", h.BulkRestore, auth)
	app.Delete("/workspaces/{workspaceId}/stories", h.BulkDelete, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/activities", h.GetActivities, auth)
	app.Post("/workspaces/{workspaceId}/stories/{id}/duplicate", h.DuplicateStory, auth)

	// Comments
	app.Post("/workspaces/{workspaceId}/stories/{id}/comments", h.CreateComment, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/comments", h.GetComments, auth)
	app.Put("/workspaces/{workspaceId}/stories/{id}/labels", h.UpdateLabels, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/links", h.GetStoryLinks, auth)
	app.Get("/workspaces/{workspaceId}/my-stories", h.MyStories, auth)

	// Attachments
	app.Post("/workspaces/{workspaceId}/stories/{id}/attachments", h.UploadStoryAttachment, auth)
	app.Get("/workspaces/{workspaceId}/stories/{id}/attachments", h.GetAttachmentsForStory, auth)
	app.Delete("/workspaces/{workspaceId}/stories/{id}/attachments/{attachmentId}", h.DeleteAttachment, auth)
}

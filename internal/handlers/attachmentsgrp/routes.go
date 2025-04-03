package attachmentsgrp

import (
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/repo/attachmentsrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

// Config represents the configuration for the attachments routes
type Config struct {
	DB          *sqlx.DB
	Log         *logger.Logger
	SecretKey   string
	Validate    *validator.Validate
	AzureConfig azure.Config
}

// Routes registers the attachment routes
func Routes(cfg Config, app *web.App) {
	// Initialize the repository
	attachmentsRepo := attachmentsrepo.New(cfg.Log, cfg.DB)

	// Initialize the Azure blob service
	azureBlobService, err := azure.New(cfg.AzureConfig, cfg.Log)
	if err != nil {
		cfg.Log.Error(nil, "failed to initialize Azure blob service", "error", err)
		return
	}

	// Initialize the service
	attachmentsService := attachments.New(cfg.Log, attachmentsRepo, azureBlobService, cfg.AzureConfig)

	// Create the handler
	h := Handler{
		log:               cfg.Log,
		attachmentService: attachmentsService,
		validate:          cfg.Validate,
	}

	// Middleware
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	// Register routes
	app.Post("/api/attachments", h.Upload, auth)
	app.Get("/api/attachments/{id}", h.Get, auth)
	app.Delete("/api/attachments/{id}", h.Delete, auth)
	app.Get("/api/stories/{id}/attachments", h.GetForStory, auth)
	app.Post("/api/stories/{id}/attachments", h.LinkToStory, auth)
	app.Delete("/api/stories/{id}/attachments/{attachmentId}", h.UnlinkFromStory, auth)

	// New dedicated endpoints for specific attachment types
	app.Post("/api/stories/{id}/upload", h.UploadStoryAttachment, auth)
	app.Post("/api/users/profile-image", h.UploadProfileImage, auth)
	app.Post("/api/workspaces/{id}/logo", h.UploadWorkspaceLogo, auth)
}

// Route represents a Route in the API
type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Description string
}

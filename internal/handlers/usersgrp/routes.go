package usersgrp

import (
	"context"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/repo/attachmentsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/usersrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB            *sqlx.DB
	Log           *logger.Logger
	SecretKey     string
	GoogleService *google.Service
	Publisher     *publisher.Publisher
	TasksService  *tasks.Service
	AzureConfig   azure.Config
	Cache         *cache.Service
}

func Routes(cfg Config, app *web.App) {
	usersRepo := usersrepo.New(cfg.Log, cfg.DB)
	usersService := users.New(cfg.Log, usersRepo, cfg.TasksService)

	// Create attachments service for profile images
	attachmentsRepo := attachmentsrepo.New(cfg.Log, cfg.DB)
	azureBlobService, err := azure.New(cfg.AzureConfig, cfg.Log)
	if err != nil {
		cfg.Log.Error(context.Background(), "failed to initialize Azure blob service", "error", err)
		return
	}
	attachmentsService := attachments.New(cfg.Log, attachmentsRepo, azureBlobService, cfg.AzureConfig)

	h := New(usersService, attachmentsService, cfg.SecretKey, cfg.GoogleService, cfg.Publisher)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	// Public endpoints
	app.Post("/users/google/verify", h.GoogleAuth)
	app.Post("/users/verify/email", h.SendEmailVerification)
	app.Post("/users/verify/email/confirm", h.VerifyEmail)

	// Protected endpoints
	app.Get("/users/session/code", h.GenerateSessionCode, auth)
	app.Get("/workspaces/{workspaceSlug}/members", h.List, auth, workspace, gzip)
	app.Get("/users/profile", h.GetProfile, auth)
	app.Put("/users/profile", h.UpdateProfile, auth)
	app.Delete("/users/profile", h.DeleteProfile, auth)
	app.Post("/workspaces/switch", h.SwitchWorkspace, auth)

	// Profile image endpoints
	app.Post("/users/profile/image", h.UploadProfileImage, auth)
	app.Delete("/users/profile/image", h.DeleteProfileImage, auth)

	// Automation preferences endpoints
	app.Get("/workspaces/{workspaceSlug}/automation/preferences", h.GetAutomationPreferences, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/automation/preferences", h.UpdateAutomationPreferences, auth, workspace)

	// User Memory endpoints
	app.Get("/workspaces/{workspaceSlug}/users/memory", h.GetUserMemory, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/users/memory", h.UpsertUserMemory, auth, workspace)
}

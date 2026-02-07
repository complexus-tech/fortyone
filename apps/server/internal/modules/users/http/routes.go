package usershttp

import (
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB             *sqlx.DB
	Log            *logger.Logger
	SecretKey      string
	GoogleService  *google.Service
	Publisher      *publisher.Publisher
	TasksService   *tasks.Service
	StorageConfig  storage.Config
	StorageService storage.StorageService
	Cache          *cache.Service
}

func Routes(cfg Config, app *web.App) {
	usersRepo := usersrepository.New(cfg.Log, cfg.DB)
	usersService := users.New(cfg.Log, usersRepo, cfg.TasksService)

	// Create attachments service for profile images
	attachmentsRepo := attachmentsrepository.New(cfg.Log, cfg.DB)
	attachmentsService := attachments.New(cfg.Log, attachmentsRepo, cfg.StorageService, cfg.StorageConfig)

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
	app.Post("/users/session", h.CreateSession, auth)
	app.Delete("/users/session", h.ClearSession)
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
	// User Memory
	app.Post("/workspaces/{workspaceSlug}/users/memory", h.AddUserMemory, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/users/memory", h.ListUserMemories, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/users/memory/{id}", h.UpdateUserMemory, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/users/memory/{id}", h.DeleteUserMemory, auth, workspace)
}

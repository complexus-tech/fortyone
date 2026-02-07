package storieshttp

import (
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	linksrepository "github.com/complexus-tech/projects-api/internal/modules/links/repository"
	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	mentionsrepository "github.com/complexus-tech/projects-api/internal/modules/mentions/repository"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB             *sqlx.DB
	Log            *logger.Logger
	SecretKey      string
	Publisher      *publisher.Publisher
	Validate       *validator.Validate
	StorageConfig  storage.Config
	StorageService storage.StorageService
	Cache          *cache.Service
}

func Routes(cfg Config, app *web.App) {
	mentionsRepo := mentionsrepository.New(cfg.Log, cfg.DB)
	storiesService := stories.New(cfg.Log, storiesrepository.New(cfg.Log, cfg.DB), mentionsRepo, cfg.Publisher)
	commentsService := comments.New(cfg.Log, commentsrepository.New(cfg.Log, cfg.DB), mentionsRepo)
	linksService := links.New(cfg.Log, linksrepository.New(cfg.Log, cfg.DB))

	attachmentsRepo := attachmentsrepository.New(cfg.Log, cfg.DB)
	attachmentsService := attachments.New(cfg.Log, attachmentsRepo, cfg.StorageService, cfg.StorageConfig)

	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(storiesService, commentsService, linksService, attachmentsService, cfg.Cache, cfg.Log)

	// Stories
	app.Get("/workspaces/{workspaceSlug}/stories", h.List, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/stories/grouped", h.ListGrouped, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/stories/group", h.LoadMoreGroup, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/stories/by-category", h.ListByCategory, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}", h.Get, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/story-by-ref/{ref}", h.QueryByRef, auth, workspace, gzip)
	app.Post("/workspaces/{workspaceSlug}/stories", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/stories/{id}", h.Update, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/stories", h.BulkUpdate, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/stories/{id}", h.Delete, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/restore", h.Restore, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/stories/restore", h.BulkRestore, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/stories/archive", h.BulkArchive, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/stories/unarchive", h.BulkUnarchive, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/stories", h.BulkDelete, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/activities", h.GetActivities, auth, workspace, gzip)
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/duplicate", h.DuplicateStory, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/count", h.CountInWorkspace, auth, workspace)

	// Comments
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/comments", h.CreateComment, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/comments", h.GetComments, auth, workspace, gzip)
	app.Put("/workspaces/{workspaceSlug}/stories/{id}/labels", h.UpdateLabels, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/links", h.GetStoryLinks, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/my-stories", h.MyStories, auth, workspace, gzip)

	// Attachments
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/attachments", h.UploadStoryAttachment, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/attachments", h.GetAttachmentsForStory, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/stories/{id}/attachments/{attachmentId}", h.DeleteAttachment, auth, workspace)

	// Associations
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/associations", h.AddAssociation, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/stories/associations/{associationId}", h.RemoveAssociation, auth, workspace)
}

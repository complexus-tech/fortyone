package storieshttp

import (
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	Publisher         *publisher.Publisher
	StorageConfig     storage.Config
	StorageService    storage.StorageService
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Stories           *stories.Service
	Users             *users.Service
	Links             *links.Service
	Attachments       *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	storiesService := cfg.Stories
	linksService := cfg.Links
	attachmentsService := cfg.Attachments

	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	gzip := mid.Gzip(cfg.Log)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	h := New(storiesService, cfg.Users, linksService, attachmentsService, cfg.Cache, cfg.Log)

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
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/comments", h.CreateComment, auth, workspace, memberAndAdmin)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/comments", h.GetComments, auth, workspace, gzip)
	app.Put("/workspaces/{workspaceSlug}/stories/{id}/labels", h.UpdateLabels, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/stories/{id}/collaborators", h.UpdateCollaborators, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/stories/{id}/watch", h.SetWatching, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/links", h.GetStoryLinks, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/my-stories", h.MyStories, auth, workspace, gzip)

	// Attachments
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/attachments", h.UploadStoryAttachment, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/attachments", h.GetAttachmentsForStory, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/stories/{id}/attachments/{attachmentId}", h.DeleteAttachment, auth, workspace)

	// Inline media
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/media", h.UploadStoryMedia, auth, workspace, memberAndAdmin)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/media/{attachmentId}", h.ResolveStoryMedia, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/stories/{id}/media/{attachmentId}", h.DeleteStoryMedia, auth, workspace, memberAndAdmin)

	// Associations
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/associations", h.AddAssociation, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/stories/{id}/associations/{associationId}", h.UpdateAssociation, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/stories/associations/{associationId}", h.RemoveAssociation, auth, workspace)
}

package documentshttp

import (
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *documents.Service
	Attachments       *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	documentsService := cfg.Service
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	h := New(documentsService, cfg.Attachments, cfg.Log)

	app.Get("/workspaces/{workspaceSlug}/documents", h.List, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/documents/related", h.ListRelatedDocuments, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/documents", h.Create, auth, workspace, memberAndAdmin)
	app.Get("/workspaces/{workspaceSlug}/documents/{id}", h.Get, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/documents/{id}/duplicate", h.Duplicate, auth, workspace, memberAndAdmin)
	app.Get("/workspaces/{workspaceSlug}/documents/{id}/media/{attachmentId}", h.ResolveMedia, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/documents/{id}", h.Update, auth, workspace, memberAndAdmin)
	app.Post("/workspaces/{workspaceSlug}/documents/{id}/media", h.UploadMedia, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/documents/{id}/media/{attachmentId}", h.DeleteMedia, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/documents/{id}", h.Archive, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/documents/{id}/permanent", h.Delete, auth, workspace, memberAndAdmin)
	app.Put("/workspaces/{workspaceSlug}/documents/{id}/access", h.SetAccess, auth, workspace, memberAndAdmin)
	app.Post("/workspaces/{workspaceSlug}/documents/{id}/relationships", h.AddRelationship, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/documents/{id}/relationships/{entityType}/{entityId}", h.RemoveRelationship, auth, workspace, memberAndAdmin)
}

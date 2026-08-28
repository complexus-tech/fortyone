package workspaceshttp

import (
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
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
	Workspaces        *workspaces.Service
	Attachments       *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	h := New(cfg.Workspaces, cfg.Log, cfg.Attachments)

	app.Get("/portals/{portalSlug}", h.GetPortal)
	app.Get("/workspaces/{workspaceSlug}/portal", h.GetPortal)
	app.Get("/workspaces/{workspaceSlug}", h.Get, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}", h.Update, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}", h.Delete, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/restore", h.Restore, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/members", h.AddMember, auth, workspace, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/members/{userId}/role", h.UpdateMemberRole, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/members/{userId}", h.RemoveMember, auth, workspace, adminOnly)
	app.Post("/workspaces", h.Create, auth)
	app.Get("/workspaces", h.List, auth)
	app.Get("/workspaces/check-availability", h.CheckSlugAvailability)
	app.Get("/workspaces/{workspaceSlug}/settings", h.GetWorkspaceSettings, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/settings", h.UpdateWorkspaceSettings, auth, workspace, adminOnly)

	// Workspace logo endpoints
	app.Post("/workspaces/{workspaceSlug}/logo", h.UploadWorkspaceLogo, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/logo", h.DeleteWorkspaceLogo, auth, workspace, adminOnly)
}

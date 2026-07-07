package adminhttp

import (
	admin "github.com/complexus-tech/projects-api/internal/modules/admin/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log       *logger.Logger
	SecretKey string
	Service   *admin.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Log, cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)

	app.Get("/admin/me", h.GetCurrentAdmin, auth)
	app.Get("/admin/summary", h.GetDashboardSummary, auth)
	app.Get("/admin/workspaces", h.ListWorkspaces, auth)
	app.Get("/admin/workspaces/{workspaceID}", h.GetWorkspace, auth)
	app.Patch("/admin/workspaces/{workspaceID}/trial", h.UpdateWorkspaceTrial, auth)
	app.Patch("/admin/workspaces/{workspaceID}/deleted", h.UpdateWorkspaceDeleted, auth)
	app.Post("/admin/workspaces/{workspaceID}/subscription-sync", h.RequestWorkspaceSubscriptionSync, auth)
	app.Get("/admin/users", h.ListUsers, auth)
	app.Get("/admin/users/{userID}", h.GetUser, auth)
	app.Patch("/admin/users/{userID}/state", h.UpdateUserState, auth)
	app.Post("/admin/users/{userID}/session-revocation", h.RequestUserSessionRevocation, auth)
	app.Get("/admin/audit-logs", h.ListAuditLogs, auth)
	app.Get("/admin/audit-logs/export", h.ExportAuditLogs, auth)
	app.Get("/admin/notes", h.ListAdminNotes, auth)
	app.Post("/admin/notes", h.CreateAdminNote, auth)
}

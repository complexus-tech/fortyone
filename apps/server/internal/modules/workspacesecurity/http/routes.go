package workspacesecurityhttp

import (
	security "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *security.Service
}

func Routes(cfg Config, app *web.App) {
	h, err := New(cfg.Service, cfg.SecretKey)
	if err != nil {
		panic("workspace security cursor configuration: " + err.Error())
	}
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	app.Get("/workspaces/{workspaceSlug}/security/policy", h.Policy, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/security/policy", h.UpdatePolicy, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/security/sessions", h.Sessions, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/security/sessions/{sessionId}", h.RevokeSession, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/security/members/{userId}/revoke-sessions", h.RevokeMember, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/security/audit", h.Audit, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/security/audit/export", h.ExportAudit, auth, workspace)
}

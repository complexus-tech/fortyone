package googledrivehttp

import (
	googledrive "github.com/complexus-tech/projects-api/internal/modules/googledrive/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *googledrive.Service
}

func Routes(cfg Config, app *web.App) {
	handlers := New(cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	member := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	app.Get("/workspaces/{workspaceSlug}/integrations/google-drive", handlers.GetIntegration, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/google-drive/connect-session", handlers.CreateConnectSession, auth, workspace, member)
	// Disconnect is personal credential cleanup and remains available after a
	// workspace-role downgrade. The service can only remove the actor's binding.
	app.Delete("/workspaces/{workspaceSlug}/integrations/google-drive", handlers.Disconnect, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/integrations/google-drive/picker-session", handlers.CreatePickerSession, auth, workspace, member)

	app.Get("/workspaces/{workspaceSlug}/google-drive/files", handlers.ListFiles, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/google-drive/files", handlers.AttachFiles, auth, workspace, member)
	app.Post("/workspaces/{workspaceSlug}/google-drive/files/create", handlers.CreateFile, auth, workspace, member)
	app.Get("/workspaces/{workspaceSlug}/google-drive/files/{referenceId}/content", handlers.ReadContent, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/google-drive/files/{referenceId}/preview", handlers.ReadPreview, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/google-drive/files/{referenceId}/refresh", handlers.RefreshFile, auth, workspace, member)
	app.Post("/workspaces/{workspaceSlug}/google-drive/files/{referenceId}/imports", handlers.ImportFile, auth, workspace, member)
	app.Delete("/workspaces/{workspaceSlug}/google-drive/files/{referenceId}", handlers.DeleteFile, auth, workspace, member)

	app.Get("/integrations/google-drive/callback", handlers.CompleteOAuth, auth)
}

package developeroauthhttp

import (
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
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
	Service           Service
}

func Routes(cfg Config, app *web.App) {
	if cfg.Service == nil {
		panic("developer OAuth application management service is required")
	}

	handlers := New(cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	integrationScope := mid.RequireScopes(platformauth.ScopeIntegrationsManage)
	mutationLimit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{
		Scope: "developer-oauth-application-mutation", Limit: 60, Window: time.Hour,
	})

	app.Post("/workspaces/{workspaceSlug}/oauth-applications", handlers.CreateManagedApplication, auth, mutationLimit, workspace, adminOnly, integrationScope)
	app.Get("/workspaces/{workspaceSlug}/oauth-applications", handlers.ListManagedApplications, auth, workspace, adminOnly, integrationScope)
	app.Get("/workspaces/{workspaceSlug}/oauth-applications/{applicationId}/secrets", handlers.ListClientSecrets, auth, workspace, adminOnly, integrationScope)
	app.Post("/workspaces/{workspaceSlug}/oauth-applications/{applicationId}/secrets/rotate", handlers.RotateClientSecret, auth, mutationLimit, workspace, adminOnly, integrationScope)
	app.Delete("/workspaces/{workspaceSlug}/oauth-applications/{applicationId}/secrets/{secretId}", handlers.RevokeClientSecret, auth, mutationLimit, workspace, adminOnly, integrationScope)

	app.Post("/workspaces/{workspaceSlug}/oauth-application-installations", handlers.InstallApplication, auth, mutationLimit, workspace, adminOnly, integrationScope)
	app.Get("/workspaces/{workspaceSlug}/oauth-application-installations", handlers.ListApplicationInstallations, auth, workspace, adminOnly, integrationScope)
	app.Put("/workspaces/{workspaceSlug}/oauth-application-installations/{installationId}", handlers.UpdateApplicationInstallation, auth, mutationLimit, workspace, adminOnly, integrationScope)
	app.Delete("/workspaces/{workspaceSlug}/oauth-application-installations/{installationId}", handlers.RevokeApplicationInstallation, auth, mutationLimit, workspace, adminOnly, integrationScope)
}

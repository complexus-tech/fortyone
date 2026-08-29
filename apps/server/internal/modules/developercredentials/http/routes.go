package developercredentialshttp

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
		panic("developer credential service is required")
	}
	handlers := New(cfg.Service)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	serviceAccountScope := mid.RequireScopes(platformauth.ScopeServiceAccountsManage)
	mutationLimit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{
		Scope: "developer-credential-mutation", Limit: 60, Window: time.Hour,
	})

	app.Get("/workspaces/{workspaceSlug}/personal-access-tokens", handlers.ListPersonalTokens, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/personal-access-tokens", handlers.CreatePersonalToken, auth, mutationLimit, workspace)
	app.Post("/workspaces/{workspaceSlug}/personal-access-tokens/{credentialId}/rotate", handlers.RotatePersonalToken, auth, mutationLimit, workspace)
	app.Delete("/workspaces/{workspaceSlug}/personal-access-tokens/{credentialId}", handlers.RevokePersonalToken, auth, mutationLimit, workspace)

	app.Get("/workspaces/{workspaceSlug}/service-accounts", handlers.ListServiceAccounts, auth, workspace, adminOnly, serviceAccountScope)
	app.Post("/workspaces/{workspaceSlug}/service-accounts", handlers.CreateServiceAccount, auth, mutationLimit, workspace, adminOnly, serviceAccountScope)
	app.Delete("/workspaces/{workspaceSlug}/service-accounts/{principalId}", handlers.DisableServiceAccount, auth, mutationLimit, workspace, adminOnly, serviceAccountScope)
	app.Get("/workspaces/{workspaceSlug}/service-accounts/{principalId}/keys", handlers.ListServiceAccountKeys, auth, workspace, adminOnly, serviceAccountScope)
	app.Post("/workspaces/{workspaceSlug}/service-accounts/{principalId}/keys", handlers.CreateServiceAccountKey, auth, mutationLimit, workspace, adminOnly, serviceAccountScope)
	app.Post("/workspaces/{workspaceSlug}/service-accounts/{principalId}/keys/{credentialId}/rotate", handlers.RotateServiceAccountKey, auth, mutationLimit, workspace, adminOnly, serviceAccountScope)
	app.Delete("/workspaces/{workspaceSlug}/service-accounts/{principalId}/keys/{credentialId}", handlers.RevokeServiceAccountKey, auth, mutationLimit, workspace, adminOnly, serviceAccountScope)
}

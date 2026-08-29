package outboundwebhookshttp

import (
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
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
		panic("outbound webhook management service is required")
	}
	cursorKey, err := pagination.DeriveSigningKey(
		"active", []byte(cfg.SecretKey), "outbound-webhook-browser-management",
	)
	if err != nil {
		panic("derive outbound webhook management cursor key: " + err.Error())
	}
	cursors, err := pagination.NewCursorCodec[endpointCursor](cursorKey)
	if err != nil {
		panic("create outbound webhook management cursor codec: " + err.Error())
	}
	handlers := New(cfg.Service, cursors)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	webhookScope := mid.RequireScopes(platformauth.ScopeWebhooksManage)
	mutationLimit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{
		Scope: "outbound-webhook-management-mutation", Limit: 60, Window: time.Hour,
	})

	app.Get("/workspaces/{workspaceSlug}/webhook-endpoints", handlers.ListEndpoints, auth, workspace, adminOnly, webhookScope)
	app.Post("/workspaces/{workspaceSlug}/webhook-endpoints", handlers.CreateEndpoint, auth, mutationLimit, workspace, adminOnly, webhookScope)
	app.Put("/workspaces/{workspaceSlug}/webhook-endpoints/{endpointId}/subscriptions", handlers.ReplaceSubscriptions, auth, mutationLimit, workspace, adminOnly, webhookScope)
	app.Post("/workspaces/{workspaceSlug}/webhook-endpoints/{endpointId}/rotate-secret", handlers.RotateSecret, auth, mutationLimit, workspace, adminOnly, webhookScope)
	app.Post("/workspaces/{workspaceSlug}/webhook-endpoints/{endpointId}/disable", handlers.DisableEndpoint, auth, mutationLimit, workspace, adminOnly, webhookScope)
}

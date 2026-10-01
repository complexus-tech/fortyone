package scimhttp

import (
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Service              Service
	PublicURL, SecretKey string
	Log                  *logger.Logger
	BrowserSessions      mid.SessionResolver
	WorkspaceResolver    mid.WorkspaceResolver
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service, cfg.PublicURL)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	app.Get("/workspaces/{workspaceSlug}/security/scim", h.AdminStatus, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/security/scim/credentials", h.Mint, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/security/scim/credentials/{credentialId}", h.Revoke, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/security/scim/retry-seats", h.RetrySeats, auth, workspace)
	app.Get("/scim/v2/{workspaceSlug}/Users", h.Users)
	app.Post("/scim/v2/{workspaceSlug}/Users", h.Users)
	app.Get("/scim/v2/{workspaceSlug}/Users/{resourceId}", h.User)
	app.Put("/scim/v2/{workspaceSlug}/Users/{resourceId}", h.User)
	app.Patch("/scim/v2/{workspaceSlug}/Users/{resourceId}", h.User)
	app.Delete("/scim/v2/{workspaceSlug}/Users/{resourceId}", h.User)
	app.Get("/scim/v2/{workspaceSlug}/ServiceProviderConfig", h.Discovery)
	app.Get("/scim/v2/{workspaceSlug}/ResourceTypes", h.Discovery)
	app.Get("/scim/v2/{workspaceSlug}/ResourceTypes/{resourceId}", h.Discovery)
	app.Get("/scim/v2/{workspaceSlug}/Schemas", h.Discovery)
	app.Get("/scim/v2/{workspaceSlug}/Schemas/{resourceId}", h.Discovery)
}

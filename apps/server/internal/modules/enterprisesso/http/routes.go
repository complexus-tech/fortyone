package enterprisessohttp

import (
	sso "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Service                                          *sso.Service
	States                                           StateStore
	Issuer                                           SessionIssuer
	CallbackURL, WebsiteURL, CookieDomain, SecretKey string
	Production                                       bool
	Log                                              *logger.Logger
	BrowserSessions                                  mid.SessionResolver
	WorkspaceResolver                                mid.WorkspaceResolver
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service, cfg.States, cfg.Issuer, cfg.CallbackURL, cfg.WebsiteURL, cfg.CookieDomain, cfg.Production)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	optional := mid.OptionalAuth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	app.Get("/auth/sso/{workspaceSlug}", h.Start, optional)
	app.Get("/auth/sso/callback", h.Callback, optional)
	app.Get("/auth/sso/{workspaceSlug}/status", h.PublicStatus)
	app.Get("/workspaces/{workspaceSlug}/security/sso", h.Get, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/security/sso", h.Create, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/security/sso", h.Update, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/security/sso", h.Archive, auth, workspace)
}

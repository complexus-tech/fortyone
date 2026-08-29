package agentreadinesshttp

import "github.com/complexus-tech/projects-api/pkg/web"

func Routes(cfg Config, app *web.App) {
	h := New(cfg)
	authorizationLimit := oauthGlobalRateLimit(
		cfg,
		"authorization",
		oauthAuthorizationRequestLimit,
		oauthAuthorizationRequestWindow,
	)
	tokenLimit := oauthGlobalRateLimit(cfg, "token", oauthTokenRequestLimit, oauthTokenRequestWindow)
	registrationLimit := oauthGlobalRateLimit(
		cfg,
		"registration",
		oauthRegistrationRequestLimit,
		oauthRegistrationRequestWindow,
	)
	revocationLimit := oauthGlobalRateLimit(
		cfg,
		"revocation",
		oauthRevocationRequestLimit,
		oauthRevocationRequestWindow,
	)
	app.Get("/openapi.json", h.OpenAPI)
	app.Get("/mcp", h.MCP)
	app.Post("/mcp", h.MCP)
	app.Delete("/mcp", h.MCP)
	app.Get("/.well-known/oauth-protected-resource", h.ProtectedResourceMetadata)
	app.Get("/.well-known/oauth-protected-resource/mcp", h.ProtectedResourceMetadata)
	app.Get("/.well-known/oauth-protected-resource/api/v1", h.APIProtectedResourceMetadata)
	app.Get("/.well-known/oauth-authorization-server", h.AuthorizationServerMetadata)
	app.Get("/oauth/authorize", h.Authorize, authorizationLimit)
	app.Post("/oauth/authorize", h.Authorize, authorizationLimit)
	app.Post("/oauth/token", h.Token, tokenLimit)
	app.Post("/oauth/register", h.RegisterClient, registrationLimit)
	app.Post("/oauth/revoke", h.Revoke, revocationLimit)
}

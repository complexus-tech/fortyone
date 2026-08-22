package agentreadinesshttp

import "github.com/complexus-tech/projects-api/pkg/web"

func Routes(cfg Config, app *web.App) {
	h := New(cfg)
	app.Get("/openapi.json", h.OpenAPI)
	app.Get("/mcp", h.MCP)
	app.Post("/mcp", h.MCP)
	app.Delete("/mcp", h.MCP)
	app.Get("/.well-known/oauth-protected-resource", h.ProtectedResourceMetadata)
	app.Get("/.well-known/oauth-protected-resource/mcp", h.ProtectedResourceMetadata)
	app.Get("/.well-known/oauth-authorization-server", h.AuthorizationServerMetadata)
	app.Get("/oauth/authorize", h.Authorize)
	app.Post("/oauth/authorize", h.Authorize)
	app.Post("/oauth/token", h.Token)
	app.Post("/oauth/register", h.RegisterClient)
	app.Post("/oauth/revoke", h.Revoke)
}

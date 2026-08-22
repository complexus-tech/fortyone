package agentreadinesshttp

import "github.com/complexus-tech/projects-api/pkg/web"

func Routes(app *web.App) {
	h := New()
	app.Get("/openapi.json", h.OpenAPI)
	app.Get("/mcp", h.MCPGet)
	app.Post("/mcp", h.MCPPost)
}

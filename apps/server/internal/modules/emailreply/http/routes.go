package emailreplyhttp

import "github.com/complexus-tech/projects-api/pkg/web"

type Config struct {
	Service Ingress
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service)
	app.Post("/webhooks/brevo/inbound-email-processed", h.HandleInboundEmailProcessed)
}

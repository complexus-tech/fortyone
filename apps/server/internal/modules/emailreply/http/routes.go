package emailreplyhttp

import (
	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Service *emailreply.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service)
	app.Post("/webhooks/brevo/inbound-email-processed", h.HandleInboundEmailProcessed)
}

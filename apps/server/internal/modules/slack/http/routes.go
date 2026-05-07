package slackhttp

import (
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log     *logger.Logger
	Service *slack.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Log, cfg.Service)
	app.Post("/integrations/slack/events", h.HandleEvents)
	app.Post("/integrations/slack/interactivity", h.HandleInteractivity)
	app.Post("/integrations/slack/commands", h.HandleCommands)
}

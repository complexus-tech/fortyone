package handlers

import (
	"net/http"
	"os"

	"github.com/complexus-tech/projects-api/internal/web"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
)

type APIConfig struct {
	DB       *sqlx.DB
	Shutdown chan os.Signal
	Log      *logger.Logger
	Tracer   trace.Tracer
}

func API(cfg APIConfig) http.Handler {

	app := web.NewApp(cfg.Shutdown, cfg.Tracer)
	app.StrictSlash(false)

	// Register health check handler.
	h := NewHealthHandler(cfg.Log, cfg.DB)
	app.Get("/readiness", h.readiness)
	app.Get("/liveness", h.liveness)

	// Register issues handlers.
	i := NewIssuesHandlers(cfg.Log, cfg.DB)
	app.Get("/issues/{id:[0-9]+}", i.get)
	app.Get("/issues", i.list)

	return app
}

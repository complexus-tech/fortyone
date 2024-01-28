package router

import (
	"net/http"
	"os"

	"github.com/complexus-tech/projects-api/internal/handlers"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	DB       *sqlx.DB
	Shutdown chan os.Signal
	Log      *logger.Logger
	Tracer   trace.Tracer
}

// New returns a new HTTP handler that defines all the API routes.
func New(cfg Config) http.Handler {

	app := web.New(cfg.Shutdown, cfg.Tracer)
	app.StrictSlash(false)

	//Register health check handler.
	h := handlers.NewHealthHandler(cfg.Log, cfg.DB)
	app.Get("/readiness", h.Readiness)
	app.Get("/liveness", h.Liveness)

	// Register issues handlers.
	i := handlers.NewIssuesHandlers(cfg.Log, cfg.DB)
	app.Get("/issues/{id}", i.Get)
	app.Get("/issues", i.List)

	return app
}

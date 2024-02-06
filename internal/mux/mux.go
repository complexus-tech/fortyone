package mux

import (
	"net/http"
	"os"

	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
)

// RouteAdder is an interface that defines a method to add routes to a web.App.
type RouteAdder interface {
	BuildAllRoutes(app *web.App, cfg Config)
}

// Config defines the configuration for the mux.
type Config struct {
	DB       *sqlx.DB
	Shutdown chan os.Signal
	Log      *logger.Logger
	Tracer   trace.Tracer
}

// New returns a new HTTP handler that defines all the API routes.
func New(cfg Config, ra RouteAdder) http.Handler {

	app := web.New(cfg.Shutdown, cfg.Tracer, mid.Logger(cfg.Log))
	app.StrictSlash(false)

	ra.BuildAllRoutes(app, cfg)

	return app
}

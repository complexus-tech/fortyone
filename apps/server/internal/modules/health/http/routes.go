package healthhttp

import (
	"context"

	platformhealth "github.com/complexus-tech/projects-api/internal/platform/health"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type ReadinessReporter interface {
	Report(context.Context) platformhealth.Report
}

type Config struct {
	Log       *logger.Logger
	Readiness ReadinessReporter
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Log, cfg.Readiness)

	app.Get("/readiness", h.Readiness)
	app.Get("/liveness", h.Liveness)

}

package healthhttp

import (
	"context"
	"net/http"
	"os"
	"runtime"

	platformhealth "github.com/complexus-tech/projects-api/internal/platform/health"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type readinessReporter interface {
	Report(context.Context) platformhealth.Report
}

type healthCheck struct {
	log       *logger.Logger
	readiness readinessReporter
}

// New returns the API process health handlers.
func New(log *logger.Logger, readiness readinessReporter) *healthCheck {
	return &healthCheck{
		log:       log,
		readiness: readiness,
	}
}

// Readiness reports whether the supervisor is accepting traffic and every
// required dependency is reachable.
func (h *healthCheck) Readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	report := platformhealth.Report{
		Status: "not_ready",
		Phase:  platformhealth.PhaseFailed,
		Checks: map[string]string{},
	}
	if h.readiness != nil {
		report = h.readiness.Report(ctx)
	}

	statusCode := http.StatusOK
	status := "ok"
	if report.Status != "ready" {
		statusCode = http.StatusServiceUnavailable
		status = report.Status
		if h.log != nil {
			h.log.Warn(ctx, "API is not ready", "phase", report.Phase, "checks", report.Checks)
		}
	}

	response := struct {
		Status string               `json:"status"`
		Phase  platformhealth.Phase `json:"phase"`
		Checks map[string]string    `json:"checks"`
	}{
		Status: status,
		Phase:  report.Phase,
		Checks: report.Checks,
	}
	return web.Respond(ctx, w, response, statusCode)
}

// Liveness reports process metadata used by the existing infrastructure probe
// and diagnostics contract.
func (h *healthCheck) Liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	data := struct {
		Status     string `json:"status"`
		Hostname   string `json:"hostname"`
		Name       string `json:"name,omitempty"`
		PodIP      string `json:"podIP,omitempty"`
		Node       string `json:"node,omitempty"`
		Namespace  string `json:"namespace,omitempty"`
		GOMAXPROCS int    `json:"GOMAXPROCS,omitempty"`
	}{
		Status:     "ok",
		Hostname:   host,
		Name:       os.Getenv("KUBERNETES_NAME"),
		PodIP:      os.Getenv("KUBERNETES_POD_IP"),
		Node:       os.Getenv("KUBERNETES_NODE_NAME"),
		Namespace:  os.Getenv("KUBERNETES_NAMESPACE"),
		GOMAXPROCS: runtime.GOMAXPROCS(0),
	}

	return web.Respond(ctx, w, data, http.StatusOK)
}

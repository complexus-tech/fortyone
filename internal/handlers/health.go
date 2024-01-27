package handlers

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type HealthHandler interface {
	readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type healthCheck struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewProfHandlers returns a new profHandlers instance.
func NewHealthHandler(log *logger.Logger, db *sqlx.DB) HealthHandler {
	return &healthCheck{
		log: log,
		db:  db,
	}
}

// readiness checks if the db is ready and service is ready to handle requests.
func (h *healthCheck) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	statusCode := http.StatusOK
	status := "ok"

	if err := h.db.PingContext(ctx); err != nil {
		statusCode = http.StatusServiceUnavailable
		status = "db not ready"
		h.log.Error(ctx, "db: not ready.", "status", status)
	}

	data := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	return web.Respond(ctx, w, data, statusCode)
}

func (h *healthCheck) liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

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

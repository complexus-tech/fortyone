package healthgrp

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

type healthCheck struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewHealthHandlers returns a new profHandlers instance.
func New(log *logger.Logger, db *sqlx.DB) *healthCheck {
	return &healthCheck{
		log: log,
		db:  db,
	}
}

// Readiness checks if the db is ready and service is ready to handle requests.
func (h *healthCheck) Readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	// h.brevoService.SendTemplatedEmail(ctx, brevo.SendTemplatedEmailRequest{

	// 	To: []brevo.EmailRecipient{{Email: "josemukorivo@gmail.com", Name: "Test"}},

	// 	TemplateID: 3,
	// 	Params: map[string]any{
	// 		"USER_NAME":            "Joseph",
	// 		"USER_EMAIL":           "joseph@complexus.app",
	// 		"WORKSPACE_NAME":       "Test",
	// 		"WORKSPACE_URL":        "https://test.com",
	// 		"APP_URL":              "https://test.com",
	// 		"NOTIFICATION_TITLE":   "Test",
	// 		"NOTIFICATION_MESSAGE": "Test",
	// 	},
	// })

	// res, err := h.brevoService.CreateOrUpdateContact(context.Background(), brevo.CreateOrUpdateContactRequest{
	// 	Email: "joseph@complexus.app",
	// 	Attributes: map[string]any{
	// 		"NAME": "Joseph Mukorivo",
	// 	},
	// 	ListIDs: []int64{6},
	// })
	// if err != nil {
	// 	h.log.Error(ctx, "Failed to create or update contact", "error", err)
	// 	return nil
	// }

	// fmt.Println("Brevo contact response", res)

	return web.Respond(context.Background(), w, data, statusCode)
}

// liveness checks if the service is alive and ready to handle requests.
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

package mux

import (
	"net/http"
	"os"

	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/azure"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/email"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v82/client"
	"go.opentelemetry.io/otel/trace"
)

// RouteAdder is an interface that defines a method to add routes to a web.App.
type RouteAdder interface {
	BuildAllRoutes(app *web.App, cfg Config)
}

// Config defines the configuration for the mux.
type Config struct {
	DB            *sqlx.DB
	Redis         *redis.Client
	Publisher     *publisher.Publisher
	Shutdown      chan os.Signal
	Log           *logger.Logger
	Tracer        trace.Tracer
	SecretKey     string
	EmailService  email.Service
	GoogleService *google.Service
	Validate      *validator.Validate
	AzureConfig   azure.Config
	Cache         *cache.Service
	TasksService  *tasks.Service
	StripeClient  *client.API
	WebhookSecret string
	SSEHub        *sse.Hub
	CorsOrigin    string
}

// New returns a new HTTP handler that defines all the API routes.
func New(cfg Config, ra RouteAdder) http.Handler {
	app := web.New(cfg.Shutdown, cfg.Tracer, mid.Logger(cfg.Log))
	app.StrictSlash(false)

	ra.BuildAllRoutes(app, cfg)

	return app
}

package mux

import (
	"net/http"
	"os"

	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	platformhealth "github.com/complexus-tech/projects-api/internal/platform/health"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/microsoft"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
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
	Redis                       *redis.Client
	Publisher                   *publisher.Publisher
	Shutdown                    chan os.Signal
	Log                         *logger.Logger
	Tracer                      trace.Tracer
	DeploymentMode              deployment.Mode
	Readiness                   *platformhealth.Readiness
	SecretKey                   string
	FeedbackIngressSecret       string
	FeedbackSecurityKey         string
	EmailReplySecurityKey       string
	MessagingMutationHMACKey    string
	CookieDomain                string
	EmailService                mailer.Service
	BrevoService                *brevo.Service
	GoogleService               *google.Service
	MicrosoftService            *microsoft.Service
	GoogleCalendarWebhookURL    string
	MicrosoftCalendarWebhookURL string
	Cache                       *cache.Service
	TasksService                *tasks.Service
	StripeClient                *client.API
	StorageConfig               storage.Config
	StorageService              storage.StorageService
	WebhookSecret               string
	WebsiteURL                  string
	APIPublicURL                string
	MCPLoginURL                 string
	GitHubAppID                 int64
	GitHubAppSlug               string
	GitHubClientID              string
	GitHubClientSecret          string
	GitHubUserID                uuid.UUID
	GitHubKeyBase64             string
	GitHubRedirect              string
	GitHubWebhook               string
	GitHubWebhookPayloadSecret  string
	SlackSigningSecret          string
	SlackClientID               string
	SlackClientSecret           string
	SlackRedirectURL            string
	SlackWebhookPayloadSecret   string
	FigmaClientID               string
	FigmaClientSecret           string
	FigmaRedirectURL            string
	FigmaWebhookURL             string
	FigmaWebhookPayloadSecret   string
	AIAPIKey                    string
	SSEHub                      *sse.Hub
	AllowedOrigins              web.OriginPolicy
}

// New returns a new HTTP handler that defines all the API routes.
func New(cfg Config, ra RouteAdder) http.Handler {
	app := web.New(
		cfg.Shutdown,
		cfg.Tracer,
		mid.Logger(cfg.Log),
		mid.RequireTrustedBrowserOrigin(cfg.AllowedOrigins),
	)
	app.SetOriginPolicy(cfg.AllowedOrigins)
	app.SetLogger(cfg.Log)
	app.StrictSlash(false)

	ra.BuildAllRoutes(app, cfg)

	return app
}

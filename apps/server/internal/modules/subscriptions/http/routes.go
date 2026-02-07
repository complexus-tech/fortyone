package subscriptionshttp

import (
	mentionsrepository "github.com/complexus-tech/projects-api/internal/modules/mentions/repository"
	objectivestatusrepository "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/repository"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	statesrepository "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	subscriptionsrepository "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository"
	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stripe/stripe-go/v82/client"
)

type Config struct {
	DB            *sqlx.DB
	Log           *logger.Logger
	SecretKey     string
	StripeClient  *client.API
	WebhookSecret string
	Publisher     *publisher.Publisher
	TasksService  *tasks.Service
	SystemUserID  uuid.UUID
	Cache         *cache.Service
}

func Routes(cfg Config, app *web.App) {
	// Initialize repository and service
	subsRepo := subscriptionsrepository.New(cfg.Log, cfg.DB)
	subsService := subscriptions.New(
		cfg.Log,
		subsRepo,
		cfg.StripeClient,
		cfg.WebhookSecret,
		cfg.TasksService,
	)

	teamsService := teams.New(cfg.Log, teamsrepository.New(cfg.Log, cfg.DB))
	mentionsRepo := mentionsrepository.New(cfg.Log, cfg.DB)
	storiesService := stories.New(cfg.Log, storiesrepository.New(cfg.Log, cfg.DB), mentionsRepo, cfg.Publisher)
	statusesService := states.New(cfg.Log, statesrepository.New(cfg.Log, cfg.DB))
	objectivestatusService := objectivestatus.New(cfg.Log, objectivestatusrepository.New(cfg.Log, cfg.DB))
	usersService := users.New(cfg.Log, usersrepository.New(cfg.Log, cfg.DB), cfg.TasksService)
	workspacesService := workspaces.New(cfg.Log, workspacesrepository.New(cfg.Log, cfg.DB), cfg.DB, teamsService, storiesService, statusesService, usersService, objectivestatusService, subsService, nil, cfg.Cache, cfg.SystemUserID, cfg.Publisher, cfg.TasksService)

	h := New(subsService, usersService, workspacesService, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	app.Post("/workspaces/{workspaceSlug}/subscriptions/checkout", h.CreateCheckoutSession, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/portal", h.CreateCustomerPortal, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/subscription", h.GetSubscription, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/invoices", h.GetInvoices, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/add-seat", h.AddSeat, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/subscriptions/change-plan", h.ChangeSubscriptionPlan, auth, workspace)

	// Public webhook endpoint
	app.Post("/webhooks/stripe", h.HandleWebhook)
}

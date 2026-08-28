package api

import (
	"fmt"

	"github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksrepository "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildSlackWebhookGateway(
	pool *pgxpool.Pool,
	repository slack.WebhookInstallationRepository,
	queue slack.EventQueue,
	config slack.Config,
) (*webhooks.Gateway, error) {
	if pool == nil {
		return nil, fmt.Errorf("build Slack webhook gateway: native database pool is required")
	}
	catalog, err := providers.BuiltInRegistry()
	if err != nil {
		return nil, fmt.Errorf("build Slack webhook provider catalog: %w", err)
	}
	runtime, err := slack.NewWebhookRuntime(repository, queue, config)
	if err != nil {
		return nil, fmt.Errorf("build Slack webhook runtime: %w", err)
	}
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime)
	if err != nil {
		return nil, fmt.Errorf("register Slack webhook runtime: %w", err)
	}
	gateway, err := webhooks.NewGateway(
		webhooksrepository.New(pool),
		runtimes,
		webhooks.Config{},
	)
	if err != nil {
		return nil, fmt.Errorf("build Slack webhook gateway: %w", err)
	}
	return gateway, nil
}

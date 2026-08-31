package workerbootstrap

import (
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksrepository "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildSlackWebhookRuntime(
	pool *pgxpool.Pool,
	repository slack.WebhookInstallationRepository,
	queue slack.EventQueue,
	config slack.Config,
) (*webhooks.Gateway, *webhooksrepository.Repository, error) {
	if pool == nil {
		return nil, nil, fmt.Errorf("build Slack webhook runtime: native database pool is required")
	}
	catalog, err := providers.BuiltInRegistry()
	if err != nil {
		return nil, nil, fmt.Errorf("build Slack webhook provider catalog: %w", err)
	}
	runtime, err := slack.NewWebhookRuntime(repository, queue, config)
	if err != nil {
		if errors.Is(err, slack.ErrSlackSigningSecretNotConfigured) {
			return nil, nil, workerSlackSigningSecretMissingFailure.Wrap(err)
		}
		return nil, nil, fmt.Errorf("build Slack webhook adapter: %w", err)
	}
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime)
	if err != nil {
		return nil, nil, fmt.Errorf("register Slack webhook runtime: %w", err)
	}
	inbox := webhooksrepository.New(pool)
	gateway, err := webhooks.NewGateway(inbox, runtimes, webhooks.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("build Slack webhook gateway: %w", err)
	}
	return gateway, inbox, nil
}

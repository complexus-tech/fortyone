package workerbootstrap

import (
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksrepository "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildGitHubWebhookRuntime(
	pool *pgxpool.Pool,
	repository github.WebhookInstallationRepository,
	queue github.WebhookQueue,
	config github.Config,
) (*webhooks.Gateway, *webhooksrepository.Repository, github.WebhookPayloadOpener, error) {
	if strings.TrimSpace(config.WebhookSecret) == "" {
		return nil, nil, nil, nil
	}
	if pool == nil {
		return nil, nil, nil, fmt.Errorf("build GitHub webhook runtime: native database pool is required")
	}
	catalog, err := providers.BuiltInRegistry()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build GitHub webhook provider catalog: %w", err)
	}
	runtime, err := github.NewWebhookRuntime(repository, queue, config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build GitHub webhook adapter: %w", err)
	}
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime.Registration)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("register GitHub webhook runtime: %w", err)
	}
	inbox := webhooksrepository.New(pool)
	gateway, err := webhooks.NewGateway(inbox, runtimes, webhooks.Config{MaxBodyBytes: 1 << 20})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build GitHub webhook gateway: %w", err)
	}
	return gateway, inbox, runtime.Payloads, nil
}

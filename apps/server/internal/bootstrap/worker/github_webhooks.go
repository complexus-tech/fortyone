package workerbootstrap

import (
	"fmt"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	webhooksrepository "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildGitHubWebhookRuntime(
	pool *pgxpool.Pool,
	queue github.WebhookQueue,
	config github.Config,
) (*webhooksrepository.Repository, github.WebhookWorkerRuntime, error) {
	if pool == nil {
		return nil, github.WebhookWorkerRuntime{}, fmt.Errorf("build GitHub webhook runtime: native database pool is required")
	}
	runtime, err := github.NewWebhookWorkerRuntime(queue, config.WebhookPayloadSecret)
	if err != nil {
		return nil, github.WebhookWorkerRuntime{}, fmt.Errorf("build GitHub webhook adapter: %w", err)
	}
	return webhooksrepository.New(pool), runtime, nil
}

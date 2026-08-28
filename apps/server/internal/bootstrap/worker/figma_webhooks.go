package workerbootstrap

import (
	"fmt"

	"github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksrepository "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildFigmaWebhookRuntime(
	pool *pgxpool.Pool,
	repository figma.Repository,
	queue figma.WebhookQueue,
	config figma.Config,
) (*webhooks.Gateway, *webhooksrepository.Repository, figma.WebhookPayloadOpener, error) {
	if pool == nil || repository == nil || queue == nil {
		return nil, nil, nil, fmt.Errorf("build Figma webhook runtime: dependencies are required")
	}
	catalog, err := providers.BuiltInRegistry()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Figma webhook provider catalog: %w", err)
	}
	runtime, err := figma.NewWebhookRuntime(repository, queue, config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Figma webhook adapter: %w", err)
	}
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime.Registration)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("register Figma webhook runtime: %w", err)
	}
	inbox := webhooksrepository.New(pool)
	gateway, err := webhooks.NewGateway(
		inbox,
		runtimes,
		webhooks.Config{MaxBodyBytes: 1 << 20},
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Figma webhook gateway: %w", err)
	}
	return gateway, inbox, runtime.Payloads, nil
}

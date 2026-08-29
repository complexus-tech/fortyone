package api

import (
	"fmt"

	"github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	figmarepository "github.com/complexus-tech/projects-api/internal/modules/figma/repository"
	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksrepository "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildFigmaWebhookGateway(
	pool *pgxpool.Pool,
	repository *figmarepository.Repository,
	queue figma.WebhookQueue,
	config figma.Config,
) (*webhooks.Gateway, *webhooksrepository.Repository, figma.WebhookPayloadOpener, error) {
	if pool == nil || repository == nil || queue == nil {
		return nil, nil, nil, fmt.Errorf("build Figma webhook gateway: dependencies are required")
	}
	catalog, err := providers.BuiltInRegistry()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Figma webhook provider catalog: %w", err)
	}
	runtime, err := figma.NewWebhookRuntime(repository, queue, config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Figma webhook runtime: %w", err)
	}
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime.Registration)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("register Figma webhook runtime: %w", err)
	}
	inbox := webhooksrepository.New(pool)
	gateway, err := webhooks.NewGateway(
		inbox,
		runtimes,
		webhooks.Config{MaxBodyBytes: maxFigmaWebhookBodyBytes},
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build Figma webhook gateway: %w", err)
	}
	return gateway, inbox, runtime.Payloads, nil
}

const maxFigmaWebhookBodyBytes = 1 << 20

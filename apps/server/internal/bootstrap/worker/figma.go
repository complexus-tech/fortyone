package workerbootstrap

import (
	"context"
	"errors"
	"fmt"

	bootstrapproviders "github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	figmamigration "github.com/complexus-tech/projects-api/internal/modules/figma/credentialmigration"
	figmarepository "github.com/complexus-tech/projects-api/internal/modules/figma/repository"
	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/jackc/pgx/v5/pgxpool"
)

type figmaWorker struct {
	service    *figma.Service
	repository *figmarepository.Repository
}

func buildFigmaWorker(
	log *logger.Logger,
	pool *pgxpool.Pool,
	cfg Config,
	updates *publisher.Publisher,
	queue *tasks.Service,
	vault *credentialvault.Vault,
	webhookPayloadSecret string,
) (figmaWorker, error) {
	if log == nil || pool == nil || queue == nil || vault == nil {
		return figmaWorker{}, errors.New("build Figma worker: dependencies are required")
	}
	repository := figmarepository.New(pool)
	config := figma.Config{
		ClientID: cfg.Figma.ClientID, ClientSecret: cfg.Figma.ClientSecret,
		RedirectURL: cfg.Figma.RedirectURL, WebhookURL: cfg.Figma.WebhookURL,
		WebsiteURL: cfg.Website.URL, Credentials: vault,
		WebhookPayloadSecret: webhookPayloadSecret,
	}
	gateway, inbox, payloads, err := buildFigmaWebhookRuntime(pool, repository, queue, config)
	if err != nil {
		return figmaWorker{}, err
	}
	config.WebhookGateway = gateway
	config.WebhookInbox = inbox
	config.WebhookPayloads = payloads

	storyService := stories.New(
		log,
		storiesrepository.New(log, pool),
		updates,
		queue,
	)
	storyService.ConfigureCommentCreator(buildStoryCommentCreator(log, pool))
	storyAdapter, err := bootstrapproviders.NewFigmaStoryAdapter(storyService)
	if err != nil {
		return figmaWorker{}, fmt.Errorf("build Figma story adapter: %w", err)
	}
	return figmaWorker{
		service:    figma.New(log, repository, storyAdapter, config),
		repository: repository,
	}, nil
}

func migrateLegacyFigmaCredentials(
	ctx context.Context,
	worker figmaWorker,
	vault *credentialvault.Vault,
	legacySecret string,
) (figmamigration.Report, error) {
	if worker.repository == nil {
		return figmamigration.Report{}, errors.New("migrate legacy Figma credentials: repository is required")
	}
	migrator, err := figmamigration.New(worker.repository, vault, legacySecret)
	if err != nil {
		return figmamigration.Report{}, fmt.Errorf("migrate legacy Figma credentials: %w", err)
	}
	report, err := migrator.Run(ctx)
	if err != nil {
		return report, fmt.Errorf("migrate legacy Figma credentials: %w", err)
	}
	return report, nil
}

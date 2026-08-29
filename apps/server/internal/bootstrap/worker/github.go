package workerbootstrap

import (
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/bootstrap/githubadapter"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	githubrepository "github.com/complexus-tech/projects-api/internal/modules/github/repository"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	integrationrequestsrepository "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type githubCompatibilityDependencies struct {
	requestStore              github.RequestStore
	autoSchedulingEligibility stories.AutoSchedulingEligibilityChecker
}

func (dependencies githubCompatibilityDependencies) validate() error {
	if dependencies.requestStore == nil {
		return errors.New("GitHub request store is required")
	}
	if dependencies.autoSchedulingEligibility == nil {
		return errors.New("GitHub auto-scheduling eligibility checker is required")
	}
	return nil
}

// buildGitHubCompatibilityDependencies keeps GitHub's compatibility workflow
// on consumer-owned interfaces while delegating Maya entitlement to its typed
// repository boundary.
func buildGitHubCompatibilityDependencies(
	pool *pgxpool.Pool,
	access mayaWorkspaceAccess,
) githubCompatibilityDependencies {
	var eligibility stories.AutoSchedulingEligibilityChecker
	if access != nil {
		eligibility = access.WorkspaceCanUseMaya
	}
	return githubCompatibilityDependencies{
		requestStore: githubadapter.NewRequestStore(
			integrationrequestsrepository.New(pool),
		),
		autoSchedulingEligibility: eligibility,
	}
}

func buildGitHubWorkerService(
	log *logger.Logger,
	pool *pgxpool.Pool,
	cfg Config,
	githubActorID, mayaActorID uuid.UUID,
	updates *publisher.Publisher,
	queue *tasks.Service,
	attachmentsService *attachments.Service,
	vault *credentialvault.Vault,
	webhookPayloadSecret string,
	compatibility githubCompatibilityDependencies,
) (*github.Service, error) {
	if err := compatibility.validate(); err != nil {
		return nil, err
	}
	repository := githubrepository.New(pool)
	config := github.Config{
		AppID:                cfg.GitHub.AppID,
		AppSlug:              cfg.GitHub.AppSlug,
		PrivateKeyBase64:     cfg.GitHub.PrivateKeyBase64,
		RedirectURL:          cfg.GitHub.RedirectURL,
		WebhookSecret:        cfg.GitHub.WebhookSecret,
		WebhookPayloadSecret: webhookPayloadSecret,
		WebsiteURL:           cfg.Website.URL,
		GitHubUserID:         githubActorID,
		CredentialVault:      vault,
	}
	gateway, inbox, payloads, err := buildGitHubWebhookRuntime(pool, repository, queue, config)
	if err != nil {
		return nil, err
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
	storyService.ConfigureMayaActor(mayaActorID)
	storyService.ConfigureAutoSchedulingEligibility(compatibility.autoSchedulingEligibility)
	service, err := github.New(
		log,
		repository,
		githubadapter.NewStoryService(storyService),
		compatibility.requestStore,
		attachmentsService,
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("construct GitHub worker service: %w", err)
	}
	if err := service.ValidateWorkerConfiguration(); err != nil {
		return nil, err
	}
	return service, nil
}

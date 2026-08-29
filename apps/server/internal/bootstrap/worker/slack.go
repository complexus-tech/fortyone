package workerbootstrap

import (
	"errors"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/bootstrap/slackadapter"
	integrationrequestsrepository "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository"
	messagingbudget "github.com/complexus-tech/projects-api/internal/modules/messaging/budget"
	messagingcontext "github.com/complexus-tech/projects-api/internal/modules/messaging/context"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	reportsrepository "github.com/complexus-tech/projects-api/internal/modules/reports/repository"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	searchrepository "github.com/complexus-tech/projects-api/internal/modules/search/repository"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	sprintsrepository "github.com/complexus-tech/projects-api/internal/modules/sprints/repository"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	statesrepository "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type slackEventProcessorDependencies struct {
	EventPublisher  *publisher.Publisher
	Tasks           *tasks.Service
	MayaActorID     uuid.UUID
	MayaAccess      mayaWorkspaceAccess
	CredentialVault *credentialvault.Vault
}

func (d slackEventProcessorDependencies) validate() error {
	if d.EventPublisher == nil {
		return errors.New("slack event processor: event publisher is required")
	}
	if d.Tasks == nil {
		return errors.New("slack event processor: tasks service is required")
	}
	if d.MayaActorID == uuid.Nil {
		return errors.New("slack event processor: Maya actor ID is required")
	}
	if d.MayaAccess == nil {
		return errors.New("slack event processor: Maya workspace access is required")
	}
	if d.CredentialVault == nil {
		return errors.New("slack event processor: credential vault is required")
	}
	return nil
}

func buildSlackEventProcessor(log *logger.Logger, pool *pgxpool.Pool, redisClient *redis.Client, cfg Config, dependencies slackEventProcessorDependencies) (*slack.EventProcessor, error) {
	if err := dependencies.validate(); err != nil {
		return nil, err
	}

	messagingRepo := messagingrepository.New(pool)
	slackRepo := slackrepository.New(pool)
	teamsService := teams.New(log, teamsrepository.New(pool))
	statesService := states.New(statesrepository.New(pool))
	usersService := users.New(log, usersrepository.New(pool), nil)
	storiesService := stories.New(
		log,
		storiesrepository.New(log, pool),
		dependencies.EventPublisher,
		dependencies.Tasks,
	)
	storiesService.ConfigureCommentCreator(buildStoryCommentCreator(log, pool))
	storiesService.ConfigureMayaActor(dependencies.MayaActorID)
	storiesService.ConfigureAutoSchedulingEligibility(dependencies.MayaAccess.WorkspaceCanUseMaya)
	searchService := search.New(log, searchrepository.New(pool))
	objectivesService := objectives.New(
		log,
		objectivesrepository.New(pool),
	)
	sprintsService := sprints.New(log, sprintsrepository.New(pool))
	reportsService := reports.New(log, reportsrepository.New(log, pool))
	contextProvider, err := messagingcontext.New(
		usersService,
		workspacesrepository.New(pool),
		teamsService,
	)
	if err != nil {
		return nil, err
	}

	toolExecutor, err := messaging.NewFortyOneToolExecutor(
		teamsService,
		storiesService,
		searchService,
		objectivesService,
		messaging.WithOperationalTools(messaging.OperationalToolServices{
			States:   statesService,
			Users:    usersService,
			Workload: reportsService,
		}),
		messaging.WithPlanningTools(messaging.PlanningToolServices{
			Sprints: sprintsService,
		}),
		messaging.WithStoryMutations(cfg.Messaging.MutationHMACKey),
		messaging.WithStoryMutationConfirmationStore(messagingRepo),
	)
	if err != nil {
		return nil, err
	}

	var assistant messaging.Assistant
	apiKey := strings.TrimSpace(cfg.AIAPIKey)
	if apiKey == "" {
		assistant = messaging.NewUnavailableAssistant("OpenAI is not configured for the worker")
	} else {
		assistant, err = messaging.NewOpenAIAssistant(messaging.OpenAIConfig{
			APIKey: apiKey,
			Model:  strings.TrimSpace(cfg.AIModel),
		}, toolExecutor)
		if err != nil {
			return nil, err
		}
	}
	callLimiter, err := messagingbudget.NewRedisCallLimiter(
		redisClient,
		messagingAssistantCallLimiterConfig(cfg),
	)
	if err != nil {
		return nil, err
	}
	webhookGateway, webhookInbox, err := buildSlackWebhookRuntime(
		pool,
		slackRepo,
		dependencies.Tasks,
		slackWebhookConfig(cfg, dependencies.CredentialVault),
	)
	if err != nil {
		return nil, err
	}
	processorConfig := slackEventProcessorConfig(cfg, dependencies.CredentialVault)
	processorConfig.CallLimiter = slackadapter.NewCallLimiter(callLimiter)
	processorConfig.UsageBudget = slackadapter.NewUsageBudget(messagingrepository.NewDailyUsageRepository(pool))
	processorConfig.ContextProvider = slackadapter.NewContextProvider(contextProvider)
	requestRepository := integrationrequestsrepository.New(pool)
	requestStore := slackadapter.NewRequestStore(requestRepository)
	storyService := slackadapter.NewStoryService(storiesService)
	processorConfig.ThreadSync = requestStore
	processorConfig.StoryReader = storyService
	processorConfig.RequestReader = requestStore
	processorConfig.ObjectiveReader = slackadapter.NewObjectiveReader(objectivesService)
	processorConfig.SprintReader = slackadapter.NewSprintReader(sprintsService)
	processorConfig.MutationConfirmer = slackadapter.NewMutationConfirmer(toolExecutor)
	processorConfig.WebhookInbox = webhookInbox
	processorConfig.WebhookRecovery = webhookGateway

	return slack.NewEventProcessor(
		log,
		slackRepo,
		slackadapter.NewMessagingStore(messagingRepo),
		slackadapter.NewAssistant(assistant),
		workspaceAssistantAccess{access: dependencies.MayaAccess},
		processorConfig,
	)
}

func messagingAssistantCallLimiterConfig(cfg Config) messagingbudget.CallLimiterConfig {
	return messagingbudget.CallLimiterConfig{
		UserLimit:      cfg.MessagingAssistant.UserCallsPerMinute,
		WorkspaceLimit: cfg.MessagingAssistant.WorkspaceCallsPerMinute,
		Window:         time.Minute,
	}
}

func slackEventProcessorConfig(cfg Config, vault *credentialvault.Vault) slack.EventProcessorConfig {
	return slack.EventProcessorConfig{
		WebsiteURL:               cfg.Website.URL,
		WebhookPayloadSecret:     cfg.Slack.WebhookPayloadSecret,
		CredentialVault:          vault,
		ClientID:                 cfg.Slack.ClientID,
		ClientSecret:             cfg.Slack.ClientSecret,
		DailyWorkspaceTokenLimit: cfg.MessagingAssistant.WorkspaceTokensPerDay,
	}
}

func slackWebhookConfig(cfg Config, vault *credentialvault.Vault) slack.Config {
	return slack.Config{
		SigningSecret:        cfg.Slack.SigningSecret,
		WebhookPayloadSecret: cfg.Slack.WebhookPayloadSecret,
		CredentialVault:      vault,
	}
}

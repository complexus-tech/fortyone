package workerbootstrap

import (
	"context"
	"errors"
	"strings"
	"time"

	integrationrequestsrepository "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository"
	mentionsrepository "github.com/complexus-tech/projects-api/internal/modules/mentions/repository"
	messagingbudget "github.com/complexus-tech/projects-api/internal/modules/messaging/budget"
	messagingcontext "github.com/complexus-tech/projects-api/internal/modules/messaging/context"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	okractivitiesrepository "github.com/complexus-tech/projects-api/internal/modules/okractivities/repository"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
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
	"github.com/complexus-tech/projects-api/internal/platform/billing"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type workspaceAssistantAccess struct {
	db *sqlx.DB
}

func (a workspaceAssistantAccess) CanUseAssistant(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	return billing.WorkspaceCanUseMaya(ctx, a.db, workspaceID)
}

type slackEventProcessorDependencies struct {
	EventPublisher *publisher.Publisher
	Tasks          *tasks.Service
}

func (d slackEventProcessorDependencies) validate() error {
	if d.EventPublisher == nil {
		return errors.New("slack event processor: event publisher is required")
	}
	if d.Tasks == nil {
		return errors.New("slack event processor: tasks service is required")
	}
	return nil
}

func buildSlackEventProcessor(log *logger.Logger, db *sqlx.DB, redisClient *redis.Client, cfg Config, dependencies slackEventProcessorDependencies) (*slack.EventProcessor, error) {
	if err := dependencies.validate(); err != nil {
		return nil, err
	}

	messagingRepo := messagingrepository.New(db)
	teamsService := teams.New(log, teamsrepository.New(log, db))
	statesService := states.New(log, statesrepository.New(log, db))
	usersService := users.New(log, usersrepository.New(log, db), nil)
	storiesService := stories.New(
		log,
		storiesrepository.New(log, db),
		mentionsrepository.New(log, db),
		dependencies.EventPublisher,
		dependencies.Tasks,
	)
	searchService := search.New(log, searchrepository.New(log, db))
	okrActivitiesService := okractivities.New(log, okractivitiesrepository.New(log, db))
	objectivesService := objectives.New(
		log,
		objectivesrepository.New(log, db),
		okrActivitiesService,
	)
	sprintsService := sprints.New(log, sprintsrepository.New(log, db))
	contextProvider, err := messagingcontext.New(
		usersService,
		workspacesrepository.New(log, db),
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
			States: statesService,
			Users:  usersService,
		}),
		messaging.WithStoryMutations(cfg.Auth.SecretKey),
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
	processorConfig := slackEventProcessorConfig(cfg, dependencies.Tasks)
	processorConfig.CallLimiter = callLimiter
	processorConfig.UsageBudget = messagingrepository.NewDailyUsageRepository(db)
	processorConfig.ContextProvider = contextProvider
	requestRepository := integrationrequestsrepository.New(log, db)
	processorConfig.ThreadSync = requestRepository
	processorConfig.StoryReader = storiesService
	processorConfig.RequestReader = requestRepository
	processorConfig.ObjectiveReader = objectivesService
	processorConfig.SprintReader = sprintsService
	processorConfig.MutationConfirmer = toolExecutor

	return slack.NewEventProcessor(
		log,
		slackrepository.New(log, db),
		messagingRepo,
		assistant,
		workspaceAssistantAccess{db: db},
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

func slackEventProcessorConfig(cfg Config, eventQueue slack.EventQueue) slack.EventProcessorConfig {
	return slack.EventProcessorConfig{
		WebsiteURL:               cfg.Website.URL,
		SecretKey:                cfg.Auth.SecretKey,
		ClientID:                 cfg.Slack.ClientID,
		ClientSecret:             cfg.Slack.ClientSecret,
		EventQueue:               eventQueue,
		DailyWorkspaceTokenLimit: cfg.MessagingAssistant.WorkspaceTokensPerDay,
	}
}

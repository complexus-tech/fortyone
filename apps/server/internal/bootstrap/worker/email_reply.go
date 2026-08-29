package workerbootstrap

import (
	"fmt"
	"strings"

	emailagent "github.com/complexus-tech/projects-api/internal/modules/emailagent/service"
	emailreplyrepository "github.com/complexus-tech/projects-api/internal/modules/emailreply/repository"
	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	feedbackrepository "github.com/complexus-tech/projects-api/internal/modules/feedback/repository"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	feedbackstory "github.com/complexus-tech/projects-api/internal/modules/feedback/storyadapter"
	keyresultsrepository "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func buildEmailReplyProcessor(
	log *logger.Logger,
	pool *pgxpool.Pool,
	mayaAccess mayaWorkspaceAccess,
	redisClient *redis.Client,
	cfg Config,
	tasksService *tasks.Service,
	mailerService mailer.Service,
	messagingRepo *messagingrepository.Repository,
	inbound *emailreply.Service,
	threads *emailthread.Service,
	mayaActorID uuid.UUID,
) (*emailreply.Processor, error) {
	eventPublisher := publisher.New(redisClient, log)
	storyService := stories.New(
		log,
		storiesrepository.New(log, pool),
		eventPublisher,
		tasksService,
	)
	storyService.ConfigureCommentCreator(buildStoryCommentCreator(log, pool))
	storyService.ConfigureMayaActor(mayaActorID)
	storyService.ConfigureAutoSchedulingEligibility(mayaAccess.WorkspaceCanUseMaya)
	objectiveService := objectives.New(
		log,
		objectivesrepository.New(pool),
		objectives.WithPublisher(eventPublisher),
	)
	keyResultService := keyresults.New(
		log,
		keyresultsrepository.New(pool),
		keyresults.WithPublisher(eventPublisher),
	)
	feedbackService := feedback.New(
		feedbackrepository.New(log, pool),
		feedbackstory.New(storyService),
		feedback.WithEventPublisher(log, eventPublisher),
	)

	contextLoader, err := emailreply.NewDBContextLoader(emailreplyrepository.New(pool))
	if err != nil {
		return nil, fmt.Errorf("initialize Maya email reply context: %w", err)
	}
	mutations, err := emailreply.NewDomainMutationApplier(
		objectiveService,
		keyResultService,
		storyService,
		feedbackService,
		contextLoader,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize Maya email reply mutations: %w", err)
	}

	openAIConfig := emailagent.OpenAIConfig{
		APIKey: strings.TrimSpace(cfg.AIAPIKey),
		Model:  strings.TrimSpace(cfg.AIModel),
	}
	generator, err := emailagent.NewOpenAIGenerator(openAIConfig)
	if err != nil {
		return nil, fmt.Errorf("initialize Maya email reply generator: %w", err)
	}
	decisionService, err := emailagent.New(generator)
	if err != nil {
		return nil, fmt.Errorf("initialize Maya email reply agent: %w", err)
	}
	summarizer, err := emailagent.NewOpenAISummarizer(openAIConfig)
	if err != nil {
		return nil, fmt.Errorf("initialize Maya email conversation summarizer: %w", err)
	}

	processor, err := emailreply.NewProcessor(emailreply.ProcessorConfig{
		Log:        log,
		Store:      emailReplyStoreAdapter{repository: messagingRepo},
		Inbound:    inbound,
		Agent:      emailDecisionAdapter{backend: decisionService},
		Summarizer: emailSummaryAdapter{backend: summarizer},
		Renderer:   emailCopyRenderer{},
		Threads:    emailReplyThreadAdapter{service: threads},
		Mailer:     mailerService,
		Context:    contextLoader,
		Mutations:  mutations,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize Maya email reply processor: %w", err)
	}
	return processor, nil
}

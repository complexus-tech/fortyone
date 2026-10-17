package workerbootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/bootstrap/agentadapter"
	"github.com/complexus-tech/projects-api/internal/bootstrap/chatadapter"
	agentrunsrepository "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository"
	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatrepository "github.com/complexus-tech/projects-api/internal/modules/chat/repository"
	chat "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	documentsrepository "github.com/complexus-tech/projects-api/internal/modules/documents/repository"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	messagingbudget "github.com/complexus-tech/projects-api/internal/modules/messaging/budget"
	messagingcontext "github.com/complexus-tech/projects-api/internal/modules/messaging/context"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	searchrepository "github.com/complexus-tech/projects-api/internal/modules/search/repository"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	statesrepository "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func registerAgentInvocationWorker(mux *asynq.ServeMux, scheduler scheduleRegistrar, log *logger.Logger,
	pool *pgxpool.Pool, redisClient *redis.Client, cfg Config, mayaID uuid.UUID, access mayaWorkspaceAccess,
	eventPublisher *publisher.Publisher, taskService *tasks.Service, notificationService *notifications.Service) error {
	storyService := stories.New(log, storiesrepository.New(log, pool), eventPublisher, taskService)
	storyService.ConfigureAutoSchedulingEligibility(access.WorkspaceCanUseMaya)
	storyService.ConfigureMayaActor(mayaID)
	teamService := teams.New(log, teamsrepository.New(pool))
	stateService := states.New(statesrepository.New(pool))
	userService := users.New(log, usersrepository.New(pool), nil)
	contextProvider, err := messagingcontext.New(userService, workspacesrepository.New(pool), teamService)
	if err != nil {
		return err
	}
	tools, err := messaging.NewFortyOneToolExecutor(teamService, storyService,
		search.New(log, searchrepository.New(pool)), objectives.New(log, objectivesrepository.New(pool)),
		messaging.WithOperationalTools(messaging.OperationalToolServices{States: stateService, Users: userService}),
		messaging.WithStoryMutations(cfg.Messaging.MutationHMACKey),
		messaging.WithStoryMutationConfirmationStore(messagingrepository.New(pool)))
	if err != nil {
		return err
	}
	var assistant messaging.Assistant = messaging.NewUnavailableAssistant("Maya is not configured")
	configured := strings.TrimSpace(cfg.AIAPIKey) != ""
	if configured {
		assistant, err = messaging.NewOpenAIAssistant(messaging.OpenAIConfig{APIKey: strings.TrimSpace(cfg.AIAPIKey), Model: strings.TrimSpace(cfg.AIModel)}, tools)
		if err != nil {
			return err
		}
	}
	limiter, err := messagingbudget.NewRedisCallLimiter(redisClient, messagingAssistantCallLimiterConfig(cfg))
	if err != nil {
		return err
	}
	runtime := &agentadapter.CommentRuntime{
		Stories: storyService, Comments: comments.New(commentsrepository.New(log, pool)), Assistant: assistant, Tools: tools,
		Memberships: workspacesrepository.New(pool),
		Runs:        agentruns.New(agentrunsrepository.New(pool), nil),
		Context:     contextProvider, Access: access, MayaID: mayaID, WebsiteURL: cfg.Website.URL, Enabled: configured,
		Limiter: limiter, Usage: messagingrepository.NewDailyUsageRepository(pool),
		DailyTokenLimit:   cfg.MessagingAssistant.WorkspaceTokensPerDay,
		NotificationRules: notifications.NewRules(log, storyService, userService, stateService), Notifications: notificationService,
	}
	documentService := documents.New(documentsrepository.New(pool))
	var documentAssistant messaging.DocumentAssistant
	if provider, ok := assistant.(messaging.DocumentAssistant); ok {
		documentAssistant = provider
	}
	documentRuntime := &agentadapter.DocumentRuntime{Documents: documentService, Assistant: documentAssistant, Shared: runtime, Runs: runtime.Runs, EditsAvailable: documentService.AgentEditsAvailable}
	chatAttachments := chatadapter.NewAttachmentRetirer(pool, cfg.Storage.Provider, cfg.Storage.AttachmentsBucket)
	chatRepository := chatrepository.New(log, pool, chatrepository.WithAttachmentRetirement(chatAttachments.Retire))
	chatService := chat.New(chatRepository, chat.WithWorkspaceSettings(chatrollout.New(workspacesrepository.New(pool))), chat.WithMayaActorID(mayaID), chat.WithAfterCommit(func(ctx context.Context) {
		if err := taskService.EnqueueChatDelivery(ctx); err != nil {
			log.Error(ctx, "enqueue chat delivery", "error", err)
		}
	}))
	chatRuntime := &agentadapter.ChatRuntime{Rollout: chatrollout.New(workspacesrepository.New(pool)), Chats: chatService, Shared: runtime, Tools: tools, Log: log}
	service := agentruns.New(agentrunsrepository.New(pool), &agentadapter.RuntimeRouter{Story: runtime, Document: documentRuntime, Chat: chatRuntime})
	mux.HandleFunc(tasks.TypeAgentInvocation, func(ctx context.Context, task *asynq.Task) error {
		var payload tasks.AgentInvocationPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil || payload.RunID == uuid.Nil {
			return asynq.SkipRetry
		}
		return service.Process(ctx, payload.RunID)
	})
	mux.HandleFunc(tasks.TypeAgentInvocationRecovery, func(ctx context.Context, _ *asynq.Task) error {
		_, err := service.ProcessPending(ctx, 3)
		return err
	})
	_, err = scheduler.Register("@every 10s", asynq.NewTask(tasks.TypeAgentInvocationRecovery, nil),
		asynq.Queue("default"), asynq.MaxRetry(1), asynq.Timeout(5*time.Minute), asynq.Unique(9*time.Second))
	if err != nil {
		return fmt.Errorf("register Maya invocation recovery: %w", err)
	}
	return nil
}

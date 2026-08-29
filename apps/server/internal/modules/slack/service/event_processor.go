package slack

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"strings"
	"time"

	platformclock "github.com/complexus-tech/projects-api/internal/platform/clock"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

const (
	slackConversationHistoryLimit = 20
	slackThreadSubscriptionTTL    = 30 * 24 * time.Hour
	slackMessageTextLimit         = 3900
	slackStateWriteTimeout        = 5 * time.Second

	assistantMessageTooLargeReply = "That message is too long for Maya. Please shorten it and try again."
	assistantUserRateLimitReply   = "You've reached Maya's per-minute message limit. Please wait a minute and try again."
	assistantWorkspaceRateReply   = "Your workspace is sending too many requests to Maya right now. Please wait a minute and try again."
	assistantDailyLimitReply      = "Your workspace has reached today's Maya usage limit. Please try again tomorrow or contact your workspace administrator."
	assistantConfigurationReply   = "Maya is temporarily unavailable because of an assistant configuration issue. Please contact your FortyOne workspace administrator."
)

var errSlackInstallationChanged = errors.New("slack installation changed while work was in progress")

type SlackEventRepository interface {
	GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackWorkspaceRecord, error)
	GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackWorkspaceRecord, error)
	FindWorkspaceByID(ctx context.Context, workspaceID uuid.UUID) (workspaceRecord, error)
	FindLinkedUserIDBySlackUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (*uuid.UUID, error)
	DeactivateSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string, installGeneration uuid.UUID) error
	ClaimRecoverableSlackUninstalls(ctx context.Context, limit int) ([]slackUninstallRecord, error)
	CompleteSlackUninstall(ctx context.Context, id uuid.UUID, message string) error
	FailSlackUninstall(ctx context.Context, id uuid.UUID, message string, nextAttemptAt *time.Time) error
}

type SlackEventStore interface {
	CreateNonce(ctx context.Context, input nonceInput) error
	UpsertConversation(ctx context.Context, input conversationInput) (uuid.UUID, error)
	FindConversation(ctx context.Context, input conversationInput) (conversationRecord, error)
	AppendMessage(ctx context.Context, conversationID uuid.UUID, externalMessageID, role, content string) error
	ListRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]messageRecord, error)
	StartOutboundDelivery(ctx context.Context, input outboundDeliveryInput) (record outboundDeliveryRecord, claimed bool, err error)
	SetOutboundDeliveryContent(ctx context.Context, id uuid.UUID, content string) error
	SetOutboundDeliveryContentAndDestination(ctx context.Context, id uuid.UUID, content, externalChannelID, externalThreadID string) error
	CompleteOutboundDelivery(ctx context.Context, id uuid.UUID, externalMessageID string) error
	FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
	CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
}

type slackProviderPayloadStore interface {
	SetOutboundDeliveryContentAndProviderPayload(ctx context.Context, id uuid.UUID, content string, providerPayload []byte) error
}

type slackOutboundContentStore interface {
	SetOutboundDeliveryContent(ctx context.Context, id uuid.UUID, content string) error
}

type slackChannelConversationFinder interface {
	FindChannelConversation(ctx context.Context, input conversationInput) (conversationRecord, error)
}

type slackConversationFinder interface {
	FindConversation(ctx context.Context, input conversationInput) (conversationRecord, error)
}

type recoverableSlackDeliveryStore interface {
	ListRecoverableOutboundDeliveries(ctx context.Context, provider string, limit int) ([]outboundDeliveryRecord, error)
}

type SlackWebhookInbox interface {
	GetByID(ctx context.Context, id uuid.UUID) (webhooks.Record, error)
	GetByExternalKey(ctx context.Context, provider integrations.ProviderKey, externalAccountID, deliveryID string) (webhooks.Record, error)
	Start(ctx context.Context, id uuid.UUID, now time.Time, lease time.Duration) (webhooks.Record, bool, error)
	Complete(ctx context.Context, id uuid.UUID, status webhooks.Status, outcomeCode string, completedAt time.Time) error
}

type SlackWebhookRecovery interface {
	Recover(ctx context.Context, provider integrations.ProviderKey, policy webhooks.RecoveryPolicy) (webhooks.RecoveryReport, error)
}

type outboundDeliveryStateStore interface {
	FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
	CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
}

type SlackOutboundMessage struct {
	ChannelID        string
	UserID           string
	ThreadTS         string
	Text             string
	ClientMessageID  string
	Ephemeral        bool
	StandardMarkdown bool
	ProviderPayload  SlackProviderPayload
}

type SlackMessageSender interface {
	Send(ctx context.Context, botToken string, message SlackOutboundMessage) (externalMessageID string, err error)
}

type AssistantAccessChecker interface {
	CanUseAssistant(ctx context.Context, workspaceID uuid.UUID) (bool, error)
}

type AssistantCallLimiter interface {
	Admit(ctx context.Context, input AssistantAdmissionInput) (AssistantAdmissionDecision, error)
}

type AssistantUsageBudget interface {
	Check(ctx context.Context, workspaceID uuid.UUID, limit int64) (dailyUsageSnapshot, error)
	Record(ctx context.Context, input dailyUsageRecordInput, limit int64) (dailyUsageSnapshot, error)
}

// AssistantContextProvider loads descriptive, server-authoritative context for
// one assistant turn. Authorization remains in the separately supplied
// workspace, actor, allowed-team scope, and shared-team scope on the request.
type AssistantContextProvider interface {
	Load(
		ctx context.Context,
		workspaceID, userID uuid.UUID,
		allowedTeamIDs []uuid.UUID,
		surface assistantRuntimeSurface,
		now time.Time,
	) (*assistantRuntimeContext, error)
}

type EventProcessorConfig struct {
	WebsiteURL               string
	WebhookPayloadSecret     string
	CredentialVault          CredentialVault
	ClientID                 string
	ClientSecret             string
	CallLimiter              AssistantCallLimiter
	UsageBudget              AssistantUsageBudget
	ContextProvider          AssistantContextProvider
	DailyWorkspaceTokenLimit int64
	ThreadSync               SlackThreadSync
	StoryReader              SlackStoryReader
	RequestReader            SlackRequestReader
	ObjectiveReader          SlackObjectiveReader
	SprintReader             SlackSprintReader
	MutationConfirmer        storyMutationConfirmer
	WebhookInbox             SlackWebhookInbox
	WebhookRecovery          SlackWebhookRecovery
}

type EventProcessor struct {
	log                      *logger.Logger
	repo                     SlackEventRepository
	store                    SlackEventStore
	assistant                assistant
	access                   AssistantAccessChecker
	callLimiter              AssistantCallLimiter
	usageBudget              AssistantUsageBudget
	contextProvider          AssistantContextProvider
	sender                   SlackMessageSender
	statusSetter             SlackAssistantStatusSetter
	webClient                *slackWebClient
	codec                    *credentialCodec
	webhookPayloads          slackWebhookPayloadCodec
	website                  string
	clientID                 string
	clientSecret             string
	random                   io.Reader
	clock                    Clock
	webhookInbox             SlackWebhookInbox
	webhookRecovery          SlackWebhookRecovery
	dailyWorkspaceTokenLimit int64
	threadSync               SlackThreadSync
	storyReader              SlackStoryReader
	requestReader            SlackRequestReader
	objectiveReader          SlackObjectiveReader
	sprintReader             SlackSprintReader
	mutationConfirmer        storyMutationConfirmer
	workObjects              *slackWorkObjectPublisher
}

func NewEventProcessor(
	log *logger.Logger,
	repo SlackEventRepository,
	store SlackEventStore,
	assistant assistant,
	access AssistantAccessChecker,
	cfg EventProcessorConfig,
) (*EventProcessor, error) {
	if repo == nil {
		return nil, errors.New("slack event repository is required")
	}
	if store == nil {
		return nil, errors.New("slack event store is required")
	}
	if assistant == nil {
		return nil, errors.New("slack assistant is required")
	}
	if access == nil {
		return nil, errors.New("slack assistant access checker is required")
	}
	if cfg.CallLimiter == nil {
		return nil, errors.New("slack assistant call limiter is required")
	}
	if cfg.UsageBudget == nil {
		return nil, errors.New("slack assistant usage budget is required")
	}
	if cfg.ContextProvider == nil {
		return nil, errors.New("slack assistant context provider is required")
	}
	if cfg.WebhookInbox == nil {
		return nil, errors.New("slack webhook inbox is required")
	}
	if cfg.WebhookRecovery == nil {
		return nil, errors.New("slack webhook recovery gateway is required")
	}
	if cfg.DailyWorkspaceTokenLimit < 0 {
		return nil, errors.New("slack assistant daily workspace token limit cannot be negative")
	}
	codec, err := newCredentialCodec(cfg.CredentialVault)
	if err != nil {
		return nil, err
	}
	webhookPayloads, err := newSlackWebhookPayloadCodec(cfg.WebhookPayloadSecret)
	if err != nil {
		return nil, err
	}
	webClient := newSlackWebClient(nil)
	return &EventProcessor{
		log:                      log,
		repo:                     repo,
		store:                    store,
		assistant:                assistant,
		access:                   access,
		callLimiter:              cfg.CallLimiter,
		usageBudget:              cfg.UsageBudget,
		contextProvider:          cfg.ContextProvider,
		sender:                   &slackAPISender{client: webClient},
		statusSetter:             &slackAssistantStatusSetter{client: webClient},
		webClient:                webClient,
		codec:                    codec,
		webhookPayloads:          webhookPayloads,
		website:                  strings.TrimRight(strings.TrimSpace(cfg.WebsiteURL), "/"),
		clientID:                 strings.TrimSpace(cfg.ClientID),
		clientSecret:             strings.TrimSpace(cfg.ClientSecret),
		random:                   rand.Reader,
		clock:                    platformclock.System{},
		webhookInbox:             cfg.WebhookInbox,
		webhookRecovery:          cfg.WebhookRecovery,
		dailyWorkspaceTokenLimit: cfg.DailyWorkspaceTokenLimit,
		threadSync:               cfg.ThreadSync,
		storyReader:              cfg.StoryReader,
		requestReader:            cfg.RequestReader,
		objectiveReader:          cfg.ObjectiveReader,
		sprintReader:             cfg.SprintReader,
		mutationConfirmer:        cfg.MutationConfirmer,
		workObjects:              newSlackWorkObjectPublisher(webClient),
	}, nil
}

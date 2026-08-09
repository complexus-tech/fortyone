package slack

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	messagingbudget "github.com/complexus-tech/projects-api/internal/modules/messaging/budget"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

const (
	slackConversationHistoryLimit = 20
	slackThreadSubscriptionTTL    = 30 * 24 * time.Hour
	slackMessageTextLimit         = 3900
	legacyCredentialBatchSize     = 100
	slackStateWriteTimeout        = 5 * time.Second

	assistantMessageTooLargeReply = "That message is too long for Maya. Please shorten it and try again."
	assistantUserRateLimitReply   = "You've reached Maya's per-minute message limit. Please wait a minute and try again."
	assistantWorkspaceRateReply   = "Your workspace is sending too many requests to Maya right now. Please wait a minute and try again."
	assistantDailyLimitReply      = "Your workspace has reached today's Maya usage limit. Please try again tomorrow or contact your workspace administrator."
	assistantConfigurationReply   = "Maya is temporarily unavailable because of an assistant configuration issue. Please contact your FortyOne workspace administrator."
)

var errSlackInstallationChanged = errors.New("Slack installation changed while work was in progress")

type legacyCredentialRepository interface {
	ListLegacySlackCredentials(ctx context.Context, limit int) ([]slackrepository.LegacySlackCredentialRecord, error)
	ScrubVersionedLegacySlackCredentials(ctx context.Context, limit int) (int, error)
}

type SlackEventRepository interface {
	GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackWorkspaceRecord, error)
	GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackrepository.SlackWorkspaceRecord, error)
	FindWorkspaceByID(ctx context.Context, workspaceID uuid.UUID) (slackrepository.WorkspaceRecord, error)
	FindLinkedUserIDBySlackUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (*uuid.UUID, error)
	UpgradeSlackCredential(ctx context.Context, slackWorkspaceID uuid.UUID, encrypted string, version int) error
	DeactivateSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string, installGeneration uuid.UUID) error
	ClaimRecoverableSlackUninstalls(ctx context.Context, limit int) ([]slackrepository.SlackUninstallRecord, error)
	CompleteSlackUninstall(ctx context.Context, id uuid.UUID, message string) error
	FailSlackUninstall(ctx context.Context, id uuid.UUID, message string, nextAttemptAt *time.Time) error
}

type SlackEventStore interface {
	CreateNonce(ctx context.Context, input messagingrepository.NonceInput) error
	GetInboundEvent(ctx context.Context, provider, externalWorkspaceID, externalEventID string) (messagingrepository.InboundEventRecord, error)
	StartInboundEvent(ctx context.Context, provider, externalWorkspaceID, externalEventID string) (record messagingrepository.InboundEventRecord, claimed bool, err error)
	CompleteInboundEvent(ctx context.Context, id uuid.UUID, status, message string) error
	UpsertConversation(ctx context.Context, input messagingrepository.ConversationInput) (uuid.UUID, error)
	FindConversation(ctx context.Context, input messagingrepository.ConversationInput) (messagingrepository.ConversationRecord, error)
	AppendMessage(ctx context.Context, conversationID uuid.UUID, externalMessageID, role, content string) error
	ListRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]messagingrepository.MessageRecord, error)
	StartOutboundDelivery(ctx context.Context, input messagingrepository.OutboundDeliveryInput) (record messagingrepository.OutboundDeliveryRecord, claimed bool, err error)
	SetOutboundDeliveryContent(ctx context.Context, id uuid.UUID, content string) error
	SetOutboundDeliveryContentAndDestination(ctx context.Context, id uuid.UUID, content, externalChannelID, externalThreadID string) error
	CompleteOutboundDelivery(ctx context.Context, id uuid.UUID, externalMessageID string) error
	FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
	CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
}

type recoverableSlackEventStore interface {
	ClaimRecoverableInboundEvents(ctx context.Context, provider string, limit int) ([]messagingrepository.InboundEventRecord, error)
	ReleaseInboundEventRecovery(ctx context.Context, id uuid.UUID, generation int) error
	ListRecoverableOutboundDeliveries(ctx context.Context, provider string, limit int) ([]messagingrepository.OutboundDeliveryRecord, error)
}

type outboundDeliveryStateStore interface {
	FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
	CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
}

type SlackOutboundMessage struct {
	ChannelID       string
	UserID          string
	ThreadTS        string
	Text            string
	ClientMessageID string
	Ephemeral       bool
}

type SlackMessageSender interface {
	Send(ctx context.Context, botToken string, message SlackOutboundMessage) (externalMessageID string, err error)
}

type AssistantAccessChecker interface {
	CanUseAssistant(ctx context.Context, workspaceID uuid.UUID) (bool, error)
}

type AssistantCallLimiter interface {
	Admit(ctx context.Context, input messagingbudget.AdmissionInput) (messagingbudget.AdmissionDecision, error)
}

type AssistantUsageBudget interface {
	Check(ctx context.Context, workspaceID uuid.UUID, limit int64) (messagingrepository.DailyUsageSnapshot, error)
	Record(ctx context.Context, input messagingrepository.DailyUsageRecordInput, limit int64) (messagingrepository.DailyUsageSnapshot, error)
}

type EventProcessorConfig struct {
	WebsiteURL               string
	SecretKey                string
	ClientID                 string
	ClientSecret             string
	EventQueue               EventQueue
	CallLimiter              AssistantCallLimiter
	UsageBudget              AssistantUsageBudget
	DailyWorkspaceTokenLimit int64
}

type EventProcessor struct {
	log                      *logger.Logger
	repo                     SlackEventRepository
	store                    SlackEventStore
	assistant                messaging.Assistant
	access                   AssistantAccessChecker
	callLimiter              AssistantCallLimiter
	usageBudget              AssistantUsageBudget
	sender                   SlackMessageSender
	webClient                *slackWebClient
	codec                    *credentialCodec
	website                  string
	clientID                 string
	clientSecret             string
	random                   io.Reader
	clock                    Clock
	eventQueue               EventQueue
	dailyWorkspaceTokenLimit int64
}

func NewEventProcessor(
	log *logger.Logger,
	repo SlackEventRepository,
	store SlackEventStore,
	assistant messaging.Assistant,
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
	if cfg.DailyWorkspaceTokenLimit < 0 {
		return nil, errors.New("slack assistant daily workspace token limit cannot be negative")
	}
	codec, err := newCredentialCodec(cfg.SecretKey)
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
		sender:                   &slackAPISender{client: webClient},
		webClient:                webClient,
		codec:                    codec,
		website:                  strings.TrimRight(strings.TrimSpace(cfg.WebsiteURL), "/"),
		clientID:                 strings.TrimSpace(cfg.ClientID),
		clientSecret:             strings.TrimSpace(cfg.ClientSecret),
		random:                   rand.Reader,
		clock:                    realClock{},
		eventQueue:               cfg.EventQueue,
		dailyWorkspaceTokenLimit: cfg.DailyWorkspaceTokenLimit,
	}, nil
}

// ProcessEvent loads the encrypted canonical payload from the durable inbox.
// Queue backends receive only the provider event ID, never message content.
func (p *EventProcessor) ProcessEvent(ctx context.Context, externalWorkspaceID, eventID string) error {
	externalWorkspaceID = strings.TrimSpace(externalWorkspaceID)
	eventID = strings.TrimSpace(eventID)
	if externalWorkspaceID == "" || eventID == "" {
		return errors.New("Slack external workspace id and event id are required")
	}
	record, err := p.store.GetInboundEvent(ctx, "slack", externalWorkspaceID, eventID)
	if err != nil {
		return err
	}
	if record.Status == "completed" || record.Status == "ignored" || record.Status == "cancelled" {
		return nil
	}
	if record.PayloadEncrypted == nil || strings.TrimSpace(*record.PayloadEncrypted) == "" {
		return p.failUnreadableEvent(ctx, externalWorkspaceID, eventID, errors.New("Slack inbox event has no encrypted payload"))
	}
	body, err := p.codec.openPayload(*record.PayloadEncrypted)
	if err != nil {
		return p.failUnreadableEvent(ctx, externalWorkspaceID, eventID, err)
	}
	envelope, err := decodeSlackEvent(body)
	if err != nil {
		return p.failUnreadableEvent(ctx, externalWorkspaceID, eventID, err)
	}
	if envelope.TeamID != externalWorkspaceID || envelope.EventID != eventID {
		return p.failUnreadableEvent(ctx, externalWorkspaceID, eventID, errors.New("Slack inbox payload does not match its workspace and event id"))
	}
	return p.Process(ctx, body)
}

func (p *EventProcessor) failUnreadableEvent(ctx context.Context, externalWorkspaceID, eventID string, cause error) error {
	receipt, claimed, err := p.store.StartInboundEvent(ctx, "slack", externalWorkspaceID, eventID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	if err := p.store.CompleteInboundEvent(context.WithoutCancel(ctx), receipt.ID, "failed", truncateError(cause)); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func (p *EventProcessor) Process(ctx context.Context, rawBody []byte) (err error) {
	envelope, err := decodeSlackEvent(rawBody)
	if err != nil {
		return err
	}
	if envelope.EventID == "" {
		return errors.New("slack event id is required")
	}
	receipt, process, err := p.store.StartInboundEvent(ctx, "slack", envelope.TeamID, envelope.EventID)
	if err != nil {
		return err
	}
	if !process {
		return nil
	}
	status := "failed"
	statusMessage := ""
	defer func() {
		if err != nil {
			statusMessage = truncateError(err)
		}
		if completeErr := p.store.CompleteInboundEvent(context.WithoutCancel(ctx), receipt.ID, status, statusMessage); completeErr != nil {
			if p.log != nil {
				p.log.Error(context.WithoutCancel(ctx), "failed updating Slack event receipt", "error", completeErr, "event_id", envelope.EventID)
			}
			if err == nil {
				err = completeErr
			}
		}
	}()

	event, supported := normalizeSlackEvent(envelope)
	if !supported {
		status = "ignored"
		return nil
	}

	installation, err := p.repo.GetSlackWorkspaceByTeamID(ctx, event.TeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			status = "ignored"
			return nil
		}
		return err
	}
	if !inboundReceiptMatchesInstallation(receipt, installation) {
		status = "ignored"
		return nil
	}
	if isSlackLifecycleEvent(event.Kind) && !slackLifecycleEventIsCurrent(envelope.EventTime, installation.AuthorizedAt) {
		status = "ignored"
		return nil
	}
	if event.Kind == slackEventKindUninstalled {
		if deactivateErr := p.repo.DeactivateSlackWorkspaceByTeamID(ctx, event.TeamID, installation.InstallGeneration); deactivateErr != nil && !slackrepository.IsNotFound(deactivateErr) {
			return deactivateErr
		}
		status = "completed"
		return nil
	}
	if event.Kind == slackEventKindRevoked {
		if installationBotTokenRevoked(installation, event.RevokedBotUserIDs) {
			if deactivateErr := p.repo.DeactivateSlackWorkspaceByTeamID(ctx, event.TeamID, installation.InstallGeneration); deactivateErr != nil && !slackrepository.IsNotFound(deactivateErr) {
				return deactivateErr
			}
		}
		status = "completed"
		return nil
	}
	workspace, err := p.repo.FindWorkspaceByID(ctx, installation.WorkspaceID)
	if err != nil {
		return err
	}
	botToken, err := p.botToken(ctx, installation)
	if err != nil {
		return err
	}
	linkedUserID, err := p.repo.FindLinkedUserIDBySlackUser(ctx, workspace.ID, event.TeamID, event.UserID)
	if err != nil {
		return err
	}
	if event.Kind == slackEventKindChannelThread {
		// An explicit app_mention event owns messages that mention the bot. Slack
		// can also emit the same Slack message through message.channels/groups;
		// ignoring that broad event prevents two answers with different event IDs.
		if installation.BotUserID != nil && containsSlackUserMention(event.Text, *installation.BotUserID) {
			status = "ignored"
			return nil
		}
		if linkedUserID == nil || *linkedUserID == uuid.Nil {
			status = "ignored"
			return nil
		}
		subscribed, subscriptionErr := p.channelThreadIsSubscribed(ctx, workspace.ID, *linkedUserID, installation, event)
		if subscriptionErr != nil {
			return subscriptionErr
		}
		if !subscribed {
			status = "ignored"
			return nil
		}
	}
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		text, linkErr := p.accountLinkMessage(ctx, workspace, event)
		if linkErr != nil {
			return linkErr
		}
		if err := p.deliver(ctx, receipt.ID, workspace.ID, installation.InstallGeneration, nil, event, botToken, "link", text); err != nil {
			return err
		}
		status = "completed"
		return nil
	}

	allowed, err := p.access.CanUseAssistant(ctx, workspace.ID)
	if err != nil {
		return err
	}
	if !allowed {
		if err := p.deliver(ctx, receipt.ID, workspace.ID, installation.InstallGeneration, linkedUserID, event, botToken, "access", "Maya is available on FortyOne paid plans and active trials."); err != nil {
			return err
		}
		status = "completed"
		return nil
	}

	prompt := event.Text
	if installation.BotUserID != nil {
		prompt = removeBotMention(prompt, *installation.BotUserID)
	}
	if prompt == "" {
		prompt = "How can you help me with my FortyOne workspace?"
	}
	if validateErr := messaging.ValidatePrompt(prompt); validateErr != nil {
		if !errors.Is(validateErr, messaging.ErrMessageTooLarge) {
			return validateErr
		}
		if err := p.deliver(ctx, receipt.ID, workspace.ID, installation.InstallGeneration, linkedUserID, event, botToken, "assistant-input-too-large", assistantMessageTooLargeReply); err != nil {
			return err
		}
		status = "completed"
		return nil
	}
	deliveryKey := event.EventID + ":assistant"
	deliveryChannelID := event.ChannelID
	deliveryThreadTS := event.ReplyTS
	deliveryExpiresAt := p.clock.Now().UTC().Add(time.Hour)
	delivery, shouldDeliver, err := p.store.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
		Provider:                "slack",
		WorkspaceID:             workspace.ID,
		UserID:                  linkedUserID,
		InstallGeneration:       &installation.InstallGeneration,
		ExternalWorkspaceID:     event.TeamID,
		ExternalRecipientUserID: event.UserID,
		InboundEventID:          &receipt.ID,
		IdempotencyKey:          deliveryKey,
		ExternalChannelID:       deliveryChannelID,
		ExternalThreadID:        deliveryThreadTS,
		Purpose:                 "assistant",
		ExpiresAt:               &deliveryExpiresAt,
	})
	if err != nil {
		return err
	}
	if delivery.ExpiresAt != nil {
		deliveryExpiresAt = delivery.ExpiresAt.UTC()
	}
	if strings.TrimSpace(delivery.ExternalChannelID) != "" {
		deliveryChannelID = strings.TrimSpace(delivery.ExternalChannelID)
	}
	if delivery.ExternalThreadID != nil {
		deliveryThreadTS = strings.TrimSpace(*delivery.ExternalThreadID)
	} else if deliveryChannelID != event.ChannelID {
		deliveryThreadTS = ""
	}
	if !shouldDeliver {
		if delivery.Status == "delivered" && delivery.Content != nil && delivery.ExternalMessageID != nil {
			content := strings.TrimSpace(*delivery.Content)
			if !isAssistantBudgetNotice(content) {
				conversationID, conversationErr := p.persistAssistantPrompt(ctx, workspace.ID, *linkedUserID, event, prompt)
				if conversationErr != nil {
					return conversationErr
				}
				if err := p.store.AppendMessage(ctx, conversationID, *delivery.ExternalMessageID, "assistant", content); err != nil {
					return err
				}
			}
		}
		status = "completed"
		return nil
	}

	reply := ""
	if delivery.Content != nil {
		reply = strings.TrimSpace(*delivery.Content)
	}
	if !p.clock.Now().UTC().Before(deliveryExpiresAt) {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack assistant delivery expired before send"); cancelErr != nil {
			return cancelErr
		}
		status = "ignored"
		return nil
	}
	contentPersisted := reply != ""
	persistConversation := !isAssistantBudgetNotice(reply)
	if reply == "" {
		if _, err := p.usageBudget.Check(ctx, workspace.ID, p.dailyWorkspaceTokenLimit); err != nil {
			if !errors.Is(err, messagingrepository.ErrDailyWorkspaceTokenLimit) {
				if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
					return errors.Join(err, failErr)
				}
				return err
			}
			reply = assistantDailyLimitReply
			persistConversation = false
		}
	}
	if reply == "" {
		admission, admissionErr := p.callLimiter.Admit(ctx, messagingbudget.AdmissionInput{
			Provider:            "slack",
			WorkspaceID:         workspace.ID,
			UserID:              *linkedUserID,
			ExternalWorkspaceID: event.TeamID,
			ExternalEventID:     event.EventID,
		})
		if admissionErr != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(admissionErr)); failErr != nil {
				return errors.Join(admissionErr, failErr)
			}
			return admissionErr
		}
		if !admission.Allowed {
			reply = assistantWorkspaceRateReply
			if admission.LimitedScope == "user" {
				reply = assistantUserRateLimitReply
			}
			persistConversation = false
		}
	}
	if event.Kind != slackEventKindDirect && isAssistantBudgetNotice(reply) && deliveryChannelID != event.UserID {
		if err := p.store.SetOutboundDeliveryContentAndDestination(ctx, delivery.ID, reply, event.UserID, ""); err != nil {
			return err
		}
		deliveryChannelID = event.UserID
		deliveryThreadTS = ""
		contentPersisted = true
	}

	conversationID := uuid.Nil
	if persistConversation {
		conversationID, err = p.persistAssistantPrompt(ctx, workspace.ID, *linkedUserID, event, prompt)
		if err != nil {
			return err
		}
	}
	if reply == "" {
		history, historyErr := p.store.ListRecentMessages(ctx, conversationID, slackConversationHistoryLimit)
		if historyErr != nil {
			return historyErr
		}
		turns := make([]messaging.ConversationTurn, 0, len(history))
		for _, message := range history {
			if message.ExternalMessageID != nil && *message.ExternalMessageID == event.MessageTS && message.Role == "user" {
				continue
			}
			role := messaging.RoleUser
			if message.Role == "assistant" {
				role = messaging.RoleAssistant
			}
			turns = append(turns, messaging.ConversationTurn{Role: role, Text: message.Content})
		}
		response, responseErr := p.assistant.Respond(ctx, messaging.Request{
			WorkspaceID:  workspace.ID,
			UserID:       *linkedUserID,
			Conversation: turns,
			Prompt:       prompt,
		})
		_, usageErr := p.recordAssistantUsage(ctx, messagingrepository.DailyUsageRecordInput{
			InboundEventID:      receipt.ID,
			WorkspaceID:         workspace.ID,
			Provider:            "slack",
			ExternalWorkspaceID: event.TeamID,
			ExternalEventID:     event.EventID,
			AttemptCount:        receipt.AttemptCount,
			Usage:               response.Usage,
		})
		if usageErr != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(usageErr)); failErr != nil {
				return errors.Join(responseErr, usageErr, failErr)
			}
			return errors.Join(responseErr, usageErr)
		}
		if responseErr != nil {
			if errors.Is(responseErr, messaging.ErrAssistantNotConfigured) || messaging.IsPermanentOpenAIError(responseErr) {
				reply = assistantConfigurationReply
			} else {
				if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(responseErr)); failErr != nil {
					return errors.Join(responseErr, failErr)
				}
				return responseErr
			}
		} else {
			reply = truncateSlackText(response.Text)
		}
		if reply == "" {
			reply = "I couldn't generate a useful response. Please try again."
		}
	}
	if !contentPersisted {
		if err := p.store.SetOutboundDeliveryContent(ctx, delivery.ID, reply); err != nil {
			return err
		}
	}
	if !p.clock.Now().UTC().Before(deliveryExpiresAt) {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack assistant delivery expired before send"); cancelErr != nil {
			return cancelErr
		}
		status = "ignored"
		return nil
	}
	currentLinkedUserID, linkErr := p.repo.FindLinkedUserIDBySlackUser(ctx, workspace.ID, event.TeamID, event.UserID)
	if linkErr != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(linkErr)); failErr != nil {
			return errors.Join(linkErr, failErr)
		}
		return linkErr
	}
	if currentLinkedUserID == nil || *currentLinkedUserID != *linkedUserID {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack assistant actor is no longer linked or active"); cancelErr != nil {
			return cancelErr
		}
		status = "ignored"
		return nil
	}
	if err := p.requireCurrentInstallation(ctx, workspace.ID, event.TeamID, installation.InstallGeneration); err != nil {
		if errors.Is(err, errSlackInstallationChanged) {
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack installation changed before assistant delivery"); cancelErr != nil {
				return errors.Join(err, cancelErr)
			}
			status = "ignored"
			return nil
		}
		return err
	}

	externalMessageID, err := p.sender.Send(ctx, botToken, SlackOutboundMessage{
		ChannelID:       deliveryChannelID,
		UserID:          event.UserID,
		ThreadTS:        deliveryThreadTS,
		Text:            reply,
		ClientMessageID: deterministicSlackMessageID(deliveryKey),
		Ephemeral:       false,
	})
	if err != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
			return errors.Join(err, failErr)
		}
		return err
	}
	if err := p.store.CompleteOutboundDelivery(ctx, delivery.ID, externalMessageID); err != nil {
		return err
	}
	if conversationID != uuid.Nil {
		if err := p.store.AppendMessage(ctx, conversationID, externalMessageID, "assistant", reply); err != nil {
			return err
		}
	}
	status = "completed"
	return nil
}

func inboundReceiptMatchesInstallation(receipt messagingrepository.InboundEventRecord, installation slackrepository.SlackWorkspaceRecord) bool {
	return receipt.InstallGeneration != nil &&
		*receipt.InstallGeneration != uuid.Nil &&
		*receipt.InstallGeneration == installation.InstallGeneration
}

func isSlackLifecycleEvent(kind slackEventKind) bool {
	return kind == slackEventKindUninstalled || kind == slackEventKindRevoked
}

func slackLifecycleEventIsCurrent(eventTimeUnix int64, authorizedAt time.Time) bool {
	if eventTimeUnix <= 0 || authorizedAt.IsZero() {
		return false
	}
	// Slack event_time has second precision. Events from the authorization
	// second are current; older lifecycle events belong to a prior install.
	return eventTimeUnix >= authorizedAt.UTC().Unix()
}

func installationBotTokenRevoked(installation slackrepository.SlackWorkspaceRecord, revokedBotUserIDs []string) bool {
	if len(revokedBotUserIDs) == 0 {
		return false
	}
	if installation.BotUserID == nil || strings.TrimSpace(*installation.BotUserID) == "" {
		return true
	}
	botUserID := strings.TrimSpace(*installation.BotUserID)
	for _, revokedUserID := range revokedBotUserIDs {
		if strings.TrimSpace(revokedUserID) == botUserID {
			return true
		}
	}
	return false
}

// BackfillLegacyCredentials encrypts bounded batches of pre-migration Slack
// tokens. It is safe to run concurrently with normal lazy credential upgrades.
func (p *EventProcessor) BackfillLegacyCredentials(ctx context.Context) (int, error) {
	repo, ok := p.repo.(legacyCredentialRepository)
	if !ok {
		return 0, errors.New("Slack repository does not support credential backfill")
	}

	updated := 0
	for {
		records, err := repo.ListLegacySlackCredentials(ctx, legacyCredentialBatchSize)
		if err != nil {
			return updated, fmt.Errorf("list legacy Slack credentials: %w", err)
		}
		for _, record := range records {
			credential, version, err := p.codec.open(record.Credential)
			if err != nil {
				return updated, fmt.Errorf("open legacy Slack credential %s: %w", record.SlackWorkspaceID, err)
			}
			encrypted := record.Credential
			currentVersion := version
			if version == 0 {
				encrypted, currentVersion, err = p.codec.seal(credential)
				if err != nil {
					return updated, fmt.Errorf("seal legacy Slack credential %s: %w", record.SlackWorkspaceID, err)
				}
			}
			if err := p.repo.UpgradeSlackCredential(ctx, record.SlackWorkspaceID, encrypted, currentVersion); err != nil {
				if slackrepository.IsNotFound(err) {
					continue
				}
				return updated, fmt.Errorf("upgrade legacy Slack credential %s: %w", record.SlackWorkspaceID, err)
			}
			updated++
		}
		if len(records) < legacyCredentialBatchSize {
			break
		}
	}
	for {
		scrubbed, err := repo.ScrubVersionedLegacySlackCredentials(ctx, legacyCredentialBatchSize)
		if err != nil {
			return updated, fmt.Errorf("scrub versioned legacy Slack credentials: %w", err)
		}
		updated += scrubbed
		if scrubbed < legacyCredentialBatchSize {
			break
		}
	}
	return updated, nil
}

// RecoverPendingEvents republishes encrypted inbox payloads when the original
// database-to-queue handoff failed or a worker task disappeared.
func (p *EventProcessor) RecoverPendingEvents(ctx context.Context) (int, error) {
	store, ok := p.store.(recoverableSlackEventStore)
	if !ok {
		return 0, errors.New("Slack event store does not support inbox recovery")
	}
	if p.eventQueue == nil {
		return 0, errors.New("Slack event recovery queue is not configured")
	}
	records, err := store.ClaimRecoverableInboundEvents(ctx, "slack", 500)
	if err != nil {
		return 0, err
	}
	recovered := 0
	var recoveryErrors []error
	for _, record := range records {
		if err := p.eventQueue.EnqueueSlackEvent(ctx, tasks.SlackEventPayload{
			ExternalWorkspaceID: record.ExternalWorkspaceID,
			EventID:             record.ExternalEventID,
			RecoveryAttempt:     record.RecoveryGeneration,
		}); err != nil {
			releaseErr := store.ReleaseInboundEventRecovery(ctx, record.ID, record.RecoveryGeneration)
			recoveryErrors = append(recoveryErrors, fmt.Errorf("re-enqueue Slack event %s: %w", record.ExternalEventID, err))
			if releaseErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack event %s recovery claim: %w", record.ExternalEventID, releaseErr))
			}
			continue
		}
		recovered++
	}
	outboundRecovered, outboundErr := p.recoverPendingOutboundDeliveries(ctx, store)
	if outboundErr != nil {
		recoveryErrors = append(recoveryErrors, outboundErr)
	}
	uninstallRecovered, uninstallErr := p.recoverSlackUninstalls(ctx)
	if uninstallErr != nil {
		recoveryErrors = append(recoveryErrors, uninstallErr)
	}
	return recovered + outboundRecovered + uninstallRecovered, errors.Join(recoveryErrors...)
}

func (p *EventProcessor) recoverSlackUninstalls(ctx context.Context) (int, error) {
	records, err := p.repo.ClaimRecoverableSlackUninstalls(ctx, 100)
	if err != nil {
		return 0, err
	}
	recovered := 0
	var recoveryErrors []error
	for _, record := range records {
		attemptCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		completed, uninstallErr := executeSlackUninstall(
			attemptCtx,
			p.repo,
			p.repo,
			p.webClient,
			p.codec,
			p.clientID,
			p.clientSecret,
			p.clock.Now(),
			record,
		)
		cancel()
		if uninstallErr != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recover Slack uninstall %s: %w", record.ID, uninstallErr))
			continue
		}
		if completed {
			recovered++
		}
	}
	return recovered, errors.Join(recoveryErrors...)
}

func (p *EventProcessor) recoverPendingOutboundDeliveries(ctx context.Context, store recoverableSlackEventStore) (int, error) {
	records, err := store.ListRecoverableOutboundDeliveries(ctx, "slack", 500)
	if err != nil {
		return 0, err
	}
	recovered := 0
	var recoveryErrors []error
	for _, record := range records {
		externalWorkspaceID := strings.TrimSpace(record.ExternalWorkspaceID)
		if externalWorkspaceID == "" {
			err := errors.New("Slack outbound delivery has no external workspace binding")
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, record.ID, err.Error()); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel malformed Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recover Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		threadID := ""
		if record.ExternalThreadID != nil {
			threadID = strings.TrimSpace(*record.ExternalThreadID)
		}
		delivery, claimed, err := p.store.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
			Provider:                "slack",
			WorkspaceID:             record.WorkspaceID,
			UserID:                  record.UserID,
			InstallGeneration:       record.InstallGeneration,
			ExternalWorkspaceID:     externalWorkspaceID,
			ExternalRecipientUserID: valueOrEmpty(record.ExternalRecipientUserID),
			InboundEventID:          record.InboundEventID,
			IdempotencyKey:          record.IdempotencyKey,
			ExternalChannelID:       record.ExternalChannelID,
			ExternalThreadID:        threadID,
			Content:                 valueOrEmpty(record.Content),
			Purpose:                 record.Purpose,
			ExpiresAt:               record.ExpiresAt,
		})
		if err != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf("claim Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if !claimed {
			continue
		}
		if delivery.ExpiresAt != nil && !p.clock.Now().UTC().Before(delivery.ExpiresAt.UTC()) {
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack delivery expired before recovery"); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel expired Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
			}
			continue
		}
		content := ""
		if delivery.Content != nil {
			content = strings.TrimSpace(*delivery.Content)
		}
		if content == "" && record.Content != nil {
			content = strings.TrimSpace(*record.Content)
		}
		if content == "" {
			err := errors.New("Slack outbound delivery has no content")
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, err.Error()); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel empty Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recover Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		installation, err := p.repo.GetSlackWorkspaceByTeamID(ctx, externalWorkspaceID)
		if err != nil {
			if slackrepository.IsNotFound(err) {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack installation no longer exists"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel disconnected Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after installation lookup: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("load Slack installation for delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if err := validateOutboundInstallation(record, installation); err != nil {
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack installation binding changed before recovery"); cancelErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel Slack delivery %s after installation change: %w", record.IdempotencyKey, cancelErr))
			}
			continue
		}
		if delivery.Purpose == "assistant" {
			if delivery.UserID == nil || delivery.ExternalRecipientUserID == nil || strings.TrimSpace(*delivery.ExternalRecipientUserID) == "" {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack assistant delivery is missing its actor binding"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel unbound Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
			linkedUserID, linkErr := p.repo.FindLinkedUserIDBySlackUser(
				ctx,
				installation.WorkspaceID,
				externalWorkspaceID,
				strings.TrimSpace(*delivery.ExternalRecipientUserID),
			)
			if linkErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("revalidate Slack delivery %s actor: %w", record.IdempotencyKey, linkErr))
				if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(linkErr)); failErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after actor lookup: %w", record.IdempotencyKey, failErr))
				}
				continue
			}
			if linkedUserID == nil || *linkedUserID != *delivery.UserID {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack assistant actor is no longer linked or active"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel unauthorized Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
		}
		botToken, err := p.botToken(ctx, installation)
		if err != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after credential lookup: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("load Slack credential for delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if err := p.requireCurrentInstallation(ctx, record.WorkspaceID, externalWorkspaceID, *record.InstallGeneration); err != nil {
			if errors.Is(err, errSlackInstallationChanged) {
				if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack installation changed before recovered delivery"); cancelErr != nil {
					recoveryErrors = append(recoveryErrors, fmt.Errorf("cancel stale Slack delivery %s: %w", record.IdempotencyKey, cancelErr))
				}
				continue
			}
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after installation recheck: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("recheck Slack installation for delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		externalMessageID, err := p.sender.Send(ctx, botToken, SlackOutboundMessage{
			ChannelID:       record.ExternalChannelID,
			ThreadTS:        threadID,
			Text:            content,
			ClientMessageID: deterministicSlackMessageID(record.IdempotencyKey),
		})
		if err != nil {
			if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
				recoveryErrors = append(recoveryErrors, fmt.Errorf("release Slack delivery %s after send failure: %w", record.IdempotencyKey, failErr))
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf("send Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		if err := p.store.CompleteOutboundDelivery(ctx, delivery.ID, externalMessageID); err != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf("complete Slack delivery %s: %w", record.IdempotencyKey, err))
			continue
		}
		recovered++
	}
	return recovered, errors.Join(recoveryErrors...)
}

func validateOutboundInstallation(record messagingrepository.OutboundDeliveryRecord, installation slackrepository.SlackWorkspaceRecord) error {
	externalWorkspaceID := strings.TrimSpace(record.ExternalWorkspaceID)
	if externalWorkspaceID == "" {
		return errors.New("Slack outbound delivery has no external workspace binding")
	}
	if installation.WorkspaceID != record.WorkspaceID {
		return fmt.Errorf(
			"Slack installation workspace mismatch for external workspace %q: delivery workspace %s, installation workspace %s",
			externalWorkspaceID,
			record.WorkspaceID,
			installation.WorkspaceID,
		)
	}
	if strings.TrimSpace(installation.SlackTeamID) != externalWorkspaceID {
		return fmt.Errorf(
			"Slack installation team mismatch: delivery team %q, installation team %q",
			externalWorkspaceID,
			installation.SlackTeamID,
		)
	}
	if record.InstallGeneration == nil || *record.InstallGeneration == uuid.Nil || *record.InstallGeneration != installation.InstallGeneration {
		return fmt.Errorf("Slack installation generation mismatch for external workspace %q", externalWorkspaceID)
	}
	if !installation.IsActive {
		return fmt.Errorf("Slack installation for external workspace %q is inactive", externalWorkspaceID)
	}
	return nil
}

func (p *EventProcessor) requireCurrentInstallation(ctx context.Context, workspaceID uuid.UUID, slackTeamID string, generation uuid.UUID) error {
	current, err := p.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return fmt.Errorf("%w: Slack team is no longer connected", errSlackInstallationChanged)
		}
		return err
	}
	if !current.IsActive || current.WorkspaceID != workspaceID || current.InstallGeneration != generation {
		return fmt.Errorf("%w: active generation no longer matches", errSlackInstallationChanged)
	}
	return nil
}

func (p *EventProcessor) botToken(ctx context.Context, installation slackrepository.SlackWorkspaceRecord) (string, error) {
	credential, version, err := p.codec.open(installation.BotAccessToken)
	if err != nil {
		return "", err
	}
	if version == 0 || installation.CredentialVersion == 0 {
		encrypted, currentVersion, sealErr := p.codec.seal(credential)
		if sealErr != nil {
			return "", sealErr
		}
		if upgradeErr := p.repo.UpgradeSlackCredential(ctx, installation.ID, encrypted, currentVersion); upgradeErr != nil && !slackrepository.IsNotFound(upgradeErr) {
			return "", upgradeErr
		}
	}
	return credential.AccessToken, nil
}

func (p *EventProcessor) accountLinkMessage(ctx context.Context, workspace slackrepository.WorkspaceRecord, event normalizedSlackEvent) (string, error) {
	nonce := make([]byte, 32)
	if _, err := io.ReadFull(p.random, nonce); err != nil {
		return "", fmt.Errorf("generate Slack account-link nonce: %w", err)
	}
	digest := sha256.Sum256(nonce)
	if err := p.store.CreateNonce(ctx, messagingrepository.NonceInput{
		Provider:            "slack",
		Purpose:             "account_link",
		NonceHash:           digest[:],
		WorkspaceID:         workspace.ID,
		ExternalWorkspaceID: event.TeamID,
		ExternalUserID:      event.UserID,
		ExpiresAt:           time.Now().UTC().Add(15 * time.Minute),
	}); err != nil {
		return "", err
	}
	link := p.website + "/" + url.PathEscape(workspace.Slug) + "/settings/integrations/slack"
	link += "?slack_link_token=" + url.QueryEscape(base64.RawURLEncoding.EncodeToString(nonce))
	return "Connect your FortyOne account before using Maya in Slack: " + link, nil
}

func (p *EventProcessor) deliver(ctx context.Context, inboundEventID uuid.UUID, workspaceID, installGeneration uuid.UUID, userID *uuid.UUID, event normalizedSlackEvent, botToken, suffix, text string) error {
	channelID := event.ChannelID
	threadTS := event.ReplyTS
	if event.Kind != slackEventKindDirect {
		channelID = event.UserID
		threadTS = ""
	}
	purpose := "assistant"
	expiresAt := p.clock.Now().UTC().Add(time.Hour)
	if suffix == "link" {
		purpose = "account_link"
		expiresAt = p.clock.Now().UTC().Add(15 * time.Minute)
	} else if suffix == "access" {
		purpose = "access"
	}
	delivery, send, err := p.store.StartOutboundDelivery(ctx, messagingrepository.OutboundDeliveryInput{
		Provider:                "slack",
		WorkspaceID:             workspaceID,
		UserID:                  userID,
		InstallGeneration:       &installGeneration,
		ExternalWorkspaceID:     event.TeamID,
		ExternalRecipientUserID: event.UserID,
		InboundEventID:          &inboundEventID,
		IdempotencyKey:          event.EventID + ":" + suffix,
		ExternalChannelID:       channelID,
		ExternalThreadID:        threadTS,
		Content:                 text,
		Purpose:                 purpose,
		ExpiresAt:               &expiresAt,
	})
	if err != nil || !send {
		return err
	}
	persistedExpiresAt := expiresAt
	if delivery.ExpiresAt != nil {
		persistedExpiresAt = delivery.ExpiresAt.UTC()
	}
	if !p.clock.Now().UTC().Before(persistedExpiresAt) {
		return cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack delivery expired before send")
	}
	if err := p.store.SetOutboundDeliveryContent(ctx, delivery.ID, text); err != nil {
		return err
	}
	if err := p.requireCurrentInstallation(ctx, workspaceID, event.TeamID, installGeneration); err != nil {
		if errors.Is(err, errSlackInstallationChanged) {
			return cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack installation changed before delivery")
		}
		return err
	}
	externalMessageID, err := p.sender.Send(ctx, botToken, SlackOutboundMessage{
		ChannelID:       channelID,
		UserID:          event.UserID,
		ThreadTS:        threadTS,
		Text:            truncateSlackText(text),
		ClientMessageID: deterministicSlackMessageID(event.EventID + ":" + suffix),
		Ephemeral:       false,
	})
	if err != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, delivery.ID, truncateError(err)); failErr != nil {
			return errors.Join(err, failErr)
		}
		return err
	}
	return p.store.CompleteOutboundDelivery(ctx, delivery.ID, externalMessageID)
}

func (p *EventProcessor) recordAssistantUsage(parent context.Context, input messagingrepository.DailyUsageRecordInput) (messagingrepository.DailyUsageSnapshot, error) {
	usageCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), slackStateWriteTimeout)
	defer cancel()
	return p.usageBudget.Record(usageCtx, input, p.dailyWorkspaceTokenLimit)
}

func (p *EventProcessor) persistAssistantPrompt(ctx context.Context, workspaceID, userID uuid.UUID, event normalizedSlackEvent, prompt string) (uuid.UUID, error) {
	conversationID, err := p.store.UpsertConversation(ctx, assistantConversationInput(workspaceID, userID, event))
	if err != nil {
		return uuid.Nil, err
	}
	if err := p.store.AppendMessage(ctx, conversationID, event.MessageTS, "user", prompt); err != nil {
		return uuid.Nil, err
	}
	return conversationID, nil
}

func (p *EventProcessor) channelThreadIsSubscribed(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	installation slackrepository.SlackWorkspaceRecord,
	event normalizedSlackEvent,
) (bool, error) {
	record, err := p.store.FindConversation(ctx, assistantConversationInput(workspaceID, userID, event))
	if err != nil {
		if errors.Is(err, messagingrepository.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return slackThreadSubscriptionIsCurrent(record, installation, p.clock.Now()), nil
}

func slackThreadSubscriptionIsCurrent(
	record messagingrepository.ConversationRecord,
	installation slackrepository.SlackWorkspaceRecord,
	now time.Time,
) bool {
	if record.ID == uuid.Nil || record.UpdatedAt.IsZero() {
		return false
	}
	updatedAt := record.UpdatedAt.UTC()
	if installation.AuthorizedAt.IsZero() || updatedAt.Before(installation.AuthorizedAt.UTC()) {
		return false
	}
	return updatedAt.After(now.UTC().Add(-slackThreadSubscriptionTTL))
}

func assistantConversationInput(workspaceID, userID uuid.UUID, event normalizedSlackEvent) messagingrepository.ConversationInput {
	return messagingrepository.ConversationInput{
		Provider:            "slack",
		WorkspaceID:         workspaceID,
		ExternalWorkspaceID: event.TeamID,
		ExternalChannelID:   event.ChannelID,
		ExternalThreadID:    conversationThreadID(event),
		UserID:              userID,
	}
}

func isAssistantBudgetNotice(content string) bool {
	switch strings.TrimSpace(content) {
	case assistantUserRateLimitReply, assistantWorkspaceRateReply, assistantDailyLimitReply:
		return true
	default:
		return false
	}
}

func failOutboundDeliveryDetached(parent context.Context, store outboundDeliveryStateStore, id uuid.UUID, message string) error {
	stateCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), slackStateWriteTimeout)
	defer cancel()
	return store.FailOutboundDelivery(stateCtx, id, message)
}

func cancelOutboundDeliveryDetached(parent context.Context, store outboundDeliveryStateStore, id uuid.UUID, message string) error {
	stateCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), slackStateWriteTimeout)
	defer cancel()
	return store.CancelOutboundDelivery(stateCtx, id, message)
}

func conversationThreadID(event normalizedSlackEvent) string {
	if event.Kind == slackEventKindDirect && event.ReplyTS == "" {
		return "dm:" + event.ChannelID
	}
	return event.ThreadTS
}

func deterministicSlackMessageID(value string) string {
	digest := sha256.Sum256([]byte(value))
	bytes := digest[:16]
	bytes[6] = (bytes[6] & 0x0f) | 0x50
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	id, err := uuid.FromBytes(bytes)
	if err != nil {
		return hex.EncodeToString(digest[:16])
	}
	return id.String()
}

func truncateSlackText(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= slackMessageTextLimit {
		return value
	}
	return strings.TrimSpace(string(runes[:slackMessageTextLimit-1])) + "…"
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	if len(message) <= 1000 {
		return message
	}
	return message[:1000]
}

type slackAPISender struct {
	client *slackWebClient
}

func (s *slackAPISender) Send(ctx context.Context, botToken string, message SlackOutboundMessage) (string, error) {
	method := "chat.postMessage"
	payload := map[string]any{
		"channel": message.ChannelID,
		"text":    message.Text,
	}
	if strings.TrimSpace(message.ThreadTS) != "" {
		payload["thread_ts"] = message.ThreadTS
	}
	if message.Ephemeral {
		method = "chat.postEphemeral"
		payload["user"] = message.UserID
	} else if strings.TrimSpace(message.ClientMessageID) != "" {
		payload["client_msg_id"] = message.ClientMessageID
	}
	var response struct {
		TS        string `json:"ts"`
		MessageTS string `json:"message_ts"`
	}
	if err := s.client.callJSON(ctx, botToken, method, payload, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.TS) != "" {
		return strings.TrimSpace(response.TS), nil
	}
	return strings.TrimSpace(response.MessageTS), nil
}

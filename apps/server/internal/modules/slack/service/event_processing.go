package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

const slackWebhookProcessingLease = 2 * time.Minute

// ProcessEvent preserves compatibility with Slack tasks enqueued before the
// inbox-ID rollout. It resolves the legacy provider identity once, then uses
// the same ID-based processing path as every new task.
func (p *EventProcessor) ProcessEvent(ctx context.Context, externalWorkspaceID, eventID string) error {
	externalWorkspaceID = strings.TrimSpace(externalWorkspaceID)
	eventID = strings.TrimSpace(eventID)
	if externalWorkspaceID == "" || eventID == "" {
		return errors.New("slack external workspace id and event id are required")
	}
	record, err := p.webhookInbox.GetByExternalKey(ctx, slackWebhookProvider, externalWorkspaceID, eventID)
	if err != nil {
		return err
	}
	return p.ProcessWebhook(ctx, slackWebhookProvider, record.ID)
}

// Process resolves raw payloads through their durable identity. Production
// workers call ProcessWebhook directly; this method remains for package API
// compatibility and never processes an unregistered payload.
func (p *EventProcessor) Process(ctx context.Context, rawBody []byte) error {
	envelope, err := decodeSlackEvent(rawBody)
	if err != nil {
		return err
	}
	return p.ProcessEvent(ctx, envelope.TeamID, envelope.EventID)
}

// ProcessWebhook claims and processes one encrypted canonical payload by its
// provider-neutral inbox identity. Queue backends never receive message text.
func (p *EventProcessor) ProcessWebhook(
	ctx context.Context,
	provider integrations.ProviderKey,
	inboxID uuid.UUID,
) (err error) {
	if provider != slackWebhookProvider || inboxID == uuid.Nil {
		return webhooks.ErrInvalidDelivery
	}
	registered, err := p.webhookInbox.GetByID(ctx, inboxID)
	if err != nil {
		return err
	}
	if registered.Provider != provider || registered.ID != inboxID {
		return webhooks.ErrInvalidDelivery
	}
	if registered.Status.Terminal() {
		return nil
	}
	record, process, err := p.webhookInbox.Start(ctx, inboxID, p.clock.Now().UTC(), slackWebhookProcessingLease)
	if err != nil {
		return err
	}
	if !process {
		return nil
	}

	status := webhooks.StatusFailed
	outcomeCode := "slack.processing_failed"
	defer func() {
		if completeErr := p.webhookInbox.Complete(
			context.WithoutCancel(ctx),
			record.ID,
			status,
			outcomeCode,
			p.clock.Now().UTC(),
		); completeErr != nil {
			if p.log != nil {
				p.log.Error(context.WithoutCancel(ctx), "failed updating Slack event receipt", "error", completeErr, "inbox_id", record.ID)
			}
			if err == nil {
				err = completeErr
			}
		}
	}()

	if record.EncryptedPayload == nil || strings.TrimSpace(*record.EncryptedPayload) == "" {
		return errors.New("slack inbox event has no encrypted payload")
	}
	body, err := p.webhookPayloads.Open(record, *record.EncryptedPayload)
	if err != nil {
		return err
	}
	defer clear(body)
	envelope, err := decodeSlackEvent(body)
	if err != nil {
		return err
	}
	if envelope.TeamID != record.ExternalAccountID || envelope.EventID != record.DeliveryID {
		return errors.New("slack inbox payload does not match its durable identity")
	}
	result, processErr := p.processClaimedEvent(ctx, inboundReceipt(record), envelope)
	status, outcomeCode, err = webhookCompletion(result, processErr)
	return err
}

func webhookCompletion(status string, err error) (webhooks.Status, string, error) {
	if err != nil {
		return webhooks.StatusFailed, "slack.processing_failed", err
	}
	switch status {
	case string(webhooks.StatusCompleted):
		return webhooks.StatusCompleted, "slack.completed", nil
	case string(webhooks.StatusIgnored):
		return webhooks.StatusIgnored, "slack.ignored", nil
	default:
		return webhooks.StatusFailed, "slack.invalid_completion", fmt.Errorf("invalid Slack event completion status %q", status)
	}
}

func (p *EventProcessor) processClaimedEvent(
	ctx context.Context,
	receipt inboundEventRecord,
	envelope slackEventEnvelope,
) (string, error) {
	event, supported := normalizeSlackEvent(envelope)
	if !supported {
		return "ignored", nil
	}
	if event.Kind == slackEventKindEntityDetails {
		// Entity detail triggers are single-use and expire before a durable worker
		// can safely consume them. The API handles new requests synchronously;
		// ignore legacy inbox rows so recovery never spends a stale trigger.
		return "ignored", nil
	}

	installation, err := p.repo.GetSlackWorkspaceByTeamID(ctx, event.TeamID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return "ignored", nil
		}
		return "failed", err
	}
	if !inboundReceiptMatchesInstallation(receipt, installation) {
		return "ignored", nil
	}
	if isSlackLifecycleEvent(event.Kind) && !slackLifecycleEventIsCurrent(envelope.EventTime, installation.AuthorizedAt) {
		return "ignored", nil
	}
	if event.Kind == slackEventKindUninstalled {
		if deactivateErr := p.repo.DeactivateSlackWorkspaceByTeamID(ctx, event.TeamID, installation.InstallGeneration); deactivateErr != nil && !isSlackRepositoryNotFound(deactivateErr) {
			return "failed", deactivateErr
		}
		return "completed", nil
	}
	if event.Kind == slackEventKindRevoked {
		if installationBotTokenRevoked(installation, event.RevokedBotUserIDs) {
			if deactivateErr := p.repo.DeactivateSlackWorkspaceByTeamID(ctx, event.TeamID, installation.InstallGeneration); deactivateErr != nil && !isSlackRepositoryNotFound(deactivateErr) {
				return "failed", deactivateErr
			}
		}
		return "completed", nil
	}

	workspace, err := p.repo.FindWorkspaceByID(ctx, installation.WorkspaceID)
	if err != nil {
		return "failed", err
	}
	botToken, err := p.botToken(ctx, installation)
	if err != nil {
		return "failed", err
	}
	linkedUserID, err := p.repo.FindLinkedUserIDBySlackUser(ctx, workspace.ID, event.TeamID, event.UserID)
	if err != nil {
		return "failed", err
	}
	if isSlackWorkObjectEvent(event.Kind) {
		if err := p.processSlackWorkObjectEvent(ctx, workspace, installation, linkedUserID, event, botToken); err != nil {
			return "failed", err
		}
		return "completed", nil
	}
	if status, done, err := p.routeAssistantThreadEvent(ctx, workspace.ID, installation, linkedUserID, event); err != nil || done {
		return status, err
	}
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		text, linkErr := p.accountLinkMessage(ctx, workspace, event)
		if linkErr != nil {
			return "failed", linkErr
		}
		if err := p.deliver(ctx, receipt.ID, workspace.ID, installation.InstallGeneration, nil, event, botToken, "link", text); err != nil {
			return "failed", err
		}
		return "completed", nil
	}

	agentSettings, err := p.agentSettings(ctx, workspace.ID)
	if err != nil {
		return "failed", err
	}
	teamScope, err := p.authorizedAssistantTeamScope(ctx, workspace.ID, installation, *linkedUserID, event)
	if err != nil {
		return "failed", err
	}
	allowedTeamIDs := teamScope.AllowedTeamIDs
	sharedTeamIDs := teamScope.SharedTeamIDs
	if event.Kind != slackEventKindDirect && len(allowedTeamIDs) == 0 {
		if err := p.deliver(ctx, receipt.ID, workspace.ID, installation.InstallGeneration, linkedUserID, event, botToken, "channel-access", "I can't access any FortyOne teams for you from this Slack channel. Ask a workspace administrator to configure the channel audience."); err != nil {
			return "failed", err
		}
		return "completed", nil
	}

	allowed, err := p.access.CanUseAssistant(ctx, workspace.ID)
	if err != nil {
		return "failed", err
	}
	if !allowed {
		if err := p.deliver(ctx, receipt.ID, workspace.ID, installation.InstallGeneration, linkedUserID, event, botToken, "access", "Maya is available on FortyOne paid plans and active trials."); err != nil {
			return "failed", err
		}
		return "completed", nil
	}

	prompt := event.Text
	if installation.BotUserID != nil {
		prompt = removeBotMention(prompt, *installation.BotUserID)
	}
	if prompt == "" {
		prompt = "How can you help me with my FortyOne workspace?"
	}
	if validateErr := validateAssistantPrompt(prompt); validateErr != nil {
		if !errors.Is(validateErr, errAssistantPromptTooLarge) {
			return "failed", validateErr
		}
		if err := p.deliver(ctx, receipt.ID, workspace.ID, installation.InstallGeneration, linkedUserID, event, botToken, "assistant-input-too-large", assistantMessageTooLargeReply); err != nil {
			return "failed", err
		}
		return "completed", nil
	}

	return p.processAssistantDelivery(ctx, assistantEventInput{
		receipt:        receipt,
		workspace:      workspace,
		installation:   installation,
		linkedUserID:   *linkedUserID,
		event:          event,
		botToken:       botToken,
		agentSettings:  agentSettings,
		allowedTeamIDs: allowedTeamIDs,
		sharedTeamIDs:  sharedTeamIDs,
		prompt:         prompt,
	})
}

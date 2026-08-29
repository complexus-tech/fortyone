package slack

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (p *EventProcessor) authorizeAndSendAssistantDelivery(
	ctx context.Context,
	input assistantEventInput,
	state *assistantDeliveryState,
) (string, error) {
	if !p.clock.Now().UTC().Before(state.expiresAt) {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, state.record.ID, "Slack assistant delivery expired before send"); cancelErr != nil {
			return "failed", cancelErr
		}
		return "ignored", nil
	}
	currentLinkedUserID, linkErr := p.repo.FindLinkedUserIDBySlackUser(ctx, input.workspace.ID, input.event.TeamID, input.event.UserID)
	if linkErr != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(linkErr)); failErr != nil {
			return "failed", errors.Join(linkErr, failErr)
		}
		return "failed", linkErr
	}
	if currentLinkedUserID == nil || *currentLinkedUserID != input.linkedUserID {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, state.record.ID, "Slack assistant actor is no longer linked or active"); cancelErr != nil {
			return "failed", cancelErr
		}
		return "ignored", nil
	}
	if err := p.requireCurrentInstallation(ctx, input.workspace.ID, input.event.TeamID, input.installation.InstallGeneration); err != nil {
		if errors.Is(err, errSlackInstallationChanged) {
			if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, state.record.ID, "Slack installation changed before assistant delivery"); cancelErr != nil {
				return "failed", errors.Join(err, cancelErr)
			}
			return "ignored", nil
		}
		return "failed", err
	}
	if current, audienceErr := p.slackChannelDeliveryAuthorizationCurrent(
		ctx,
		input.workspace.ID,
		input.installation,
		input.linkedUserID,
		state.channelID,
		input.event.UserID,
		state.providerPayload,
	); audienceErr != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(audienceErr)); failErr != nil {
			return "failed", errors.Join(audienceErr, failErr)
		}
		return "failed", audienceErr
	} else if !current {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, state.record.ID, "Slack channel audience narrowed before assistant delivery"); cancelErr != nil {
			return "failed", cancelErr
		}
		return "ignored", nil
	}
	currentSettings, settingsErr := p.agentSettings(ctx, input.workspace.ID)
	if settingsErr != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(settingsErr)); failErr != nil {
			return "failed", errors.Join(settingsErr, failErr)
		}
		return "failed", settingsErr
	}
	if !assistantSettingsAllowDelivery(currentSettings, state.providerPayload) {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, state.record.ID, "Slack agent settings changed before assistant delivery"); cancelErr != nil {
			return "failed", cancelErr
		}
		return "ignored", nil
	}

	externalMessageID, err := p.sender.Send(ctx, input.botToken, SlackOutboundMessage{
		ChannelID:        state.channelID,
		UserID:           input.event.UserID,
		ThreadTS:         state.threadTS,
		Text:             state.reply,
		ClientMessageID:  deterministicSlackMessageID(state.key),
		Ephemeral:        false,
		StandardMarkdown: len(state.providerPayload.Blocks) == 0,
		ProviderPayload:  state.providerPayload,
	})
	if err != nil {
		if failErr := failOutboundDeliveryDetached(ctx, p.store, state.record.ID, truncateError(err)); failErr != nil {
			return "failed", errors.Join(err, failErr)
		}
		return "failed", err
	}
	// Slack only auto-clears status when the response is posted into the exact
	// status thread. Direct-message replies may be posted as new messages, so
	// explicitly clear after every successful response. If this call fails, the
	// deferred cleanup makes one more bounded attempt.
	if state.thinkingStatusActive && p.clearAssistantThinkingStatus(ctx, input.event, input.botToken) {
		state.thinkingStatusActive = false
	}
	if err := p.store.CompleteOutboundDelivery(ctx, state.record.ID, externalMessageID); err != nil {
		return "failed", err
	}
	if state.conversationID != uuid.Nil {
		if err := p.store.AppendMessage(ctx, state.conversationID, externalMessageID, "assistant", state.reply); err != nil {
			return "failed", err
		}
	}
	return "completed", nil
}

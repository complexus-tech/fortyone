package slack

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type assistantEventInput struct {
	receipt        inboundEventRecord
	workspace      workspaceRecord
	installation   slackWorkspaceRecord
	linkedUserID   uuid.UUID
	event          normalizedSlackEvent
	botToken       string
	agentSettings  CoreSlackAgentSettings
	allowedTeamIDs []uuid.UUID
	sharedTeamIDs  []uuid.UUID
	prompt         string
}

type assistantDeliveryState struct {
	record               outboundDeliveryRecord
	key                  string
	channelID            string
	threadTS             string
	expiresAt            time.Time
	providerPayload      SlackProviderPayload
	reply                string
	contentPersisted     bool
	persistConversation  bool
	conversationID       uuid.UUID
	thinkingStatusActive bool
}

func (p *EventProcessor) processAssistantDelivery(ctx context.Context, input assistantEventInput) (string, error) {
	state, status, done, err := p.startAssistantDelivery(ctx, input)
	if err != nil || done {
		return status, err
	}
	defer func() {
		if state.thinkingStatusActive {
			p.clearAssistantThinkingStatus(ctx, input.event, input.botToken)
		}
	}()

	if err := p.prepareAssistantReply(ctx, input, state); err != nil {
		return "failed", err
	}
	if !state.contentPersisted {
		if err := persistSlackOutboundContent(ctx, p.store, state.record.ID, state.reply, state.providerPayload); err != nil {
			return "failed", err
		}
	}
	return p.authorizeAndSendAssistantDelivery(ctx, input, state)
}

func (p *EventProcessor) startAssistantDelivery(
	ctx context.Context,
	input assistantEventInput,
) (*assistantDeliveryState, string, bool, error) {
	deliveryKey := input.event.EventID + ":assistant"
	deliveryChannelID := input.event.ChannelID
	deliveryThreadTS := input.event.ReplyTS
	deliveryExpiresAt := p.clock.Now().UTC().Add(time.Hour)
	providerPayload := SlackProviderPayload{}
	if len(input.allowedTeamIDs) > 0 {
		actorID := input.linkedUserID
		providerPayload.Authorization = &SlackDeliveryAuthorization{
			AllowedTeamIDs: append([]uuid.UUID(nil), input.allowedTeamIDs...),
			SharedTeamIDs:  append([]uuid.UUID(nil), input.sharedTeamIDs...),
			ActorUserID:    &actorID,
		}
	}
	var encodedProviderPayload []byte
	if !slackProviderPayloadIsEmpty(providerPayload) {
		var err error
		encodedProviderPayload, err = EncodeSlackProviderPayload(providerPayload)
		if err != nil {
			return nil, "failed", true, err
		}
	}
	delivery, shouldDeliver, err := p.store.StartOutboundDelivery(ctx, outboundDeliveryInput{
		Provider:                "slack",
		WorkspaceID:             input.workspace.ID,
		UserID:                  &input.linkedUserID,
		InstallGeneration:       &input.installation.InstallGeneration,
		ExternalWorkspaceID:     input.event.TeamID,
		ExternalRecipientUserID: input.event.UserID,
		InboundEventID:          &input.receipt.ID,
		IdempotencyKey:          deliveryKey,
		ExternalChannelID:       deliveryChannelID,
		ExternalThreadID:        deliveryThreadTS,
		ProviderPayload:         encodedProviderPayload,
		Purpose:                 "assistant",
		ExpiresAt:               &deliveryExpiresAt,
	})
	if err != nil {
		return nil, "failed", true, err
	}
	if delivery.ExpiresAt != nil {
		deliveryExpiresAt = delivery.ExpiresAt.UTC()
	}
	if strings.TrimSpace(delivery.ExternalChannelID) != "" {
		deliveryChannelID = strings.TrimSpace(delivery.ExternalChannelID)
	}
	if delivery.ExternalThreadID != nil {
		deliveryThreadTS = strings.TrimSpace(*delivery.ExternalThreadID)
	} else if deliveryChannelID != input.event.ChannelID {
		deliveryThreadTS = ""
	}
	if !shouldDeliver {
		if delivery.Status == "delivered" && delivery.Content != nil && delivery.ExternalMessageID != nil {
			content := strings.TrimSpace(*delivery.Content)
			if !isAssistantBudgetNotice(content) {
				conversationID, conversationErr := p.persistAssistantPrompt(
					ctx,
					input.workspace.ID,
					input.linkedUserID,
					input.event,
					input.prompt,
					input.allowedTeamIDs,
					input.sharedTeamIDs,
				)
				if conversationErr != nil {
					return nil, "failed", true, conversationErr
				}
				if err := p.store.AppendMessage(ctx, conversationID, *delivery.ExternalMessageID, "assistant", content); err != nil {
					return nil, "failed", true, err
				}
			}
		}
		return nil, "completed", true, nil
	}

	persistedProviderPayload, payloadErr := DecodeSlackProviderPayload(delivery.ProviderPayload)
	if payloadErr != nil {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack assistant delivery has an invalid provider payload"); cancelErr != nil {
			return nil, "failed", true, errors.Join(payloadErr, cancelErr)
		}
		return nil, "failed", true, payloadErr
	}
	reply := ""
	if delivery.Content != nil {
		reply = strings.TrimSpace(*delivery.Content)
	}
	if !p.clock.Now().UTC().Before(deliveryExpiresAt) {
		if cancelErr := cancelOutboundDeliveryDetached(ctx, p.store, delivery.ID, "Slack assistant delivery expired before send"); cancelErr != nil {
			return nil, "failed", true, cancelErr
		}
		return nil, "ignored", true, nil
	}
	return &assistantDeliveryState{
		record:              delivery,
		key:                 deliveryKey,
		channelID:           deliveryChannelID,
		threadTS:            deliveryThreadTS,
		expiresAt:           deliveryExpiresAt,
		providerPayload:     persistedProviderPayload,
		reply:               reply,
		contentPersisted:    reply != "",
		persistConversation: !isAssistantBudgetNotice(reply),
	}, "", false, nil
}

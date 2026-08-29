package slack

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	delivery, send, err := p.store.StartOutboundDelivery(ctx, outboundDeliveryInput{
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

func persistSlackOutboundContent(
	ctx context.Context,
	store slackOutboundContentStore,
	deliveryID uuid.UUID,
	content string,
	providerPayload SlackProviderPayload,
) error {
	if slackProviderPayloadIsEmpty(providerPayload) {
		return store.SetOutboundDeliveryContent(ctx, deliveryID, content)
	}
	payloadStore, ok := store.(slackProviderPayloadStore)
	if !ok {
		return errors.New("slack provider payload store is not configured")
	}
	encoded, err := EncodeSlackProviderPayload(providerPayload)
	if err != nil {
		return err
	}
	return payloadStore.SetOutboundDeliveryContentAndProviderPayload(ctx, deliveryID, content, encoded)
}

func slackProviderPayloadIsEmpty(payload SlackProviderPayload) bool {
	return len(payload.Blocks) == 0 && payload.Metadata == nil && payload.UnfurlLinks == nil && payload.UnfurlMedia == nil && payload.Authorization == nil && payload.RequestThreadBinding == nil
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

type slackAPISender struct {
	client *slackWebClient
}

func (s *slackAPISender) Send(ctx context.Context, botToken string, message SlackOutboundMessage) (string, error) {
	method := "chat.postMessage"
	payload := map[string]any{
		"channel": message.ChannelID,
	}
	if message.StandardMarkdown {
		if len(message.ProviderPayload.Blocks) > 0 {
			return "", errors.New("slack standard Markdown cannot be combined with Block Kit blocks")
		}
		payload["markdown_text"] = message.Text
	} else {
		payload["text"] = message.Text
	}
	if !slackProviderPayloadIsEmpty(message.ProviderPayload) {
		if _, err := EncodeSlackProviderPayload(message.ProviderPayload); err != nil {
			return "", err
		}
		applySlackProviderPayload(payload, message.ProviderPayload)
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

package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

func (s *Service) HandleEvents(ctx context.Context, rawBody []byte) (EventResponse, error) {
	return s.handleEvents(ctx, webhooks.SignedRequest{Body: append([]byte(nil), rawBody...)})
}

// HandleSignedEvents performs Slack-specific filtering before handing durable
// callbacks to the shared webhook gateway. The HTTP adapter supplies the exact
// signed request fields; commands and interactions continue to use their own
// synchronous paths.
func (s *Service) HandleSignedEvents(ctx context.Context, request webhooks.SignedRequest) (EventResponse, error) {
	request.Body = append([]byte(nil), request.Body...)
	return s.handleEvents(ctx, request)
}

func (s *Service) handleEvents(ctx context.Context, request webhooks.SignedRequest) (EventResponse, error) {
	rawBody := request.Body
	payload, err := decodeSlackEvent(rawBody)
	if err != nil {
		return EventResponse{}, fmt.Errorf("%w: %v", ErrSlackInvalidEventPayload, err)
	}
	if payload.Type == "url_verification" {
		return EventResponse{Challenge: payload.Challenge}, nil
	}
	if payload.Type != slackEventCallback {
		return EventResponse{}, nil
	}
	if payload.EventID == "" || payload.TeamID == "" {
		return EventResponse{}, fmt.Errorf("%w: callback is missing event_id or team_id", ErrSlackInvalidEventPayload)
	}
	event, supported := normalizeSlackEvent(payload)
	if !supported {
		if strings.TrimSpace(payload.Event.Type) == slackEventLinkShared && s.log != nil {
			s.log.Warn(ctx, "Slack story preview event ignored at ingress",
				"event_id", payload.EventID,
				"slack_team_id", payload.TeamID,
				"slack_channel_id", firstNonEmptyString(payload.Event.Channel, payload.Event.ChannelID),
				"unfurl_source", strings.TrimSpace(payload.Event.Source),
				"has_unfurl_id", strings.TrimSpace(payload.Event.UnfurlID) != "",
				"has_message_ts", strings.TrimSpace(payload.Event.MessageTS) != "" || strings.TrimSpace(payload.Event.TS) != "",
				"link_count", len(payload.Event.Links),
				"is_external_shared_channel", payload.IsExtSharedChannel,
				"reason", "unsupported_event_shape",
			)
		}
		// message.channels and message.groups are intentionally broad. Discard
		// unrelated roots, bot messages, edits, and unsupported event shapes
		// before encrypting or retaining their content in the durable inbox.
		return EventResponse{}, nil
	}
	if event.Kind == slackEventKindLinkShared && s.log != nil {
		s.log.Info(ctx, "Slack story preview event received",
			"event_id", event.EventID,
			"slack_team_id", event.TeamID,
			"slack_channel_id", event.ChannelID,
			"unfurl_source", slackStoryPreviewSource(event),
			"unfurl_destination", slackStoryPreviewDestination(event),
			"link_count", len(event.Links),
		)
	}
	if event.Kind == slackEventKindEntityDetails {
		workCtx, cancel := s.newSlackWorkObjectTriggerContext(ctx)
		err := s.handleSlackEntityDetailsEvent(workCtx, event)
		cancel()
		if err != nil && s.log != nil {
			s.log.Error(
				context.WithoutCancel(ctx),
				"failed processing Slack entity details within the trigger window",
				"error", err,
				"terminal", isSlackEntityDetailsTerminalError(err),
				"event_id", event.EventID,
				"slack_team_id", event.TeamID,
				"slack_user_id", event.UserID,
			)
		}
		s.dispatchFirstInteractionGuideByTeam(ctx, event.TeamID, event.UserID)
		// entity_details_requested carries a single-use trigger that expires in
		// three seconds. It is intentionally never persisted, queued, or retried;
		// the user can request a fresh trigger by refreshing the flexpane.
		return EventResponse{}, nil
	}
	if s.eventGateway == nil || s.eventInbox == nil {
		return EventResponse{}, ErrSlackEventRuntimeNotConfigured
	}
	installation, installationErr := s.repo.GetSlackWorkspaceByTeamID(ctx, payload.TeamID)
	if isSlackRepositoryNotFound(installationErr) {
		if event.Kind == slackEventKindLinkShared && s.log != nil {
			s.log.Warn(ctx, "Slack story preview event ignored at ingress",
				"event_id", event.EventID,
				"slack_team_id", event.TeamID,
				"slack_channel_id", event.ChannelID,
				"unfurl_source", slackStoryPreviewSource(event),
				"reason", "installation_not_found",
			)
		}
		// A valid Slack signature proves the sender is Slack, not that FortyOne
		// still owns this installation. Disconnected and orphaned installations
		// have no recoverable work, so do not retain their message content.
		return EventResponse{}, nil
	}
	if installationErr != nil {
		return EventResponse{}, fmt.Errorf("resolve Slack installation for event receipt: %w", installationErr)
	}
	if event.Kind == slackEventKindChannelThread {
		if installation.BotUserID != nil && containsSlackUserMention(event.Text, *installation.BotUserID) {
			return EventResponse{}, nil
		}
		linkedUserID, linkErr := s.repo.FindLinkedUserIDBySlackUser(ctx, installation.WorkspaceID, event.TeamID, event.UserID)
		if linkErr != nil {
			return EventResponse{}, fmt.Errorf("resolve Slack thread actor: %w", linkErr)
		}
		subscribed := false
		if linkedUserID != nil && *linkedUserID != uuid.Nil {
			conversation, conversationErr := findSlackConversation(ctx, s.eventInbox, assistantConversationInput(installation.WorkspaceID, *linkedUserID, event))
			switch {
			case conversationErr == nil:
				subscribed = slackThreadSubscriptionIsCurrent(conversation, installation, s.clock.Now())
			case !errors.Is(conversationErr, errMessagingRecordNotFound):
				return EventResponse{}, fmt.Errorf("resolve Slack assistant thread subscription: %w", conversationErr)
			}
		}
		if !subscribed && s.requests != nil {
			subscribed, err = s.requests.HasCurrentProviderThread(ctx, providerThreadLookupInput{
				WorkspaceID:            installation.WorkspaceID,
				Provider:               providerSlack,
				ExternalWorkspaceID:    event.TeamID,
				InstallationGeneration: installation.InstallGeneration,
				ExternalChannelID:      event.ChannelID,
				ExternalThreadID:       event.ThreadTS,
			})
			if err != nil {
				return EventResponse{}, fmt.Errorf("resolve Slack request thread subscription: %w", err)
			}
		}
		if !subscribed {
			return EventResponse{}, nil
		}
	}
	if slackEventCountsAsFirstInteraction(event.Kind) {
		s.dispatchFirstInteractionGuide(ctx, installation, event.UserID)
	}
	receipt, err := s.eventGateway.Receive(ctx, slackWebhookProvider, request)
	if err != nil {
		return EventResponse{}, fmt.Errorf("receive Slack event through durable gateway: %w", err)
	}
	if receipt.Ignored {
		return EventResponse{}, nil
	}
	if event.Kind == slackEventKindLinkShared && s.log != nil {
		s.log.Info(ctx, "Slack story preview event queued",
			"event_id", event.EventID,
			"workspace_id", installation.WorkspaceID,
			"slack_team_id", event.TeamID,
			"slack_channel_id", event.ChannelID,
			"unfurl_source", slackStoryPreviewSource(event),
			"unfurl_destination", slackStoryPreviewDestination(event),
			"link_count", len(event.Links),
			"inbox_id", receipt.ID,
			"created", receipt.Created,
			"queued", receipt.Queued,
		)
	}
	return EventResponse{}, nil
}

package slack

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) handleMessageAction(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	messageAuthorID := strings.TrimSpace(payload.Message.User)
	messageAuthor := messageAuthorID
	if strings.EqualFold(messageAuthorID, strings.TrimSpace(payload.User.ID)) && strings.TrimSpace(payload.User.Username) != "" {
		messageAuthor = strings.TrimSpace(payload.User.Username)
	}

	title := messageToTitle(payload.Message.Text)
	description := buildPrefilledDescription(requestSourceContext{
		SlackUserID:   messageAuthorID,
		SlackUsername: messageAuthor,
		SlackText:     strings.TrimSpace(payload.Message.Text),
	})
	source := requestSourceContext{
		SlackTeamID:     strings.TrimSpace(payload.Team.ID),
		SlackTeamDomain: strings.TrimSpace(payload.Team.Domain),
		SlackChannelID:  strings.TrimSpace(payload.Channel.ID),
		SlackChannel:    strings.TrimSpace(payload.Channel.Name),
		SlackMessageTS:  strings.TrimSpace(payload.Message.TS),
		SlackThreadTS:   strings.TrimSpace(payload.Message.ThreadTS),
		SlackUserID:     strings.TrimSpace(payload.User.ID),
		SlackUsername:   strings.TrimSpace(payload.User.Username),
		SlackText:       strings.TrimSpace(payload.Message.Text),
		ResponseURL:     strings.TrimSpace(payload.ResponseURL),
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return InteractionResponse{}, ErrSlackNoWorkspaceLinked
		}
		return InteractionResponse{}, err
	}

	linkedUserID, connectURL, err := s.resolveLinkedSlackUser(ctx, slackWorkspace.WorkspaceID, source)
	if err != nil {
		return InteractionResponse{}, err
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return InteractionResponse{}, err
	}
	if linkedUserID == uuid.Nil {
		message := buildConnectSlackAccountMessage(connectURL)
		if responseURL := strings.TrimSpace(payload.ResponseURL); responseURL != "" {
			if responseErr := s.postCommandResponse(ctx, responseURL, message); responseErr == nil {
				return InteractionResponse{StatusCode: http.StatusOK}, nil
			} else {
				s.log.Warn(ctx, "failed posting slack connect prompt via response_url", "error", responseErr)
			}
		}
		if responseErr := s.postEphemeralMessage(ctx, botToken, source.SlackChannelID, source.SlackUserID, message); responseErr != nil {
			return InteractionResponse{}, fmt.Errorf("post Slack account connection prompt: %w", responseErr)
		}
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}

	if err := s.openCreateTaskModal(ctx, payload.TriggerID, title, description, source, slackWorkspace.WorkspaceID, linkedUserID, botToken); err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK}, nil
}

func (s *Service) updateSlackInteractiveMessage(ctx context.Context, botToken, channelID, messageTS, text string) error {
	return s.updateSlackInteractiveMessageWithProviderPayload(
		ctx,
		botToken,
		channelID,
		messageTS,
		text,
		SlackProviderPayload{},
	)
}

func (s *Service) updateSlackInteractiveMessageWithProviderPayload(
	ctx context.Context,
	botToken, channelID, messageTS, text string,
	providerPayload SlackProviderPayload,
) error {
	blocks := providerPayload.Blocks
	if blocks == nil {
		blocks = []SlackBlock{}
	}
	payload := map[string]any{
		"channel": strings.TrimSpace(channelID),
		"ts":      strings.TrimSpace(messageTS),
		"text":    truncateSlackText(text),
		"blocks":  blocks,
	}
	return s.slackClient().callJSON(ctx, botToken, "chat.update", payload, nil)
}

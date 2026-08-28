package slack

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) HandleCommand(ctx context.Context, rawBody []byte) (CommandResponse, error) {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return CommandResponse{}, err
	}
	if values.Get("ssl_check") == "1" {
		return CommandResponse{}, nil
	}

	triggerID := strings.TrimSpace(values.Get("trigger_id"))
	if triggerID == "" {
		return CommandResponse{}, errors.New("missing trigger_id")
	}

	source := requestSourceContext{
		SlackTeamID:     strings.TrimSpace(values.Get("team_id")),
		SlackTeamDomain: strings.TrimSpace(values.Get("team_domain")),
		SlackChannelID:  strings.TrimSpace(values.Get("channel_id")),
		SlackChannel:    strings.TrimSpace(values.Get("channel_name")),
		SlackUserID:     strings.TrimSpace(values.Get("user_id")),
		SlackUsername:   strings.TrimSpace(values.Get("user_name")),
		SlackText:       strings.TrimSpace(values.Get("text")),
		ResponseURL:     strings.TrimSpace(values.Get("response_url")),
	}
	if strings.EqualFold(strings.TrimSpace(source.SlackText), "disconnect") {
		disconnected, disconnectErr := s.disconnectSlackAccountBySource(ctx, source)
		if disconnectErr != nil {
			return CommandResponse{}, disconnectErr
		}
		message := "Your Slack account is already disconnected from FortyOne."
		if disconnected {
			message = "Your Slack account has been disconnected from FortyOne."
		}
		return CommandResponse{ResponseType: "ephemeral", Text: message}, nil
	}
	title := parseCommandTitle(values.Get("text"))
	s.dispatchCommand(ctx, triggerID, title, source)
	s.dispatchFirstInteractionGuideByTeam(ctx, source.SlackTeamID, source.SlackUserID)

	// A successful slash command opens a modal and needs no conversational
	// acknowledgement. Returning an empty response prevents Slack from adding a
	// noisy ephemeral "Opening..." message to the channel.
	return CommandResponse{}, nil
}

func (s *Service) dispatchCommand(parent context.Context, triggerID, title string, source requestSourceContext) {
	baseCtx := context.WithoutCancel(parent)
	go func() {
		workCtx, cancel := context.WithTimeout(baseCtx, slackInteractiveWorkTimeout)
		feedback, err := s.processCommand(workCtx, triggerID, title, source)
		cancel()
		if err != nil {
			s.log.Error(baseCtx, "failed processing slack command", "error", err, "slack_team_id", source.SlackTeamID, "slack_user_id", source.SlackUserID)
			feedback = "Unable to open the FortyOne create story form. Please try again."
		}
		if strings.TrimSpace(feedback) == "" {
			return
		}
		if strings.TrimSpace(source.ResponseURL) == "" {
			s.log.Warn(baseCtx, "cannot post slack command feedback without a response URL", "slack_team_id", source.SlackTeamID, "slack_user_id", source.SlackUserID)
			return
		}

		feedbackCtx, feedbackCancel := context.WithTimeout(baseCtx, slackFailureFeedbackTimeout)
		defer feedbackCancel()
		if notifyErr := s.postCommandResponse(feedbackCtx, source.ResponseURL, feedback); notifyErr != nil {
			s.log.Error(baseCtx, "failed posting slack command feedback", "error", notifyErr, "slack_team_id", source.SlackTeamID, "slack_user_id", source.SlackUserID)
		}
	}()
}

func (s *Service) processCommand(ctx context.Context, triggerID, title string, source requestSourceContext) (string, error) {
	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return "Slack is not connected to this FortyOne workspace.", nil
		}
		return "", err
	}

	linkedUserID, connectURL, err := s.resolveLinkedSlackUser(ctx, slackWorkspace.WorkspaceID, source)
	if err != nil {
		return "", err
	}
	if linkedUserID == uuid.Nil {
		return buildConnectSlackAccountMessage(connectURL), nil
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return "", err
	}
	if err := s.openCreateTaskModal(ctx, triggerID, title, "", source, slackWorkspace.WorkspaceID, linkedUserID, botToken); err != nil {
		return "", err
	}
	return "", nil
}

func parseCommandTitle(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "New task"
	}
	parts := strings.Fields(trimmed)
	if len(parts) >= 2 && strings.EqualFold(parts[0], "create") && strings.EqualFold(parts[1], "task") {
		parts = parts[2:]
	}
	if len(parts) == 0 {
		return "New task"
	}
	return truncateRunes(strings.TrimSpace(strings.Join(parts, " ")), modalTitleMaxRunes)
}

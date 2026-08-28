package slack

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func (p *EventProcessor) publishSlackStoryUnfurl(
	ctx context.Context,
	event normalizedSlackEvent,
	workspaceID uuid.UUID,
	request SlackChatUnfurlRequest,
	entityCount int,
	authRequired bool,
	botToken string,
) error {
	if p.log != nil {
		p.log.Info(ctx, "Publishing Slack story preview",
			"event_id", event.EventID,
			"workspace_id", workspaceID,
			"slack_team_id", event.TeamID,
			"slack_channel_id", event.ChannelID,
			"slack_user_id", event.UserID,
			"unfurl_source", slackStoryPreviewSource(event),
			"unfurl_destination", slackStoryPreviewDestination(event),
			"entity_count", entityCount,
			"authentication_required", authRequired,
		)
	}
	err := p.workObjects.Unfurl(ctx, botToken, request)
	if err != nil {
		if p.log != nil {
			fields := []any{
				"error", err,
				"event_id", event.EventID,
				"workspace_id", workspaceID,
				"slack_team_id", event.TeamID,
				"slack_channel_id", event.ChannelID,
				"unfurl_source", slackStoryPreviewSource(event),
				"unfurl_destination", slackStoryPreviewDestination(event),
				"entity_count", entityCount,
				"authentication_required", authRequired,
			}
			if code, ok := SlackAPIErrorCode(err); ok {
				fields = append(fields, "slack_error_code", code)
			}
			if retryAfter, ok := SlackRetryAfter(err); ok {
				fields = append(fields, "retry_after", retryAfter.String())
			}
			p.log.Error(ctx, "Slack story preview publish failed", fields...)
		}
		return err
	}
	if p.log != nil {
		p.log.Info(ctx, "Slack story preview published",
			"event_id", event.EventID,
			"workspace_id", workspaceID,
			"slack_team_id", event.TeamID,
			"slack_channel_id", event.ChannelID,
			"unfurl_source", slackStoryPreviewSource(event),
			"unfurl_destination", slackStoryPreviewDestination(event),
			"entity_count", entityCount,
			"authentication_required", authRequired,
		)
	}
	return nil
}

func slackStoryPreviewSource(event normalizedSlackEvent) string {
	source := strings.ToLower(strings.TrimSpace(event.Source))
	if source == "" {
		return "conversations_history"
	}
	return source
}

func slackStoryPreviewDestination(event normalizedSlackEvent) string {
	if slackStoryPreviewSource(event) == "composer" {
		return "composer"
	}
	return "message"
}

func validSlackUnfurlEventDestination(event normalizedSlackEvent) bool {
	if strings.EqualFold(strings.TrimSpace(event.Source), "composer") {
		return strings.TrimSpace(event.UnfurlID) != ""
	}
	return strings.TrimSpace(event.ChannelID) != "" && strings.TrimSpace(event.MessageTS) != ""
}

func applySlackUnfurlEventDestination(request *SlackChatUnfurlRequest, event normalizedSlackEvent) {
	if request == nil || !strings.EqualFold(strings.TrimSpace(event.Source), "composer") {
		return
	}
	request.Channel = ""
	request.TS = ""
	request.UnfurlID = strings.TrimSpace(event.UnfurlID)
	request.Source = "composer"
}

func (p *EventProcessor) slackStoryAccessGranted(
	ctx context.Context,
	workspaceID uuid.UUID,
	installation slackWorkspaceRecord,
	userID uuid.UUID,
	channelID string,
	teamID uuid.UUID,
) (bool, error) {
	repository, ok := p.repo.(slackWorkObjectRepository)
	if !ok {
		return false, errors.New("slack Work Object repository is not configured")
	}
	channelID = strings.TrimSpace(channelID)
	if channelID == "" || strings.HasPrefix(strings.ToUpper(channelID), "D") {
		_, err := repository.FindTeamMemberByID(ctx, teamID, userID)
		if err != nil {
			if isSlackRepositoryNotFound(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}
	teamIDs, err := repository.ListAuthorizedChannelTeamIDs(ctx, workspaceID, installation.ID, channelID, userID)
	if err != nil {
		return false, err
	}
	for _, authorizedTeamID := range teamIDs {
		if authorizedTeamID == teamID {
			return true, nil
		}
	}
	return false, nil
}

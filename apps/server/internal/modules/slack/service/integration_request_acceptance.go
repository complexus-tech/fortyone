package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) AcceptIntegrationRequest(ctx context.Context, request integrationRequest, story singleStory) error {
	if request.Provider != providerSlack {
		return nil
	}
	channelID := metadataString(request.Metadata, "slack_channel_id")
	threadTS := metadataString(request.Metadata, "slack_thread_ts")
	if threadTS == "" {
		threadTS = metadataString(request.Metadata, "slack_message_ts")
	}
	providerThread, threadErr := s.requests.FindProviderThread(ctx, request.WorkspaceID, request.ID, providerSlack)
	if threadErr == nil {
		channelID = strings.TrimSpace(providerThread.ExternalChannelID)
		threadTS = strings.TrimSpace(providerThread.ExternalThreadID)
	} else if !errors.Is(threadErr, errProviderThreadNotFound) {
		return fmt.Errorf("find canonical Slack request thread: %w", threadErr)
	}
	if channelID == "" || threadTS == "" {
		return nil
	}

	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return nil
		}
		return err
	}
	requestSlackTeamID := metadataString(request.Metadata, "slack_team_id")
	if requestSlackTeamID == "" || requestSlackTeamID != slackWorkspace.SlackTeamID {
		s.log.Warn(ctx, "skipping Slack acceptance update for a replaced installation", "request_id", request.ID, "request_slack_team_id", requestSlackTeamID, "active_slack_team_id", slackWorkspace.SlackTeamID)
		return nil
	}
	if threadErr == nil && (providerThread.InstallationGeneration == nil || *providerThread.InstallationGeneration != slackWorkspace.InstallGeneration || providerThread.ExternalWorkspaceID != slackWorkspace.SlackTeamID) {
		s.log.Warn(ctx, "skipping Slack acceptance update for a stale provider thread", "request_id", request.ID)
		return nil
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return err
	}
	workspaceSlug := metadataString(request.Metadata, "workspace_slug")
	teamCode := strings.TrimSpace(story.TeamCode)
	if teamCode == "" {
		teamCode = metadataString(request.Metadata, "team_code")
	}
	creatorName := "A team member"
	if story.Reporter != nil && *story.Reporter != uuid.Nil {
		creator, creatorErr := s.repo.FindTeamMemberByID(ctx, request.TeamID, *story.Reporter)
		if creatorErr != nil {
			return fmt.Errorf("find accepted Slack request story creator: %w", creatorErr)
		}
		creatorName = storyCreatorDisplayName(creator)
	}
	s.postSlackTaskAck(
		ctx,
		request.WorkspaceID,
		slackWorkspace.InstallGeneration,
		fmt.Sprintf("slack:%s:request:%s:accepted", request.WorkspaceID, request.ID),
		requestSourceContext{
			SlackTeamID:    slackWorkspace.SlackTeamID,
			SlackChannelID: channelID,
			SlackThreadTS:  threadTS,
			SlackUserID:    metadataString(request.Metadata, "slack_user_id"),
		},
		botToken,
		workspaceSlug,
		teamCode,
		creatorName,
		"",
		slackStoryReceiptActionLinkedRequest,
		story,
	)
	return nil
}

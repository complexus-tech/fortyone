package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type slackStoryReceiptAction string

const (
	slackStoryReceiptActionCreated       slackStoryReceiptAction = "created"
	slackStoryReceiptActionLinkedRequest slackStoryReceiptAction = "linked_request"
)

func (s *Service) postSlackRequestAck(ctx context.Context, workspaceID, installGeneration uuid.UUID, idempotencyKey string, source requestSourceContext, botToken, workspaceSlug string, request integrationRequest, actorID uuid.UUID, creatorName string) string {
	requestURL := buildRequestURL(s.cfg.WebsiteURL, workspaceSlug, request.TeamID.String(), request.ID.String())
	text := buildSlackRequestOpenedText(creatorName, requestURL)
	disableUnfurls := false
	providerPayload := SlackProviderPayload{
		UnfurlLinks: &disableUnfurls,
		UnfurlMedia: &disableUnfurls,
		Authorization: &SlackDeliveryAuthorization{
			AllowedTeamIDs: []uuid.UUID{request.TeamID},
			ActorUserID:    &actorID,
			Scope:          slackDeliveryAuthorizationScopeActorMembership,
		},
		RequestThreadBinding: &SlackRequestThreadBinding{
			IntegrationRequestID:    request.ID,
			ExternalSourceMessageID: strings.TrimSpace(source.SlackMessageTS),
			SourceURL:               permalinkFromContext(source),
		},
	}
	receipt, err := s.buildSlackRequestCreationReceipt(ctx, actorID, source.SlackUserID, requestURL, creatorName, request)
	if err != nil {
		if s.log != nil {
			s.log.Warn(ctx, "failed building rich Slack request receipt", "error", err, "workspace_id", workspaceID, "request_id", request.ID)
		}
	} else {
		providerPayload.Metadata = receipt.ProviderPayload.Metadata
		providerPayload.UnfurlLinks = receipt.ProviderPayload.UnfurlLinks
		providerPayload.UnfurlMedia = receipt.ProviderPayload.UnfurlMedia
	}
	return s.postSlackCreationAckWithPayload(ctx, workspaceID, installGeneration, idempotencyKey, source, botToken, text, providerPayload)
}

func (s *Service) postSlackTaskAck(ctx context.Context, workspaceID, installGeneration uuid.UUID, idempotencyKey string, source requestSourceContext, botToken, workspaceSlug, teamCode, creatorName, creatorSlackUserID string, action slackStoryReceiptAction, story singleStory) string {
	storyCode := buildStoryCode(teamCode, story.SequenceID)
	taskURL := buildTaskURL(
		s.cfg.WebsiteURL,
		workspaceSlug,
		buildStoryReference(teamCode, story.SequenceID, story.ID.String()),
	)
	text := buildSlackStoryReceiptText(action, creatorName, storyCode, taskURL)
	authorization := &SlackDeliveryAuthorization{
		AllowedTeamIDs: []uuid.UUID{story.Team},
		Scope:          slackDeliveryAuthorizationScopeActorMembership,
	}
	if story.Reporter != nil && *story.Reporter != uuid.Nil {
		authorization.ActorUserID = story.Reporter
	} else if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(source.SlackChannelID)), "D") {
		linkedUserID, linkErr := s.repo.FindLinkedUserIDBySlackUser(ctx, workspaceID, strings.TrimSpace(source.SlackTeamID), strings.TrimSpace(source.SlackUserID))
		if linkErr != nil {
			s.log.Warn(ctx, "failed resolving Slack receipt recipient", "error", linkErr, "workspace_id", workspaceID, "story_id", story.ID)
		} else if linkedUserID != nil && *linkedUserID != uuid.Nil {
			authorization.ActorUserID = linkedUserID
		}
	}
	receipt, err := s.buildSlackStoryCreationReceipt(ctx, creatorSlackUserID, taskURL, creatorName, story)
	if err != nil {
		if s.log != nil {
			s.log.Warn(ctx, "failed building rich Slack story receipt", "error", err, "workspace_id", workspaceID, "story_id", story.ID)
		}
		disableUnfurls := false
		return s.postSlackCreationAckWithPayload(ctx, workspaceID, installGeneration, idempotencyKey, source, botToken, text, SlackProviderPayload{
			UnfurlLinks:   &disableUnfurls,
			UnfurlMedia:   &disableUnfurls,
			Authorization: authorization,
		})
	}
	receipt.Text = text
	receipt.ProviderPayload.Authorization = authorization
	return s.postSlackCreationAckWithPayload(ctx, workspaceID, installGeneration, idempotencyKey, source, botToken, receipt.Text, receipt.ProviderPayload)
}

func buildSlackStoryCreatedText(creatorName, storyCode, taskURL string) string {
	return buildSlackStoryReceiptText(slackStoryReceiptActionCreated, creatorName, storyCode, taskURL)
}

func buildSlackStoryLinkedRequestText(creatorName, storyCode, taskURL string) string {
	return buildSlackStoryReceiptText(slackStoryReceiptActionLinkedRequest, creatorName, storyCode, taskURL)
}

func buildSlackStoryReceiptText(action slackStoryReceiptAction, creatorName, storyCode, taskURL string) string {
	creatorLabel := slackMrkdwnTextEscaper.Replace(strings.TrimSpace(creatorName))
	if creatorLabel == "" {
		creatorLabel = "A team member"
	}
	storyLabel := strings.TrimSpace(storyCode)
	if storyLabel == "" {
		storyLabel = "a story"
	}
	if taskURL = strings.TrimSpace(taskURL); taskURL != "" {
		storyLabel = fmt.Sprintf("<%s|%s>", taskURL, storyLabel)
	}
	if action == slackStoryReceiptActionLinkedRequest {
		return fmt.Sprintf("%s linked a request to %s", creatorLabel, storyLabel)
	}
	return fmt.Sprintf("%s created %s", creatorLabel, storyLabel)
}

func buildSlackRequestOpenedText(creatorName, requestURL string) string {
	creatorLabel := slackMrkdwnTextEscaper.Replace(strings.TrimSpace(creatorName))
	if creatorLabel == "" {
		creatorLabel = "A team member"
	}
	requestLabel := "opened a request"
	if requestURL = strings.TrimSpace(requestURL); requestURL != "" {
		requestLabel = fmt.Sprintf("<%s|%s>", requestURL, requestLabel)
	}
	return fmt.Sprintf("%s %s", creatorLabel, requestLabel)
}

func (s *Service) buildSlackRequestCreationReceipt(
	ctx context.Context,
	actorID uuid.UUID,
	actorSlackUserID, requestURL, creatorName string,
	request integrationRequest,
) (SlackRequestCreationReceipt, error) {
	repository, ok := s.repo.(slackWorkObjectRepository)
	if !ok {
		return SlackRequestCreationReceipt{}, errors.New("slack Work Object repository is not configured")
	}
	input, err := buildSlackRequestWorkObjectInput(ctx, repository, actorID, actorSlackUserID, requestURL, request)
	if err != nil {
		return SlackRequestCreationReceipt{}, err
	}
	return BuildSlackRequestCreationReceipt(creatorName, input)
}

func (s *Service) buildSlackStoryCreationReceipt(
	ctx context.Context,
	creatorSlackUserID string,
	storyURL, creatorName string,
	story singleStory,
) (SlackStoryCreationReceipt, error) {
	statusName := ""
	if story.Status != nil {
		statuses, err := s.repo.ListTeamStatuses(ctx, story.Team)
		if err != nil {
			return SlackStoryCreationReceipt{}, err
		}
		for _, status := range statuses {
			if status.ID == *story.Status {
				statusName = status.Name
				break
			}
		}
	}
	assigneeName := ""
	assigneeSlackUserID := ""
	if story.Assignee != nil {
		member, err := s.repo.FindTeamMemberByID(ctx, story.Team, *story.Assignee)
		if err != nil && !isSlackRepositoryNotFound(err) {
			return SlackStoryCreationReceipt{}, err
		}
		if err == nil {
			assigneeName = slackMemberDisplayName(member)
		}
		if story.Reporter != nil && *story.Reporter == *story.Assignee {
			assigneeSlackUserID = strings.TrimSpace(creatorSlackUserID)
		}
	}
	description := ""
	if story.Description != nil {
		description = strings.TrimSpace(*story.Description)
	}
	return BuildSlackStoryCreationReceipt(creatorName, SlackStoryWorkObjectInput{
		AccessGranted:       true,
		ExternalID:          story.ID.String(),
		StoryURL:            storyURL,
		Title:               story.Title,
		Description:         description,
		Status:              statusName,
		Priority:            story.Priority,
		AssigneeSlackUserID: assigneeSlackUserID,
		AssigneeName:        assigneeName,
		CreatorSlackUserID:  strings.TrimSpace(creatorSlackUserID),
		CreatorName:         creatorName,
		DueDate:             story.EndDate,
		CreatedAt:           story.CreatedAt,
		UpdatedAt:           story.UpdatedAt,
	})
}

func slackMemberDisplayName(member slackTeamMemberRecord) string {
	for _, value := range []string{member.FullName, member.Username, member.Email} {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

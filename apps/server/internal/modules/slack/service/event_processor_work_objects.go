package slack

import (
	"context"
	"errors"
	"strings"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

// SlackStoryReader is deliberately read-only. Work Object events never mutate
// a story and never bypass the actor/channel authorization checks below.
type SlackStoryReader interface {
	QueryByRef(ctx context.Context, workspaceID uuid.UUID, storyRef string) (stories.CoreSingleStory, error)
}

// SlackRequestReader is deliberately permission-aware and read-only. Its
// contract requires the current FortyOne actor so request previews cannot
// bypass workspace or team membership.
type SlackRequestReader interface {
	GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error)
}

type slackWorkObjectRepository interface {
	ListAuthorizedChannelTeamIDs(
		ctx context.Context,
		workspaceID, slackWorkspaceID uuid.UUID,
		slackChannelID string,
		userID uuid.UUID,
	) ([]uuid.UUID, error)
	ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]slackrepository.StatusRecord, error)
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]slackrepository.TeamMemberRecord, error)
	FindTeamMemberByID(ctx context.Context, teamID, userID uuid.UUID) (slackrepository.TeamMemberRecord, error)
}

func isSlackWorkObjectEvent(kind slackEventKind) bool {
	return kind == slackEventKindLinkShared
}

func (p *EventProcessor) processSlackWorkObjectEvent(
	ctx context.Context,
	workspace slackrepository.WorkspaceRecord,
	installation slackrepository.SlackWorkspaceRecord,
	linkedUserID *uuid.UUID,
	event normalizedSlackEvent,
	botToken string,
) error {
	if (p.storyReader == nil && p.requestReader == nil) || p.workObjects == nil {
		return errors.New("Slack Work Object runtime is not configured")
	}
	switch event.Kind {
	case slackEventKindLinkShared:
		return p.processSlackLinkShared(ctx, workspace, installation, linkedUserID, event, botToken)
	default:
		return nil
	}
}

func (p *EventProcessor) processSlackLinkShared(
	ctx context.Context,
	workspace slackrepository.WorkspaceRecord,
	installation slackrepository.SlackWorkspaceRecord,
	linkedUserID *uuid.UUID,
	event normalizedSlackEvent,
	botToken string,
) error {
	if p.log != nil {
		p.log.Info(ctx, "Slack work item preview processing started",
			"event_id", event.EventID,
			"workspace_id", workspace.ID,
			"workspace_slug", workspace.Slug,
			"slack_team_id", event.TeamID,
			"slack_channel_id", event.ChannelID,
			"slack_user_id", event.UserID,
			"unfurl_source", slackStoryPreviewSource(event),
			"unfurl_destination", slackStoryPreviewDestination(event),
			"link_count", len(event.Links),
		)
	}
	storyLinks := make([]FortyOneStoryLink, 0, len(event.Links))
	requestLinks := make([]FortyOneRequestLink, 0, len(event.Links))
	seen := make(map[string]struct{}, len(event.Links))
	for _, shared := range event.Links {
		storyLink, storyErr := ParseFortyOneStoryURL(shared.URL)
		if storyErr == nil {
			if p.storyReader == nil || !strings.EqualFold(storyLink.WorkspaceSlug, workspace.Slug) {
				continue
			}
			if _, exists := seen[storyLink.CanonicalURL]; exists {
				continue
			}
			seen[storyLink.CanonicalURL] = struct{}{}
			storyLinks = append(storyLinks, storyLink)
			continue
		}

		requestLink, requestErr := ParseFortyOneRequestURL(shared.URL)
		if requestErr == nil {
			if p.requestReader == nil || !strings.EqualFold(requestLink.WorkspaceSlug, workspace.Slug) {
				continue
			}
			if _, exists := seen[requestLink.CanonicalURL]; exists {
				continue
			}
			seen[requestLink.CanonicalURL] = struct{}{}
			requestLinks = append(requestLinks, requestLink)
			continue
		}

		if p.log != nil {
			p.log.Warn(ctx, "Slack work item preview link ignored",
				"event_id", event.EventID,
				"slack_team_id", event.TeamID,
				"slack_channel_id", event.ChannelID,
				"link_domain", strings.TrimSpace(shared.Domain),
				"reason", "invalid_fortyone_url",
			)
		}
	}
	eligibleLinkCount := len(storyLinks) + len(requestLinks)
	if eligibleLinkCount == 0 {
		if p.log != nil {
			p.log.Warn(ctx, "Slack work item preview produced no eligible links",
				"event_id", event.EventID,
				"workspace_id", workspace.ID,
				"workspace_slug", workspace.Slug,
				"slack_team_id", event.TeamID,
				"slack_channel_id", event.ChannelID,
				"unfurl_source", slackStoryPreviewSource(event),
				"received_link_count", len(event.Links),
			)
		}
		return nil
	}
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		if p.log != nil {
			p.log.Info(ctx, "Slack work item preview requires account link",
				"event_id", event.EventID,
				"workspace_id", workspace.ID,
				"slack_team_id", event.TeamID,
				"slack_channel_id", event.ChannelID,
				"slack_user_id", event.UserID,
				"eligible_link_count", eligibleLinkCount,
			)
		}
		authURL, err := p.accountLinkURL(ctx, workspace, event)
		if err != nil {
			return err
		}
		request, err := BuildSlackStoryAuthenticationUnfurlRequest(event.ChannelID, event.MessageTS, authURL)
		if err != nil {
			return err
		}
		applySlackUnfurlEventDestination(&request, event)
		return p.publishSlackStoryUnfurl(ctx, event, workspace.ID, request, 0, true, botToken)
	}

	metadata := SlackWorkObjectMetadata{Entities: make([]SlackWorkObjectEntity, 0, eligibleLinkCount)}
	accessChannelID := event.ChannelID
	if strings.EqualFold(strings.TrimSpace(event.Source), "composer") {
		// A composer preview is visible only to the author and is not yet part
		// of a channel audience. Slack sends another event after posting, when
		// the final channel audience is authorized independently.
		accessChannelID = ""
	}
	for _, link := range storyLinks {
		story, err := p.storyReader.QueryByRef(ctx, workspace.ID, link.StoryReference)
		if err != nil {
			if errors.Is(err, stories.ErrNotFound) || errors.Is(err, stories.ErrInvalidStoryReference) || slackrepository.IsNotFound(err) {
				if p.log != nil {
					p.log.Warn(ctx, "Slack story preview story not found",
						"event_id", event.EventID,
						"workspace_id", workspace.ID,
						"story_reference", link.StoryReference,
					)
				}
				continue
			}
			return err
		}
		accessGranted, err := p.slackStoryAccessGranted(ctx, workspace.ID, installation, *linkedUserID, accessChannelID, story.Team)
		if err != nil {
			return err
		}
		if !accessGranted {
			if p.log != nil {
				p.log.Warn(ctx, "Slack story preview access denied",
					"event_id", event.EventID,
					"workspace_id", workspace.ID,
					"user_id", *linkedUserID,
					"story_reference", link.StoryReference,
					"story_team_id", story.Team,
					"slack_channel_id", event.ChannelID,
					"unfurl_source", slackStoryPreviewSource(event),
				)
			}
			// A linked but unauthorized actor gets no card and no indication that
			// the reference exists.
			continue
		}
		input, err := p.slackStoryWorkObjectInput(ctx, *linkedUserID, event.UserID, link.PostedURL, story, false)
		if err != nil {
			return err
		}
		request, err := BuildSlackStoryUnfurlRequest(event.ChannelID, event.MessageTS, input)
		if err != nil {
			return err
		}
		metadata.Entities = append(metadata.Entities, request.Metadata.Entities...)
	}
	for _, link := range requestLinks {
		request, err := p.requestReader.GetForUser(ctx, workspace.ID, link.RequestID, *linkedUserID)
		if err != nil {
			if slackrepository.IsNotFound(err) {
				continue
			}
			return err
		}
		if request.WorkspaceID != workspace.ID || request.TeamID != link.TeamID || request.ID != link.RequestID {
			continue
		}
		accessGranted, err := p.slackStoryAccessGranted(ctx, workspace.ID, installation, *linkedUserID, accessChannelID, request.TeamID)
		if err != nil {
			return err
		}
		if !accessGranted {
			continue
		}
		input, err := p.slackRequestWorkObjectInput(ctx, *linkedUserID, event.UserID, link.PostedURL, request)
		if err != nil {
			return err
		}
		unfurl, err := BuildSlackRequestUnfurlRequest(event.ChannelID, event.MessageTS, input)
		if err != nil {
			return err
		}
		metadata.Entities = append(metadata.Entities, unfurl.Metadata.Entities...)
	}
	if len(metadata.Entities) == 0 {
		if p.log != nil {
			p.log.Warn(ctx, "Slack work item preview produced no authorized entities",
				"event_id", event.EventID,
				"workspace_id", workspace.ID,
				"slack_team_id", event.TeamID,
				"slack_channel_id", event.ChannelID,
				"eligible_link_count", eligibleLinkCount,
			)
		}
		return nil
	}
	request := SlackChatUnfurlRequest{
		Channel:  event.ChannelID,
		TS:       event.MessageTS,
		Metadata: &metadata,
	}
	applySlackUnfurlEventDestination(&request, event)
	return p.publishSlackStoryUnfurl(ctx, event, workspace.ID, request, len(metadata.Entities), false, botToken)
}

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
	installation slackrepository.SlackWorkspaceRecord,
	userID uuid.UUID,
	channelID string,
	teamID uuid.UUID,
) (bool, error) {
	repository, ok := p.repo.(slackWorkObjectRepository)
	if !ok {
		return false, errors.New("Slack Work Object repository is not configured")
	}
	channelID = strings.TrimSpace(channelID)
	if channelID == "" || strings.HasPrefix(strings.ToUpper(channelID), "D") {
		_, err := repository.FindTeamMemberByID(ctx, teamID, userID)
		if err != nil {
			if slackrepository.IsNotFound(err) {
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

func (p *EventProcessor) slackStoryWorkObjectInput(
	ctx context.Context,
	actorID uuid.UUID,
	actorSlackUserID, storyURL string,
	story stories.CoreSingleStory,
	editable bool,
) (SlackStoryWorkObjectInput, error) {
	repository, ok := p.repo.(slackWorkObjectRepository)
	if !ok {
		return SlackStoryWorkObjectInput{}, errors.New("Slack Work Object repository is not configured")
	}
	return buildSlackStoryWorkObjectInput(ctx, repository, actorID, actorSlackUserID, storyURL, story, editable)
}

func (p *EventProcessor) slackRequestWorkObjectInput(
	ctx context.Context,
	actorID uuid.UUID,
	actorSlackUserID, requestURL string,
	request integrationrequests.CoreIntegrationRequest,
) (SlackRequestWorkObjectInput, error) {
	repository, ok := p.repo.(slackWorkObjectRepository)
	if !ok {
		return SlackRequestWorkObjectInput{}, errors.New("Slack Work Object repository is not configured")
	}
	return buildSlackRequestWorkObjectInput(ctx, repository, actorID, actorSlackUserID, requestURL, request)
}

func buildSlackStoryWorkObjectInput(
	ctx context.Context,
	repository slackWorkObjectRepository,
	actorID uuid.UUID,
	actorSlackUserID, storyURL string,
	story stories.CoreSingleStory,
	editable bool,
) (SlackStoryWorkObjectInput, error) {
	statusName, statusID := "", ""
	statusOptions := make([]SlackWorkObjectSelectOption, 0)
	if story.Status != nil || editable {
		statuses, err := repository.ListTeamStatuses(ctx, story.Team)
		if err != nil {
			return SlackStoryWorkObjectInput{}, err
		}
		for _, status := range statuses {
			if story.Status != nil && status.ID == *story.Status {
				statusName = status.Name
				statusID = status.ID.String()
			}
			if editable {
				statusOptions = append(statusOptions, newSlackWorkObjectSelectOption(status.ID.String(), status.Name))
			}
		}
		if !validSlackWorkObjectSelectOptions(statusOptions) {
			statusOptions = nil
		}
	}
	assigneeName, assigneeID, assigneeSlackUserID := "", "", ""
	assigneeOptions := make([]SlackWorkObjectSelectOption, 0)
	if story.Assignee != nil || editable {
		members, err := repository.ListTeamMembers(ctx, story.Team)
		if err != nil {
			return SlackStoryWorkObjectInput{}, err
		}
		for _, member := range members {
			if story.Assignee != nil && member.UserID == *story.Assignee {
				assigneeID = member.UserID.String()
				assigneeName = slackMemberDisplayName(member)
				if member.UserID == actorID {
					assigneeSlackUserID = actorSlackUserID
				}
			}
			if editable {
				assigneeOptions = append(assigneeOptions, newSlackWorkObjectSelectOption(member.UserID.String(), slackMemberDisplayName(member)))
			}
		}
		if !validSlackWorkObjectSelectOptions(assigneeOptions) {
			assigneeOptions = nil
		}
	}
	creatorName, creatorSlackUserID := "", ""
	if story.Reporter != nil {
		member, err := repository.FindTeamMemberByID(ctx, story.Team, *story.Reporter)
		if err != nil && !slackrepository.IsNotFound(err) {
			return SlackStoryWorkObjectInput{}, err
		}
		if err == nil {
			creatorName = slackMemberDisplayName(member)
		}
		if *story.Reporter == actorID {
			creatorSlackUserID = actorSlackUserID
		}
	}
	description := ""
	if story.Description != nil {
		description = *story.Description
	}
	return SlackStoryWorkObjectInput{
		AccessGranted:       true,
		Editable:            editable,
		ExternalID:          story.ID.String(),
		StoryURL:            storyURL,
		Title:               story.Title,
		Description:         description,
		Status:              statusName,
		StatusID:            statusID,
		StatusOptions:       statusOptions,
		Priority:            story.Priority,
		AssigneeID:          assigneeID,
		AssigneeOptions:     assigneeOptions,
		AssigneeSlackUserID: assigneeSlackUserID,
		AssigneeName:        assigneeName,
		CreatorSlackUserID:  creatorSlackUserID,
		CreatorName:         creatorName,
		DueDate:             story.EndDate,
		CreatedAt:           story.CreatedAt,
		UpdatedAt:           story.UpdatedAt,
	}, nil
}

func buildSlackRequestWorkObjectInput(
	ctx context.Context,
	repository slackWorkObjectRepository,
	actorID uuid.UUID,
	actorSlackUserID, requestURL string,
	request integrationrequests.CoreIntegrationRequest,
) (SlackRequestWorkObjectInput, error) {
	assigneeName, assigneeSlackUserID := "", ""
	if request.AssigneeID != nil && *request.AssigneeID != uuid.Nil {
		member, err := repository.FindTeamMemberByID(ctx, request.TeamID, *request.AssigneeID)
		if err != nil && !slackrepository.IsNotFound(err) {
			return SlackRequestWorkObjectInput{}, err
		}
		if err == nil {
			assigneeName = slackMemberDisplayName(member)
		}
		if *request.AssigneeID == actorID {
			assigneeSlackUserID = strings.TrimSpace(actorSlackUserID)
		}
	}

	creatorName, creatorSlackUserID := "", ""
	if request.CreatedByUserID != nil && *request.CreatedByUserID != uuid.Nil {
		member, err := repository.FindTeamMemberByID(ctx, request.TeamID, *request.CreatedByUserID)
		if err != nil && !slackrepository.IsNotFound(err) {
			return SlackRequestWorkObjectInput{}, err
		}
		if err == nil {
			creatorName = slackMemberDisplayName(member)
		}
		if *request.CreatedByUserID == actorID {
			creatorSlackUserID = strings.TrimSpace(actorSlackUserID)
		}
	}

	description := ""
	if request.Description != nil {
		description = *request.Description
	}
	return SlackRequestWorkObjectInput{
		AccessGranted:       true,
		RequestURL:          requestURL,
		Title:               request.Title,
		Description:         description,
		Status:              slackRequestStatusLabel(request.Status),
		Priority:            request.Priority,
		AssigneeSlackUserID: assigneeSlackUserID,
		AssigneeName:        assigneeName,
		CreatorSlackUserID:  creatorSlackUserID,
		CreatorName:         creatorName,
		DueDate:             request.EndDate,
		CreatedAt:           request.CreatedAt,
		UpdatedAt:           request.UpdatedAt,
	}, nil
}

func slackRequestStatusLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case integrationrequests.StatusPending:
		return "Pending"
	case integrationrequests.StatusAccepted:
		return "Accepted"
	case integrationrequests.StatusDeclined:
		return "Declined"
	default:
		return strings.TrimSpace(status)
	}
}

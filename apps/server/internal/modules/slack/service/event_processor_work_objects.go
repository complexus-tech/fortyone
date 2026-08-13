package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
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

// SlackObjectiveReader lists objectives through the actor-aware path so Slack
// previews cannot bypass objective visibility rules.
type SlackObjectiveReader interface {
	List(ctx context.Context, workspaceID, userID uuid.UUID, filters map[string]any) ([]objectives.CoreObjective, error)
}

// SlackSprintReader lists sprints through the actor-aware path so Slack
// previews cannot bypass team membership checks.
type SlackSprintReader interface {
	List(ctx context.Context, workspaceID, userID uuid.UUID, filters map[string]any) ([]sprints.CoreSprint, error)
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
	if (p.storyReader == nil && p.requestReader == nil && p.objectiveReader == nil && p.sprintReader == nil) || p.workObjects == nil {
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
	objectiveLinks := make([]FortyOneObjectiveLink, 0, len(event.Links))
	sprintLinks := make([]FortyOneSprintLink, 0, len(event.Links))
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

		objectiveLink, objectiveErr := ParseFortyOneObjectiveURL(shared.URL)
		if objectiveErr == nil {
			if p.objectiveReader == nil || !strings.EqualFold(objectiveLink.WorkspaceSlug, workspace.Slug) {
				continue
			}
			if _, exists := seen[objectiveLink.CanonicalURL]; exists {
				continue
			}
			seen[objectiveLink.CanonicalURL] = struct{}{}
			objectiveLinks = append(objectiveLinks, objectiveLink)
			continue
		}

		sprintLink, sprintErr := ParseFortyOneSprintURL(shared.URL)
		if sprintErr == nil {
			if p.sprintReader == nil || !strings.EqualFold(sprintLink.WorkspaceSlug, workspace.Slug) {
				continue
			}
			if _, exists := seen[sprintLink.CanonicalURL]; exists {
				continue
			}
			seen[sprintLink.CanonicalURL] = struct{}{}
			sprintLinks = append(sprintLinks, sprintLink)
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
	eligibleLinkCount := len(storyLinks) + len(requestLinks) + len(objectiveLinks) + len(sprintLinks)
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
	for _, link := range objectiveLinks {
		objective, found, err := p.slackObjectiveForUser(ctx, workspace.ID, *linkedUserID, link)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		accessGranted, err := p.slackStoryAccessGranted(ctx, workspace.ID, installation, *linkedUserID, accessChannelID, objective.Team)
		if err != nil {
			return err
		}
		if !accessGranted {
			continue
		}
		input, err := p.slackObjectiveWorkObjectInput(ctx, *linkedUserID, event.UserID, link.PostedURL, objective)
		if err != nil {
			return err
		}
		unfurl, err := BuildSlackObjectiveUnfurlRequest(event.ChannelID, event.MessageTS, input)
		if err != nil {
			return err
		}
		applySlackUnfurlEventDestination(&unfurl, event)
		metadata.Entities = append(metadata.Entities, unfurl.Metadata.Entities...)
	}
	for _, link := range sprintLinks {
		sprint, found, err := p.slackSprintForUser(ctx, workspace.ID, *linkedUserID, link)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		accessGranted, err := p.slackStoryAccessGranted(ctx, workspace.ID, installation, *linkedUserID, accessChannelID, sprint.Team)
		if err != nil {
			return err
		}
		if !accessGranted {
			continue
		}
		input := slackSprintWorkObjectInput(link.PostedURL, sprint)
		unfurl, err := BuildSlackSprintUnfurlRequest(event.ChannelID, event.MessageTS, input)
		if err != nil {
			return err
		}
		applySlackUnfurlEventDestination(&unfurl, event)
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

func (p *EventProcessor) slackObjectiveForUser(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	link FortyOneObjectiveLink,
) (objectives.CoreObjective, bool, error) {
	if p.objectiveReader == nil {
		return objectives.CoreObjective{}, false, nil
	}
	items, err := p.objectiveReader.List(ctx, workspaceID, userID, map[string]any{"objective_id": link.ObjectiveID})
	if err != nil {
		return objectives.CoreObjective{}, false, err
	}
	for _, item := range items {
		if item.ID == link.ObjectiveID && item.Workspace == workspaceID && item.Team == link.TeamID {
			return item, true, nil
		}
	}
	return objectives.CoreObjective{}, false, nil
}

func (p *EventProcessor) slackSprintForUser(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	link FortyOneSprintLink,
) (sprints.CoreSprint, bool, error) {
	if p.sprintReader == nil {
		return sprints.CoreSprint{}, false, nil
	}
	items, err := p.sprintReader.List(ctx, workspaceID, userID, map[string]any{"sprint_id": link.SprintID})
	if err != nil {
		return sprints.CoreSprint{}, false, err
	}
	for _, item := range items {
		if item.ID == link.SprintID && item.Workspace == workspaceID && item.Team == link.TeamID {
			return item, true, nil
		}
	}
	return sprints.CoreSprint{}, false, nil
}

func (p *EventProcessor) slackObjectiveWorkObjectInput(
	ctx context.Context,
	actorID uuid.UUID,
	actorSlackUserID, objectiveURL string,
	objective objectives.CoreObjective,
) (SlackObjectiveWorkObjectInput, error) {
	repository, ok := p.repo.(slackWorkObjectRepository)
	if !ok {
		return SlackObjectiveWorkObjectInput{}, errors.New("Slack Work Object repository is not configured")
	}
	return buildSlackObjectiveWorkObjectInput(ctx, repository, actorID, actorSlackUserID, objectiveURL, objective)
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

func buildSlackObjectiveWorkObjectInput(
	ctx context.Context,
	repository slackWorkObjectRepository,
	actorID uuid.UUID,
	actorSlackUserID, objectiveURL string,
	objective objectives.CoreObjective,
) (SlackObjectiveWorkObjectInput, error) {
	leadName, leadSlackUserID := "", ""
	if objective.LeadUser != nil && *objective.LeadUser != uuid.Nil {
		member, err := repository.FindTeamMemberByID(ctx, objective.Team, *objective.LeadUser)
		if err != nil && !slackrepository.IsNotFound(err) {
			return SlackObjectiveWorkObjectInput{}, err
		}
		if err == nil {
			leadName = slackMemberDisplayName(member)
		}
		if *objective.LeadUser == actorID {
			leadSlackUserID = strings.TrimSpace(actorSlackUserID)
		}
	}

	description := ""
	if objective.Description != nil {
		description = *objective.Description
	}
	health := ""
	if objective.Health != nil {
		health = string(*objective.Health)
	}
	return SlackObjectiveWorkObjectInput{
		AccessGranted:   true,
		ObjectiveURL:    objectiveURL,
		ExternalID:      objective.ID.String(),
		Title:           objective.Name,
		Description:     description,
		Health:          health,
		Progress:        slackWorkProgressLabel(objective.CompletedStories, objective.TotalStories),
		LeadSlackUserID: leadSlackUserID,
		LeadName:        leadName,
		StartDate:       objective.StartDate,
		EndDate:         objective.EndDate,
		CreatedAt:       objective.CreatedAt,
		UpdatedAt:       objective.UpdatedAt,
	}, nil
}

func slackSprintWorkObjectInput(sprintURL string, sprint sprints.CoreSprint) SlackSprintWorkObjectInput {
	return SlackSprintWorkObjectInput{
		AccessGranted: true,
		SprintURL:     sprintURL,
		ExternalID:    sprint.ID.String(),
		Title:         sprint.Name,
		Goal:          stringValue(sprint.Goal),
		Status:        slackSprintStatus(sprint, time.Now().UTC()),
		Progress:      slackWorkProgressLabel(sprint.CompletedStories, sprint.TotalStories),
		StartDate:     &sprint.StartDate,
		EndDate:       &sprint.EndDate,
		CreatedAt:     sprint.CreatedAt,
		UpdatedAt:     sprint.UpdatedAt,
	}
}

func slackWorkProgressLabel(completed, total int) string {
	if total <= 0 {
		return "No stories"
	}
	if completed < 0 {
		completed = 0
	}
	if completed > total {
		completed = total
	}
	return fmt.Sprintf("%d%% (%d/%d stories)", completed*100/total, completed, total)
}

func slackSprintStatus(sprint sprints.CoreSprint, now time.Time) string {
	if sprint.TotalStories > 0 && sprint.CompletedStories >= sprint.TotalStories {
		return "Completed"
	}
	if now.Before(sprint.StartDate) {
		return "Upcoming"
	}
	if now.After(sprint.EndDate) {
		return "Completed"
	}
	return "Active"
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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

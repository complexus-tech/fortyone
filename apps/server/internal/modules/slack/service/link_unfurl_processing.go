package slack

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func (p *EventProcessor) processSlackLinkShared(
	ctx context.Context,
	workspace workspaceRecord,
	installation slackWorkspaceRecord,
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
			if errors.Is(err, errStoryNotFound) || errors.Is(err, errInvalidStoryReference) || isSlackRepositoryNotFound(err) {
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
			if isSlackRepositoryNotFound(err) {
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
		accessGranted, err := p.slackStoryAccessGranted(ctx, workspace.ID, installation, *linkedUserID, accessChannelID, sprint.TeamID)
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

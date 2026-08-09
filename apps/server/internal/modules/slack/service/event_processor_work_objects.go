package slack

import (
	"context"
	"errors"
	"strings"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

// SlackStoryReader is deliberately read-only. Work Object events never mutate
// a story and never bypass the actor/channel authorization checks below.
type SlackStoryReader interface {
	QueryByRef(ctx context.Context, workspaceID uuid.UUID, storyRef string) (stories.CoreSingleStory, error)
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
	if p.storyReader == nil || p.workObjects == nil {
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
	links := make([]FortyOneStoryLink, 0, len(event.Links))
	seen := make(map[string]struct{}, len(event.Links))
	for _, shared := range event.Links {
		link, err := ParseFortyOneStoryURL(shared.URL)
		if err != nil || !strings.EqualFold(link.WorkspaceSlug, workspace.Slug) {
			continue
		}
		if _, exists := seen[link.CanonicalURL]; exists {
			continue
		}
		seen[link.CanonicalURL] = struct{}{}
		links = append(links, link)
	}
	if len(links) == 0 {
		return nil
	}
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		authURL, err := p.accountLinkURL(ctx, workspace, event)
		if err != nil {
			return err
		}
		request, err := BuildSlackStoryAuthenticationUnfurlRequest(event.ChannelID, event.MessageTS, authURL)
		if err != nil {
			return err
		}
		return p.workObjects.Unfurl(ctx, botToken, request)
	}

	metadata := SlackWorkObjectMetadata{Entities: make([]SlackWorkObjectEntity, 0, len(links))}
	for _, link := range links {
		story, err := p.storyReader.QueryByRef(ctx, workspace.ID, link.StoryReference)
		if err != nil {
			if errors.Is(err, stories.ErrNotFound) || errors.Is(err, stories.ErrInvalidStoryReference) || slackrepository.IsNotFound(err) {
				continue
			}
			return err
		}
		accessGranted, err := p.slackStoryAccessGranted(ctx, workspace.ID, installation, *linkedUserID, event.ChannelID, story.Team)
		if err != nil {
			return err
		}
		if !accessGranted {
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
	if len(metadata.Entities) == 0 {
		return nil
	}
	return p.workObjects.Unfurl(ctx, botToken, SlackChatUnfurlRequest{
		Channel:  event.ChannelID,
		TS:       event.MessageTS,
		Metadata: &metadata,
	})
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

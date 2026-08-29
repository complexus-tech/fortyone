package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (p *EventProcessor) slackStoryWorkObjectInput(
	ctx context.Context,
	actorID uuid.UUID,
	actorSlackUserID, storyURL string,
	story singleStory,
	editable bool,
) (SlackStoryWorkObjectInput, error) {
	repository, ok := p.repo.(slackWorkObjectRepository)
	if !ok {
		return SlackStoryWorkObjectInput{}, errors.New("slack Work Object repository is not configured")
	}
	return buildSlackStoryWorkObjectInput(ctx, repository, actorID, actorSlackUserID, storyURL, story, editable)
}

func (p *EventProcessor) slackRequestWorkObjectInput(
	ctx context.Context,
	actorID uuid.UUID,
	actorSlackUserID, requestURL string,
	request integrationRequest,
) (SlackRequestWorkObjectInput, error) {
	repository, ok := p.repo.(slackWorkObjectRepository)
	if !ok {
		return SlackRequestWorkObjectInput{}, errors.New("slack Work Object repository is not configured")
	}
	return buildSlackRequestWorkObjectInput(ctx, repository, actorID, actorSlackUserID, requestURL, request)
}

func (p *EventProcessor) slackObjectiveForUser(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	link FortyOneObjectiveLink,
) (objective, bool, error) {
	if p.objectiveReader == nil {
		return objective{}, false, nil
	}
	items, err := p.objectiveReader.ListByID(ctx, workspaceID, userID, link.ObjectiveID)
	if err != nil {
		return objective{}, false, err
	}
	for _, item := range items {
		if item.ID == link.ObjectiveID && item.Workspace == workspaceID && item.Team == link.TeamID {
			return item, true, nil
		}
	}
	return objective{}, false, nil
}

func (p *EventProcessor) slackSprintForUser(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	link FortyOneSprintLink,
) (sprint, bool, error) {
	if p.sprintReader == nil {
		return sprint{}, false, nil
	}
	items, err := p.sprintReader.ListByID(ctx, workspaceID, userID, link.SprintID)
	if err != nil {
		return sprint{}, false, err
	}
	for _, item := range items {
		if item.ID == link.SprintID && item.WorkspaceID == workspaceID && item.TeamID == link.TeamID {
			return item, true, nil
		}
	}
	return sprint{}, false, nil
}

func (p *EventProcessor) slackObjectiveWorkObjectInput(
	ctx context.Context,
	actorID uuid.UUID,
	actorSlackUserID, objectiveURL string,
	objective objective,
) (SlackObjectiveWorkObjectInput, error) {
	repository, ok := p.repo.(slackWorkObjectRepository)
	if !ok {
		return SlackObjectiveWorkObjectInput{}, errors.New("slack Work Object repository is not configured")
	}
	return buildSlackObjectiveWorkObjectInput(ctx, repository, actorID, actorSlackUserID, objectiveURL, objective)
}

func buildSlackStoryWorkObjectInput(
	ctx context.Context,
	repository slackWorkObjectRepository,
	actorID uuid.UUID,
	actorSlackUserID, storyURL string,
	story singleStory,
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
		if err != nil && !isSlackRepositoryNotFound(err) {
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
	request integrationRequest,
) (SlackRequestWorkObjectInput, error) {
	assigneeName, assigneeSlackUserID := "", ""
	if request.AssigneeID != nil && *request.AssigneeID != uuid.Nil {
		member, err := repository.FindTeamMemberByID(ctx, request.TeamID, *request.AssigneeID)
		if err != nil && !isSlackRepositoryNotFound(err) {
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
		if err != nil && !isSlackRepositoryNotFound(err) {
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
	objective objective,
) (SlackObjectiveWorkObjectInput, error) {
	leadName, leadSlackUserID := "", ""
	if objective.LeadUser != nil && *objective.LeadUser != uuid.Nil {
		member, err := repository.FindTeamMemberByID(ctx, objective.Team, *objective.LeadUser)
		if err != nil && !isSlackRepositoryNotFound(err) {
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

func slackSprintWorkObjectInput(sprintURL string, sprint sprint) SlackSprintWorkObjectInput {
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

func slackSprintStatus(sprint sprint, now time.Time) string {
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
	case integrationRequestStatusPending:
		return "Pending"
	case integrationRequestStatusAccepted:
		return "Accepted"
	case integrationRequestStatusDeclined:
		return "Declined"
	default:
		return strings.TrimSpace(status)
	}
}

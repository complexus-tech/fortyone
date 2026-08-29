package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) handleViewSubmission(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if payload.View.CallbackID != "fortyone_create_task" {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}

	submission, err := parseViewSubmission(payload)
	if err != nil {
		s.log.Error(ctx, "failed parsing slack view submission payload", "error", err)
		return interactionValidationErrors(map[string]string{
			"title": interactionErrorMessage(err),
		})
	}

	errorsByBlock := map[string]string{}
	if submission.Title == "" {
		errorsByBlock["title"] = "Title is required"
	}
	if len([]rune(submission.Title)) > modalTitleMaxRunes {
		errorsByBlock[modalBlockTitle] = "Title must be 255 characters or fewer"
	}
	if len([]rune(submission.Description)) > modalDescriptionMaxRunes {
		errorsByBlock[modalBlockDescription] = "Description must be 3000 characters or fewer"
	}
	if len(errorsByBlock) > 0 {
		return interactionValidationErrors(errorsByBlock)
	}
	submission.Source, err = interactionSourceForPayload(payload, submission.Source)
	if err != nil {
		s.log.Warn(ctx, "rejected slack view submission with mismatched actor", "error", err)
		return interactionValidationErrors(map[string]string{"title": "This Slack form no longer belongs to the current user. Open it again and retry."})
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, submission.Source.SlackTeamID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return interactionValidationErrors(map[string]string{"title": "Slack workspace is not connected"})
		}
		s.log.Error(ctx, "failed loading slack workspace by team id", "error", err, "slack_team_id", submission.Source.SlackTeamID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	actorID, err := s.findLinkedInteractionActor(ctx, slackWorkspace.WorkspaceID, submission.Source)
	if err != nil {
		if errors.Is(err, ErrSlackUserNotLinked) {
			return interactionValidationErrors(map[string]string{"title": "Connect your FortyOne account first, then submit again."})
		}
		s.log.Error(ctx, "failed resolving slack actor for view submission", "error", err, "workspace_id", slackWorkspace.WorkspaceID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		s.log.Error(ctx, "failed loading Slack credential for view submission", "error", err, "workspace_id", slackWorkspace.WorkspaceID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}

	workspace, err := s.repo.FindWorkspaceByID(ctx, slackWorkspace.WorkspaceID)
	if err != nil {
		s.log.Error(ctx, "failed loading workspace for slack submission", "error", err, "workspace_id", slackWorkspace.WorkspaceID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	if submission.TeamID == uuid.Nil {
		return interactionValidationErrors(map[string]string{modalBlockTeam: "Team is required"})
	}
	if err := s.ensureTeamAvailableForSlackCreation(ctx, workspace.ID, actorID, submission.TeamID); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return interactionValidationErrors(map[string]string{modalBlockTeam: "Selected team is no longer available to you"})
		}
		s.log.Error(ctx, "failed validating Slack story creation team membership", "error", err, "workspace_id", workspace.ID, "team_id", submission.TeamID, "actor_id", actorID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	team, err := s.findTeamForActor(ctx, workspace.ID, actorID, submission.TeamID)
	if err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return interactionValidationErrors(map[string]string{modalBlockTeam: "Selected team is no longer available to you"})
		}
		s.log.Error(ctx, "failed validating selected team for slack submission", "error", err, "workspace_id", workspace.ID, "team_id", submission.TeamID, "actor_id", actorID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}

	sourceURL := permalinkFromContext(submission.Source)
	sourceExternalID := buildSourceExternalID(submission.Source)
	if viewID := strings.TrimSpace(payload.View.ID); viewID != "" {
		sourceExternalID = fmt.Sprintf("view:%s:%s", viewID, actorID)
	}
	if sourceExternalID == "" {
		sourceExternalID = fmt.Sprintf("slack:%d", s.clock.Now().UnixNano())
	}
	creationKey := fmt.Sprintf("slack:%s:%s", workspace.ID, sourceExternalID)

	description := strings.TrimSpace(submission.Description)
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	metadata := map[string]any{
		"workspace_slug":    workspace.Slug,
		"workspace_name":    workspace.Name,
		"team_code":         team.Code,
		"team_name":         team.Name,
		"slack_team_id":     submission.Source.SlackTeamID,
		"slack_team_domain": submission.Source.SlackTeamDomain,
		"slack_channel_id":  submission.Source.SlackChannelID,
		"slack_channel":     submission.Source.SlackChannel,
		"slack_message_ts":  submission.Source.SlackMessageTS,
		"slack_thread_ts":   submission.Source.SlackThreadTS,
		"slack_user_id":     submission.Source.SlackUserID,
		"slack_username":    submission.Source.SlackUsername,
		"slack_text":        submission.Source.SlackText,
	}

	sendToRequests := submission.StatusKind == slackStatusKindRequest

	if submission.StatusID != nil {
		statuses, statusErr := s.repo.ListTeamStatuses(ctx, team.ID)
		if statusErr != nil {
			s.log.Error(ctx, "failed loading team statuses for slack submission", "error", statusErr, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(statusErr)})
		}
		_, found := findStatusByID(statuses, *submission.StatusID)
		if found {
			sendToRequests = false
		} else {
			return interactionValidationErrors(map[string]string{submission.BlockIDs.Status: "Selected status is no longer available"})
		}
	}
	if submission.StatusKind == slackStatusKindStory && submission.StatusID == nil {
		return interactionValidationErrors(map[string]string{submission.BlockIDs.Status: "Selected status is no longer available"})
	}

	var assigneeID *uuid.UUID
	if submission.AssigneeID != nil {
		members, membersErr := s.repo.ListTeamMembers(ctx, team.ID)
		if membersErr != nil {
			s.log.Error(ctx, "failed loading team members for slack submission", "error", membersErr, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(membersErr)})
		}
		if teamMemberExists(members, *submission.AssigneeID) {
			assigneeID = submission.AssigneeID
		} else {
			return interactionValidationErrors(map[string]string{submission.BlockIDs.Assignee: "Selected assignee is no longer available"})
		}
	}

	var objectiveID *uuid.UUID
	if submission.ObjectiveID != nil {
		if _, objectiveErr := s.repo.FindTeamObjectiveByID(ctx, workspace.ID, team.ID, *submission.ObjectiveID); objectiveErr != nil {
			if !isSlackRepositoryNotFound(objectiveErr) {
				s.log.Error(ctx, "failed validating objective for slack submission", "error", objectiveErr, "workspace_id", workspace.ID, "team_id", team.ID)
				return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(objectiveErr)})
			}
			return interactionValidationErrors(map[string]string{submission.BlockIDs.Objective: "Selected objective is no longer available"})
		} else {
			objectiveID = submission.ObjectiveID
		}
	}

	priority := normalizeSlackPriority(submission.Priority)
	var labelIDs []uuid.UUID
	if len(submission.LabelIDs) > 0 {
		labels, labelsErr := s.repo.ListTeamLabels(ctx, workspace.ID, team.ID)
		if labelsErr != nil {
			s.log.Error(ctx, "failed loading team labels for slack submission", "error", labelsErr, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(labelsErr)})
		}
		labelIDs = filterValidLabelIDs(labels, submission.LabelIDs)
		if len(labelIDs) != len(submission.LabelIDs) {
			return interactionValidationErrors(map[string]string{submission.BlockIDs.Labels: "One or more selected labels are no longer available"})
		}
	}

	creator, err := s.repo.FindTeamMemberByID(ctx, team.ID, actorID)
	if err != nil {
		s.log.Error(ctx, "failed loading Slack submission creator for confirmation", "error", err, "workspace_id", workspace.ID, "team_id", team.ID, "actor_id", actorID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}
	creatorName := storyCreatorDisplayName(creator)

	if sendToRequests {
		metadata["label_ids"] = uuidStrings(labelIDs)
		requestInput := upsertIntegrationRequestInput{
			WorkspaceID:      workspace.ID,
			TeamID:           team.ID,
			Provider:         providerSlack,
			SourceType:       SourceTypeSlackMessage,
			SourceExternalID: sourceExternalID,
			Title:            submission.Title,
			Description:      descriptionPtr,
			Priority:         priority,
			AssigneeID:       assigneeID,
			ObjectiveID:      objectiveID,
			LabelIDs:         labelIDs,
			Metadata:         metadata,
			CreatedByUserID:  &actorID,
		}
		if sourceURL != "" {
			requestInput.SourceURL = &sourceURL
		}
		request, err := s.requests.UpsertPending(ctx, requestInput)
		if err != nil {
			s.log.Error(ctx, "failed creating slack integration request", "error", err, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
		}
		ackKey := fmt.Sprintf("slack:%s:request:%s:confirmation", workspace.ID, sourceExternalID)
		s.postSlackRequestAck(ctx, workspace.ID, slackWorkspace.InstallGeneration, ackKey, submission.Source, botToken, workspace.Slug, request, actorID, creatorName)
		return interactionClearResponse()
	}

	var statusID *uuid.UUID
	if submission.StatusID != nil {
		statusID = submission.StatusID
	} else {
		statusID, err = s.repo.FindFirstStatusByCategory(ctx, team.ID, "unstarted")
		if err != nil {
			s.log.Error(ctx, "failed loading unstarted status for slack task creation", "error", err, "workspace_id", workspace.ID, "team_id", team.ID)
			return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
		}
	}

	story, err := s.stories.Create(ctx, newStory{
		Title:       submission.Title,
		Description: descriptionPtr,
		Status:      statusID,
		Assignee:    assigneeID,
		Objective:   objectiveID,
		Team:        team.ID,
		Reporter:    &actorID,
		Priority:    priority,
		LabelIDs:    labelIDs,
		CreationKey: &creationKey,
	}, workspace.ID)
	if err != nil {
		s.log.Error(ctx, "failed creating story from slack submission", "error", err, "workspace_id", workspace.ID, "team_id", team.ID)
		return interactionValidationErrors(map[string]string{"title": interactionErrorMessage(err)})
	}

	s.postSlackTaskAck(ctx, workspace.ID, slackWorkspace.InstallGeneration, creationKey+":confirmation", submission.Source, botToken, workspace.Slug, team.Code, creatorName, submission.Source.SlackUserID, slackStoryReceiptActionCreated, story)
	return interactionClearResponse()
}

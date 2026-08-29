package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) handleBlockActions(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if isSlackStoryCompactEditAction(payload) {
		return s.handleSlackStoryCompactEditAction(ctx, payload)
	}
	if payload.View.CallbackID != "fortyone_create_task" {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	if len(payload.Actions) == 0 {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}
	firstAction := payload.Actions[0]
	if firstAction.BlockID != modalBlockTeam || firstAction.ActionID != modalActionTeamSelect {
		return InteractionResponse{StatusCode: http.StatusOK}, nil
	}

	submission, err := parseViewSubmission(payload)
	if err != nil {
		s.log.Error(ctx, "failed parsing slack block actions payload", "error", err)
		return InteractionResponse{}, fmt.Errorf("parse Slack block action: %w", err)
	}
	selectedTeamIDRaw := strings.TrimSpace(firstAction.SelectedOption.Value)
	if selectedTeamIDRaw != "" {
		selectedTeamID, parseErr := uuid.Parse(selectedTeamIDRaw)
		if parseErr != nil || selectedTeamID == uuid.Nil {
			s.log.Warn(ctx, "slack team change contained an invalid team id", "team_id", selectedTeamIDRaw)
			return InteractionResponse{}, ErrSlackTeamSelectionRequired
		}
		submission.TeamID = selectedTeamID
	}

	metadata, err := parseSlackModalPrivateMetadata(payload.View.PrivateMetadata)
	if err != nil {
		s.log.Error(ctx, "failed parsing slack modal metadata for team change", "error", err)
		return InteractionResponse{}, fmt.Errorf("parse Slack modal metadata: %w", err)
	}
	submission.Source, err = interactionSourceForPayload(payload, submission.Source)
	if err != nil {
		s.log.Warn(ctx, "rejected slack team change with mismatched actor", "error", err)
		return InteractionResponse{}, ErrSlackInteractionActorMismatch
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, submission.Source.SlackTeamID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return InteractionResponse{}, ErrSlackNoWorkspaceLinked
		}
		return InteractionResponse{}, err
	}
	actorID, err := s.findLinkedInteractionActor(ctx, slackWorkspace.WorkspaceID, submission.Source)
	if err != nil {
		if errors.Is(err, ErrSlackUserNotLinked) {
			s.log.Warn(ctx, "rejected slack team change from unlinked user", "slack_user_id", submission.Source.SlackUserID)
			return InteractionResponse{}, ErrSlackUserNotLinked
		}
		return InteractionResponse{}, err
	}
	if err := s.ensureTeamAvailableForSlackCreation(ctx, slackWorkspace.WorkspaceID, actorID, submission.TeamID); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			s.log.Warn(ctx, "rejected slack team change outside actor membership", "actor_id", actorID, "team_id", submission.TeamID)
			return InteractionResponse{}, ErrSlackTeamNotAvailable
		}
		return InteractionResponse{}, err
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return InteractionResponse{}, err
	}

	selection := createTaskModalSelection{
		StatusKind:  submission.StatusKind,
		TeamID:      submission.TeamID,
		StatusID:    submission.StatusID,
		Priority:    submission.Priority,
		AssigneeID:  submission.AssigneeID,
		LabelIDs:    submission.LabelIDs,
		ObjectiveID: submission.ObjectiveID,
	}
	previousTeamID, _ := uuid.Parse(strings.TrimSpace(metadata.SelectedTeamID))
	if previousTeamID != uuid.Nil && previousTeamID != submission.TeamID {
		selection = createTaskModalSelection{
			TeamID:   submission.TeamID,
			Priority: submission.Priority,
		}
	}

	view, err := s.buildCreateTaskModalView(ctx, createTaskModalViewInput{
		Title:       submission.Title,
		Description: submission.Description,
		Source:      submission.Source,
		WorkspaceID: slackWorkspace.WorkspaceID,
		ActorID:     actorID,
		Selection:   selection,
	})
	if err != nil {
		return InteractionResponse{}, err
	}

	updatePayload := map[string]any{
		"view_id": payload.View.ID,
		"hash":    payload.View.Hash,
		"view":    view,
	}
	if err := s.callSlackAPI(ctx, botToken, "https://slack.com/api/views.update", updatePayload, nil); err != nil {
		return InteractionResponse{}, err
	}

	return InteractionResponse{StatusCode: http.StatusOK}, nil
}

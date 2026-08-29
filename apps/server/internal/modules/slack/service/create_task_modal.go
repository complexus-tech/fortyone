package slack

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) openCreateTaskModal(ctx context.Context, triggerID, title, description string, source requestSourceContext, workspaceID, actorID uuid.UUID, botToken string) error {
	if strings.TrimSpace(botToken) == "" {
		return errors.New("missing slack bot token")
	}
	if strings.TrimSpace(triggerID) == "" {
		return errors.New("missing trigger id")
	}
	if workspaceID == uuid.Nil {
		return errors.New("missing workspace id")
	}
	if actorID == uuid.Nil {
		return errors.New("missing actor id")
	}

	var opened struct {
		View struct {
			ID string `json:"id"`
		} `json:"view"`
	}
	if err := s.callSlackAPI(ctx, botToken, "https://slack.com/api/views.open", map[string]any{
		"trigger_id": triggerID,
		"view":       createTaskLoadingModalView(),
	}, &opened); err != nil {
		return err
	}
	viewID := strings.TrimSpace(opened.View.ID)
	if viewID == "" {
		return errors.New("slack did not return an opened modal view ID")
	}

	// The trigger is now safely consumed. Team/status/member lookups can finish
	// under a separate bounded context without risking Slack's three-second
	// trigger expiry and generic shortcut error.
	hydrationCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), slackModalHydrationTimeout)
	defer cancel()
	view, err := s.buildCreateTaskModalView(hydrationCtx, createTaskModalViewInput{
		Title:       title,
		Description: description,
		Source:      source,
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		s.replaceCreateTaskModalWithError(hydrationCtx, botToken, viewID)
		return err
	}

	payload := map[string]any{
		"view_id": viewID,
		"view":    view,
	}
	if err := s.callSlackAPI(hydrationCtx, botToken, "https://slack.com/api/views.update", payload, nil); err != nil {
		s.replaceCreateTaskModalWithError(hydrationCtx, botToken, viewID)
		return err
	}
	return nil
}

func createTaskLoadingModalView() map[string]any {
	return map[string]any{
		"type":        "modal",
		"callback_id": "fortyone_create_task",
		"title": map[string]string{
			"type": "plain_text",
			"text": "Create Story",
		},
		"close": map[string]string{
			"type": "plain_text",
			"text": "Cancel",
		},
		"blocks": []map[string]any{{
			"type": "section",
			"text": map[string]string{
				"type": "plain_text",
				"text": "Loading story form…",
			},
		}},
	}
}

func createTaskErrorModalView() map[string]any {
	return map[string]any{
		"type":        "modal",
		"callback_id": "fortyone_create_task",
		"title": map[string]string{
			"type": "plain_text",
			"text": "Create Story",
		},
		"close": map[string]string{
			"type": "plain_text",
			"text": "Close",
		},
		"blocks": []map[string]any{{
			"type": "section",
			"text": map[string]string{
				"type": "plain_text",
				"text": "FortyOne couldn't load this form. Close it and try again.",
			},
		}},
	}
}

func (s *Service) replaceCreateTaskModalWithError(ctx context.Context, botToken, viewID string) {
	if err := s.callSlackAPI(ctx, botToken, "https://slack.com/api/views.update", map[string]any{
		"view_id": strings.TrimSpace(viewID),
		"view":    createTaskErrorModalView(),
	}, nil); err != nil && s.log != nil {
		s.log.Warn(ctx, "failed replacing Slack create story modal with error state", "error", err, "view_id", viewID)
	}
}

type createTaskModalViewInput struct {
	Title       string
	Description string
	Source      requestSourceContext
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	Selection   createTaskModalSelection
}

type createTaskModalSelection struct {
	StatusKind  string
	TeamID      uuid.UUID
	StatusID    *uuid.UUID
	Priority    string
	AssigneeID  *uuid.UUID
	LabelIDs    []uuid.UUID
	ObjectiveID *uuid.UUID
}

func (s *Service) buildCreateTaskModalView(ctx context.Context, input createTaskModalViewInput) (map[string]any, error) {
	if input.ActorID == uuid.Nil {
		return nil, errors.New("actor id is required")
	}
	teams, err := s.availableTeamsForSlackCreation(ctx, input.WorkspaceID, input.ActorID)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, ErrSlackNoTeamsAvailable
	}

	selectedTeam := selectTeam(teams, input.Selection.TeamID)
	selectedTeamOption := slackTeamOption(selectedTeam)

	statuses, err := s.repo.ListTeamStatuses(ctx, selectedTeam.ID)
	if err != nil {
		return nil, err
	}

	useExternalStatusSelect := len(statuses)+1 > slackSelectMaxOptions
	statusOptions := make([]map[string]any, 0, min(len(statuses)+1, slackSelectMaxOptions))
	requestStatusOption := toSlackOption("Request", slackRequestStatusValue)
	if !useExternalStatusSelect {
		statusOptions = append(statusOptions, requestStatusOption)
	}
	selectedStatusOption := requestStatusOption
	for index, status := range statuses {
		option := toSlackOption(status.Name, status.ID.String())
		if !useExternalStatusSelect {
			statusOptions = append(statusOptions, option)
		}
		if index == 0 && input.Selection.StatusKind == "" && input.Selection.StatusID == nil {
			selectedStatusOption = option
		}
		if input.Selection.StatusKind == slackStatusKindStory && input.Selection.StatusID != nil && *input.Selection.StatusID == status.ID {
			selectedStatusOption = option
		}
	}

	var selectedAssigneeOption map[string]any
	if input.Selection.AssigneeID != nil && *input.Selection.AssigneeID != uuid.Nil {
		member, memberErr := s.repo.FindTeamMemberByID(ctx, selectedTeam.ID, *input.Selection.AssigneeID)
		if memberErr == nil {
			selectedAssigneeOption = toSlackOption(teamMemberDisplayName(member), member.UserID.String())
		}
	}
	selectedLabelOptions := make([]map[string]any, 0, len(input.Selection.LabelIDs))
	for _, labelID := range input.Selection.LabelIDs {
		label, labelErr := s.repo.FindTeamLabelByID(ctx, input.WorkspaceID, selectedTeam.ID, labelID)
		if labelErr != nil {
			continue
		}
		selectedLabelOptions = append(selectedLabelOptions, toSlackOption(label.Name, label.ID.String()))
	}

	var selectedObjectiveOption map[string]any
	if input.Selection.ObjectiveID != nil && *input.Selection.ObjectiveID != uuid.Nil {
		objective, objectiveErr := s.repo.FindTeamObjectiveByID(ctx, input.WorkspaceID, selectedTeam.ID, *input.Selection.ObjectiveID)
		if objectiveErr == nil {
			selectedObjectiveOption = toSlackOption(objective.Name, objective.ID.String())
		}
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "New task"
	}
	metadataSource := input.Source
	metadataSource.SlackText = truncateRunes(metadataSource.SlackText, modalSourceTextMaxRunes)
	metadataPayload, err := json.Marshal(slackModalPrivateMetadata{
		Source:         metadataSource,
		SelectedTeamID: selectedTeam.ID.String(),
	})
	if err != nil {
		return nil, err
	}
	if len(metadataPayload) > modalMetadataMaxBytes {
		metadataSource.SlackText = ""
		metadataPayload, err = json.Marshal(slackModalPrivateMetadata{
			Source:         metadataSource,
			SelectedTeamID: selectedTeam.ID.String(),
		})
		if err != nil {
			return nil, err
		}
	}
	if len(metadataPayload) > modalMetadataMaxBytes {
		return nil, errors.New("slack modal metadata exceeds the provider limit")
	}

	priorityOption := toSlackOption(normalizeSlackPriority(input.Selection.Priority), normalizeSlackPriority(input.Selection.Priority))
	statusBlockID := modalTeamScopedID(modalBlockStatus, selectedTeam.ID)
	statusActionID := modalTeamScopedID(modalActionStatusSelect, selectedTeam.ID)
	assigneeBlockID := modalTeamScopedID(modalBlockAssignee, selectedTeam.ID)
	assigneeActionID := modalTeamScopedID(modalActionAssigneeSelect, selectedTeam.ID)
	labelsBlockID := modalTeamScopedID(modalBlockLabels, selectedTeam.ID)
	labelsActionID := modalTeamScopedID(modalActionLabelsMultiSelect, selectedTeam.ID)
	objectiveBlockID := modalTeamScopedID(modalBlockObjective, selectedTeam.ID)
	objectiveActionID := modalTeamScopedID(modalActionObjectiveSelect, selectedTeam.ID)
	statusBlock := selectInputBlock(statusBlockID, statusActionID, "Status", statusOptions, selectedStatusOption, true, false)
	if useExternalStatusSelect {
		statusBlock = externalSelectInputBlock(statusBlockID, statusActionID, "Status", selectedStatusOption, true, slackExternalSearchMinRunes, false)
	}

	blocks := []map[string]any{
		externalSelectInputBlock(modalBlockTeam, modalActionTeamSelect, "Team", selectedTeamOption, false, slackExternalSearchMinRunes, true),
		plainInputBlock(modalBlockTitle, modalActionTitleInput, "Title", truncateRunes(title, modalTitleMaxRunes), false, "", false),
		plainInputBlock(modalBlockDescription, modalActionDescriptionInput, "Description", truncateRunes(input.Description, modalDescriptionMaxRunes), true, "", true),
		statusBlock,
		externalSelectInputBlock(assigneeBlockID, assigneeActionID, "Assignee", selectedAssigneeOption, true, 2, false),
		externalMultiSelectInputBlock(labelsBlockID, labelsActionID, "Labels", selectedLabelOptions, true, 2),
		externalSelectInputBlock(objectiveBlockID, objectiveActionID, "Objective", selectedObjectiveOption, true, 2, false),
	}
	blocks = append(blocks, selectInputBlock(modalBlockPriority, modalActionPrioritySelect, "Priority", slackPriorityOptions(), priorityOption, true, false))

	return map[string]any{
		"type":             "modal",
		"callback_id":      "fortyone_create_task",
		"private_metadata": string(metadataPayload),
		"title": map[string]string{
			"type": "plain_text",
			"text": "Create Story",
		},
		"submit": map[string]string{
			"type": "plain_text",
			"text": "Create",
		},
		"close": map[string]string{
			"type": "plain_text",
			"text": "Cancel",
		},
		"blocks": blocks,
	}, nil
}

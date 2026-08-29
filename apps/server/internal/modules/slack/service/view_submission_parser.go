package slack

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

func parseViewSubmission(payload interactionPayload) (viewSubmissionData, error) {
	state := payload.View.State.Values
	readValue := func(blockID string) string {
		block := state[blockID]
		for _, action := range block {
			if strings.TrimSpace(action.Value) != "" {
				return strings.TrimSpace(action.Value)
			}
		}
		return ""
	}
	readSelectedOption := func(blockID string) string {
		block := state[blockID]
		for _, action := range block {
			if strings.TrimSpace(action.SelectedOption.Value) != "" {
				return strings.TrimSpace(action.SelectedOption.Value)
			}
		}
		return ""
	}

	metadata, err := parseSlackModalPrivateMetadata(payload.View.PrivateMetadata)
	if err != nil {
		return viewSubmissionData{}, err
	}
	source := metadata.Source

	selectedTeamID := readSelectedOption(modalBlockTeam)
	if selectedTeamID == "" {
		selectedTeamID = strings.TrimSpace(metadata.SelectedTeamID)
	}
	if selectedTeamID == "" {
		return viewSubmissionData{}, ErrSlackTeamSelectionRequired
	}
	teamID, err := uuid.Parse(selectedTeamID)
	if err != nil {
		return viewSubmissionData{}, errors.New("invalid selected team")
	}

	var statusID *uuid.UUID
	statusKind := slackStatusKindRequest
	statusState, statusBlockID, _ := modalDependentStateValue(
		state,
		modalBlockStatus,
		modalActionStatusSelect,
		teamID,
	)
	selectedStatusID := strings.TrimSpace(statusState.SelectedOption.Value)
	if selectedStatusID == slackRequestStatusValue || selectedStatusID == "" {
		statusKind = slackStatusKindRequest
	} else {
		parsedStatusID, parseErr := uuid.Parse(selectedStatusID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected status")
		}
		statusKind = slackStatusKindStory
		statusID = &parsedStatusID
	}

	var assigneeID *uuid.UUID
	assigneeState, assigneeBlockID, _ := modalDependentStateValue(
		state,
		modalBlockAssignee,
		modalActionAssigneeSelect,
		teamID,
	)
	selectedAssigneeID := strings.TrimSpace(assigneeState.SelectedOption.Value)
	if selectedAssigneeID != "" {
		parsedAssigneeID, parseErr := uuid.Parse(selectedAssigneeID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected assignee")
		}
		assigneeID = &parsedAssigneeID
	}

	labelsState, labelsBlockID, _ := modalDependentStateValue(
		state,
		modalBlockLabels,
		modalActionLabelsMultiSelect,
		teamID,
	)
	selectedLabelIDs := make([]uuid.UUID, 0)
	for _, selected := range labelsState.SelectedOptions {
		selectedLabelID := strings.TrimSpace(selected.Value)
		if selectedLabelID == "" {
			continue
		}
		parsedLabelID, parseErr := uuid.Parse(selectedLabelID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected label")
		}
		selectedLabelIDs = append(selectedLabelIDs, parsedLabelID)
	}

	var objectiveID *uuid.UUID
	objectiveState, objectiveBlockID, _ := modalDependentStateValue(
		state,
		modalBlockObjective,
		modalActionObjectiveSelect,
		teamID,
	)
	selectedObjectiveID := strings.TrimSpace(objectiveState.SelectedOption.Value)
	if selectedObjectiveID != "" {
		parsedObjectiveID, parseErr := uuid.Parse(selectedObjectiveID)
		if parseErr != nil {
			return viewSubmissionData{}, errors.New("invalid selected objective")
		}
		objectiveID = &parsedObjectiveID
	}

	return viewSubmissionData{
		Title:       readValue(modalBlockTitle),
		Description: readValue(modalBlockDescription),
		TeamID:      teamID,
		StatusKind:  statusKind,
		StatusID:    statusID,
		Priority:    readSelectedOption(modalBlockPriority),
		AssigneeID:  assigneeID,
		LabelIDs:    selectedLabelIDs,
		ObjectiveID: objectiveID,
		Source:      source,
		BlockIDs: modalDependentBlockIDs{
			Status:    statusBlockID,
			Assignee:  assigneeBlockID,
			Labels:    labelsBlockID,
			Objective: objectiveBlockID,
		},
	}, nil
}

func parseSourceFromPrivateMetadata(privateMetadata string) (requestSourceContext, error) {
	metadata, err := parseSlackModalPrivateMetadata(privateMetadata)
	if err != nil {
		return requestSourceContext{}, err
	}
	return metadata.Source, nil
}

func selectedTeamIDFromState(values interactionViewStateValues) string {
	block := values[modalBlockTeam]
	for _, action := range block {
		value := strings.TrimSpace(action.SelectedOption.Value)
		if value != "" {
			return value
		}
	}
	return ""
}

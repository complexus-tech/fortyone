package slack

import (
	"strings"

	"github.com/google/uuid"
)

const (
	modalElementIDMaxBytes     = 255
	modalTeamScopedIDSeparator = "__team_"
)

type interactionViewStateValue struct {
	Type           string `json:"type"`
	Value          string `json:"value"`
	SelectedDate   string `json:"selected_date"`
	SelectedOption struct {
		Value string `json:"value"`
	} `json:"selected_option"`
	SelectedOptions []struct {
		Value string `json:"value"`
	} `json:"selected_options"`
}

type interactionViewStateValues map[string]map[string]interactionViewStateValue

// modalTeamScopedID gives team-dependent inputs a new identity whenever the
// selected team changes. Slack otherwise preserves state across views.update
// calls when both block_id and action_id remain unchanged.
func modalTeamScopedID(base string, teamID uuid.UUID) string {
	base = strings.TrimSpace(base)
	if base == "" || teamID == uuid.Nil {
		return base
	}
	suffix := modalTeamScopedIDSeparator + strings.ReplaceAll(teamID.String(), "-", "")
	if len(base)+len(suffix) > modalElementIDMaxBytes {
		base = base[:modalElementIDMaxBytes-len(suffix)]
	}
	return base + suffix
}

func modalDependentStateValue(
	state interactionViewStateValues,
	blockBase string,
	actionBase string,
	teamID uuid.UUID,
) (interactionViewStateValue, string, bool) {
	expectedBlockID := modalTeamScopedID(blockBase, teamID)
	if block, ok := state[expectedBlockID]; ok {
		expectedActionID := modalTeamScopedID(actionBase, teamID)
		value, found := block[expectedActionID]
		return value, expectedBlockID, found
	}

	// Accept pre-versioning modals that were already open during a deploy. New
	// modals never use these IDs, so this compatibility path cannot preserve a
	// stale selection through a team-scoped view update.
	if block, ok := state[blockBase]; ok {
		if value, found := block[actionBase]; found {
			return value, blockBase, true
		}
		for _, value := range block {
			return value, blockBase, true
		}
		return interactionViewStateValue{}, blockBase, false
	}

	return interactionViewStateValue{}, expectedBlockID, false
}

func modalDependentActionBase(actionID string, teamID uuid.UUID) (string, bool) {
	actionID = strings.TrimSpace(actionID)
	for _, base := range []string{
		modalActionStatusSelect,
		modalActionAssigneeSelect,
		modalActionLabelsMultiSelect,
		modalActionObjectiveSelect,
	} {
		if actionID == base || actionID == modalTeamScopedID(base, teamID) {
			return base, true
		}
	}
	return "", false
}

func modalDependentBlockMatches(blockID, blockBase string, teamID uuid.UUID) bool {
	blockID = strings.TrimSpace(blockID)
	return blockID == "" || blockID == blockBase || blockID == modalTeamScopedID(blockBase, teamID)
}

func modalDependentBlockForAction(actionBase string) string {
	switch actionBase {
	case modalActionStatusSelect:
		return modalBlockStatus
	case modalActionAssigneeSelect:
		return modalBlockAssignee
	case modalActionLabelsMultiSelect:
		return modalBlockLabels
	case modalActionObjectiveSelect:
		return modalBlockObjective
	default:
		return ""
	}
}

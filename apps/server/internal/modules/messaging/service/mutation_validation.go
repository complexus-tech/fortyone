package messaging

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func validateStoryTimeMutationClaims(mutation storyTimeMutation) error {
	estimatedDurationMinutes, err := resolvedStoryTimeValue(
		nil,
		mutation.estimatedDurationAction,
		mutation.estimatedDurationMinutes,
		"estimated_duration",
	)
	if err != nil {
		return err
	}
	minimumFocusBlockMinutes, err := resolvedStoryTimeValue(
		nil,
		mutation.minimumFocusBlockAction,
		mutation.minimumFocusBlockMinutes,
		"minimum_focus_block",
	)
	if err != nil {
		return err
	}

	switch {
	case mutation.estimatedDurationAction == storyTimeActionClear && mutation.minimumFocusBlockAction == storyTimeActionSet:
		return storyFocusBlockRequiresDurationError()
	case mutation.estimatedDurationAction == storyTimeActionSet && mutation.minimumFocusBlockAction == storyTimeActionSet:
		return validateStoryTimeContract(estimatedDurationMinutes, minimumFocusBlockMinutes)
	default:
		return nil
	}
}

func validateStoryMutationClaims(claims storyMutationClaims) error {
	if claims.Title != nil {
		title, err := normalizedStoryTitle(*claims.Title)
		if err != nil || title != *claims.Title {
			return fmt.Errorf("%w: invalid title claim", ErrInvalidConfirmation)
		}
	}
	if claims.Priority != nil {
		priority, err := normalizedStoryPriority(claims.Priority, "")
		if err != nil || priority != *claims.Priority {
			return fmt.Errorf("%w: invalid priority claim", ErrInvalidConfirmation)
		}
	}

	switch claims.Operation {
	case StoryMutationCreate:
		if claims.Title == nil || claims.Priority == nil || claims.StoryID != nil || claims.ExpectedUpdatedAt != nil {
			return fmt.Errorf("%w: malformed create claims", ErrInvalidConfirmation)
		}
		if claims.AutoSchedulingLocked != nil {
			return fmt.Errorf("%w: create claims cannot set auto-scheduling lock state", ErrInvalidConfirmation)
		}
		if claims.AssigneeAction != assigneeActionMe && claims.AssigneeAction != assigneeActionUnassigned {
			return fmt.Errorf("%w: invalid create assignee claim", ErrInvalidConfirmation)
		}
		if err := validateStoryTimeContract(claims.EstimatedDurationMinutes, claims.MinimumFocusBlockMinutes); err != nil {
			return fmt.Errorf("%w: invalid create time claim: %v", ErrInvalidConfirmation, err)
		}
		if err := validateStoryAutoSchedulingContract(optionalBoolValue(claims.AutoSchedulingEnabled), false, storyAutoSchedulingStatusOff); err != nil {
			return fmt.Errorf("%w: invalid create auto-scheduling claim: %v", ErrInvalidConfirmation, err)
		}
	case StoryMutationUpdate:
		if claims.StoryID == nil || *claims.StoryID == uuid.Nil || claims.ExpectedUpdatedAt == nil || claims.ExpectedUpdatedAt.IsZero() {
			return fmt.Errorf("%w: malformed update claims", ErrInvalidConfirmation)
		}
		if claims.AssigneeAction != assigneeActionUnchanged && claims.AssigneeAction != assigneeActionMe && claims.AssigneeAction != assigneeActionUnassigned {
			return fmt.Errorf("%w: invalid update assignee claim", ErrInvalidConfirmation)
		}
		if err := validateStoryTimeMutationClaims(storyTimeMutation{
			estimatedDurationAction:  claims.EstimatedDurationAction,
			estimatedDurationMinutes: claims.EstimatedDurationMinutes,
			minimumFocusBlockAction:  claims.MinimumFocusBlockAction,
			minimumFocusBlockMinutes: claims.MinimumFocusBlockMinutes,
		}); err != nil {
			return fmt.Errorf("%w: invalid update time claim: %v", ErrInvalidConfirmation, err)
		}
		if claims.Title == nil && claims.Priority == nil && claims.AssigneeAction == assigneeActionUnchanged && claims.StatusID == nil && claims.SprintID == nil && claims.ObjectiveID == nil && claims.KeyResultID == nil && claims.StartDate == nil && claims.EndDate == nil && claims.LabelIDs == nil && claims.EstimatedDurationAction == storyTimeActionUnchanged && claims.MinimumFocusBlockAction == storyTimeActionUnchanged && claims.AutoSchedulingEnabled == nil && claims.AutoSchedulingLocked == nil {
			return fmt.Errorf("%w: empty update claims", ErrInvalidConfirmation)
		}
	case StoryMutationComment:
		if claims.StoryID == nil || claims.Comment == nil || strings.TrimSpace(*claims.Comment) == "" {
			return fmt.Errorf("%w: malformed comment claims", ErrInvalidConfirmation)
		}
	case StoryMutationRelation:
		if claims.StoryID == nil || claims.RelationStoryID == nil || claims.RelationType == "" || *claims.StoryID == *claims.RelationStoryID {
			return fmt.Errorf("%w: malformed relationship claims", ErrInvalidConfirmation)
		}
		if claims.RelationType != "blocking" && claims.RelationType != "related" && claims.RelationType != "duplicate" {
			return fmt.Errorf("%w: invalid relationship type", ErrInvalidConfirmation)
		}
	default:
		return fmt.Errorf("%w: unsupported operation", ErrInvalidConfirmation)
	}
	return nil
}

func sameUUIDSet(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(left))
	for _, value := range left {
		seen[value] = struct{}{}
	}
	for _, value := range right {
		if _, ok := seen[value]; !ok {
			return false
		}
	}
	return true
}

func mutationConfirmationFromToolResult(raw json.RawMessage) (*StoryMutationConfirmation, bool, error) {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &header); err != nil || header.Kind != storyMutationConfirmationKind {
		return nil, false, nil
	}
	var result storyMutationConfirmationToolResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, true, fmt.Errorf("%w: decode mutation confirmation: %v", ErrMalformedResponse, err)
	}
	confirmation := result.Confirmation
	if confirmation.Token == "" || confirmation.Prompt == "" || confirmation.ExpiresAt.IsZero() {
		return nil, true, fmt.Errorf("%w: incomplete mutation confirmation", ErrMalformedResponse)
	}
	if confirmation.Operation == StoryMutationCreateBatch {
		if len(confirmation.Stories) == 0 || len(confirmation.Stories) > maximumBatchStoryCount {
			return nil, true, fmt.Errorf("%w: incomplete batch mutation confirmation", ErrMalformedResponse)
		}
		teamID := confirmation.Stories[0].TeamID
		if teamID == uuid.Nil {
			return nil, true, fmt.Errorf("%w: incomplete batch mutation confirmation", ErrMalformedResponse)
		}
		for _, story := range confirmation.Stories[1:] {
			if story.TeamID != teamID {
				return nil, true, fmt.Errorf("%w: batch mutation spans multiple teams", ErrMalformedResponse)
			}
		}
	} else if confirmation.Story.TeamID == uuid.Nil {
		return nil, true, fmt.Errorf("%w: incomplete mutation confirmation", ErrMalformedResponse)
	}
	return &confirmation, true, nil
}

func validateToolScope(scope *ToolScope) error {
	if scope.WorkspaceID == uuid.Nil || scope.UserID == uuid.Nil {
		return fmt.Errorf("%w: workspace and user are required", ErrInvalidRequest)
	}
	allowedTeamIDs, err := normalizedAllowedTeamIDs(scope.AllowedTeamIDs)
	if err != nil {
		return err
	}
	scope.AllowedTeamIDs = allowedTeamIDs
	scope.SourceURL, err = normalizedSourceURL(scope.SourceURL)
	if err != nil {
		return err
	}
	return nil
}

func parseAccessibleTeamID(raw string, joined map[uuid.UUID]messagingTeam) (uuid.UUID, error) {
	teamID, err := parseRequiredUUID(raw, "team_id")
	if err != nil {
		return uuid.Nil, err
	}
	if _, ok := joined[teamID]; !ok {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, teamID)
	}
	return teamID, nil
}

func parseRequiredUUID(raw, field string) (uuid.UUID, error) {
	value, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || value == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: %s must be a UUID", ErrInvalidToolArguments, field)
	}
	return value, nil
}

func normalizedStoryTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if title == "" || len([]rune(title)) > maximumStoryTitleRunes {
		return "", fmt.Errorf("%w: title must contain 1-%d characters", ErrInvalidToolArguments, maximumStoryTitleRunes)
	}
	return title, nil
}

func normalizedStoryReference(raw string) (string, string, error) {
	reference := strings.ToUpper(strings.TrimSpace(raw))
	separator := strings.LastIndex(reference, "-")
	if separator < 1 || separator == len(reference)-1 {
		return "", "", fmt.Errorf("%w: story_reference must use TEAM-123 format", ErrInvalidToolArguments)
	}
	teamCode := strings.TrimSpace(reference[:separator])
	sequenceRaw := strings.TrimSpace(reference[separator+1:])
	sequenceID, err := strconv.Atoi(sequenceRaw)
	if teamCode == "" || err != nil || sequenceID < 1 {
		return "", "", fmt.Errorf("%w: story_reference must use TEAM-123 format", ErrInvalidToolArguments)
	}
	return fmt.Sprintf("%s-%d", teamCode, sequenceID), teamCode, nil
}

func normalizedStoryPriority(raw *string, fallback string) (string, error) {
	if raw == nil {
		return fallback, nil
	}
	priority := strings.TrimSpace(*raw)
	if _, ok := storyPriorities[priority]; !ok {
		return "", fmt.Errorf("%w: unsupported priority %q", ErrInvalidToolArguments, priority)
	}
	return priority, nil
}

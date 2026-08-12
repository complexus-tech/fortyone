package emailagent

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxCheckInRunes = 4_000
	dateOnlyLayout  = "2006-01-02"
)

func resolveActionProposal(request Request, validated validatedRequest, model *ModelActionProposal) (ActionProposal, error) {
	if model == nil {
		return ActionProposal{}, fmt.Errorf("%w: propose intent requires a proposal", ErrInvalidDecision)
	}
	model.Summary = strings.TrimSpace(model.Summary)
	if model.Summary == "" || utf8.RuneCountInString(model.Summary) > maxProposalSummaryRunes {
		return ActionProposal{}, fmt.Errorf("%w: proposal has an invalid summary", ErrInvalidDecision)
	}
	if containsModelAuthoredURL(model.Summary) {
		return ActionProposal{}, fmt.Errorf("%w: proposal summary contains a model-authored URL", ErrInvalidDecision)
	}
	if payloadCount(model) != 1 {
		return ActionProposal{}, fmt.Errorf("%w: proposal must contain exactly one action payload", ErrInvalidDecision)
	}

	proposal := ActionProposal{
		WorkspaceID: request.WorkspaceID,
		ActorID:     request.ActorID,
		Kind:        model.Kind,
		Summary:     model.Summary,
	}
	var err error
	switch model.Kind {
	case ActionObjectiveUpdate:
		if model.Objective == nil {
			return ActionProposal{}, mismatchedPayload(model.Kind)
		}
		proposal.Objective, err = resolveObjectiveAction(validated, *model.Objective)
	case ActionKeyResultUpdate:
		if model.KeyResult == nil {
			return ActionProposal{}, mismatchedPayload(model.Kind)
		}
		proposal.KeyResult, err = resolveKeyResultAction(validated, *model.KeyResult)
	case ActionStoryUpdate:
		if model.Story == nil {
			return ActionProposal{}, mismatchedPayload(model.Kind)
		}
		proposal.Story, err = resolveStoryAction(validated, *model.Story)
	case ActionFeedbackStatus:
		if model.Feedback == nil {
			return ActionProposal{}, mismatchedPayload(model.Kind)
		}
		proposal.Feedback, err = resolveFeedbackStatusAction(validated, *model.Feedback)
	default:
		return ActionProposal{}, fmt.Errorf("%w: unsupported proposal kind %q", ErrInvalidDecision, model.Kind)
	}
	if err != nil {
		return ActionProposal{}, err
	}
	proposal.Summary = trustedProposalSummary(proposal)
	if utf8.RuneCountInString(proposal.Summary) > maxProposalSummaryRunes {
		return ActionProposal{}, fmt.Errorf("%w: trusted proposal summary exceeds %d runes", ErrInvalidDecision, maxProposalSummaryRunes)
	}
	return proposal, nil
}

func validateProposalCopy(copy EmailCopy, proposal ActionProposal) error {
	plainText := copy.PlainText
	required := []string{"CONFIRM", "CANCEL"}
	switch proposal.Kind {
	case ActionObjectiveUpdate:
		required = append(required,
			proposal.Objective.Target.DisplayName,
			string(*proposal.Objective.Health),
		)
		if proposal.Objective.CheckIn != nil {
			required = append(required, *proposal.Objective.CheckIn)
		}
	case ActionKeyResultUpdate:
		required = append(required,
			proposal.KeyResult.Target.DisplayName,
			strconv.FormatFloat(*proposal.KeyResult.CurrentValue, 'f', -1, 64),
		)
		if proposal.KeyResult.CheckIn != nil {
			required = append(required, *proposal.KeyResult.CheckIn)
		}
	case ActionStoryUpdate:
		required = append(required, proposal.Story.Target.DisplayName)
		if proposal.Story.DueDate != nil && proposal.Story.DueDate.Operation == DateSet {
			required = append(required, proposal.Story.DueDate.Date)
		}
		if proposal.Story.Status != nil {
			required = append(required, proposal.Story.Status.StatusName)
		}
		if proposal.Story.Assignee != nil && proposal.Story.Assignee.Operation == AssigneeAssign {
			required = append(required, proposal.Story.Assignee.AssigneeName)
		}
	case ActionFeedbackStatus:
		required = append(required,
			proposal.Feedback.Target.DisplayName,
			string(proposal.Feedback.Status),
		)
	}
	for _, value := range required {
		if !strings.Contains(plainText, value) {
			return fmt.Errorf("%w: proposal copy omits exact value %q", ErrInvalidDecision, value)
		}
	}
	if proposal.Story != nil && proposal.Story.DueDate != nil && proposal.Story.DueDate.Operation == DateClear {
		lower := strings.ToLower(plainText)
		if !strings.Contains(lower, "clear") || !strings.Contains(lower, "due date") {
			return fmt.Errorf("%w: proposal copy does not clearly state that the due date will be cleared", ErrInvalidDecision)
		}
	}
	if proposal.Story != nil && proposal.Story.Assignee != nil && proposal.Story.Assignee.Operation == AssigneeUnassign {
		if !strings.Contains(strings.ToLower(plainText), "unassign") {
			return fmt.Errorf("%w: proposal copy does not clearly state that the story will be unassigned", ErrInvalidDecision)
		}
	}
	return nil
}

func trustedProposalSummary(proposal ActionProposal) string {
	switch proposal.Kind {
	case ActionObjectiveUpdate:
		return fmt.Sprintf("Set %q health to %s", proposal.Objective.Target.DisplayName, *proposal.Objective.Health)
	case ActionKeyResultUpdate:
		return fmt.Sprintf(
			"Set %q current value to %s",
			proposal.KeyResult.Target.DisplayName,
			strconv.FormatFloat(*proposal.KeyResult.CurrentValue, 'f', -1, 64),
		)
	case ActionStoryUpdate:
		changes := make([]string, 0, 3)
		if proposal.Story.DueDate != nil {
			if proposal.Story.DueDate.Operation == DateClear {
				changes = append(changes, "clear the due date")
			} else {
				changes = append(changes, "set the due date to "+proposal.Story.DueDate.Date)
			}
		}
		if proposal.Story.Status != nil {
			changes = append(changes, "set the status to "+proposal.Story.Status.StatusName)
		}
		if proposal.Story.Assignee != nil {
			if proposal.Story.Assignee.Operation == AssigneeUnassign {
				changes = append(changes, "unassign the story")
			} else {
				changes = append(changes, "assign it to "+proposal.Story.Assignee.AssigneeName)
			}
		}
		return fmt.Sprintf("Update %q: %s", proposal.Story.Target.DisplayName, strings.Join(changes, "; "))
	case ActionFeedbackStatus:
		return fmt.Sprintf("Set %q feedback status to %s", proposal.Feedback.Target.DisplayName, proposal.Feedback.Status)
	default:
		return ""
	}
}

func resolveObjectiveAction(validated validatedRequest, model ModelObjectiveAction) (*ObjectiveAction, error) {
	target, err := resolveTarget(validated.targets, model.TargetReference, TargetObjective)
	if err != nil {
		return nil, err
	}
	if model.Health == nil || !validObjectiveHealth(*model.Health) {
		return nil, fmt.Errorf("%w: objective proposal requires a supported health", ErrInvalidDecision)
	}
	checkIn, err := normalizeOptionalCheckIn(model.CheckIn)
	if err != nil {
		return nil, err
	}
	if err := validateCheckInGrounding(checkIn, validated.checkInSources); err != nil {
		return nil, err
	}
	health := *model.Health
	return &ObjectiveAction{
		Target:  snapshot(target),
		Health:  &health,
		CheckIn: checkIn,
	}, nil
}

func resolveKeyResultAction(validated validatedRequest, model ModelKeyResultAction) (*KeyResultAction, error) {
	target, err := resolveTarget(validated.targets, model.TargetReference, TargetKeyResult)
	if err != nil {
		return nil, err
	}
	if model.CurrentValue == nil || math.IsNaN(*model.CurrentValue) || math.IsInf(*model.CurrentValue, 0) {
		return nil, fmt.Errorf("%w: key-result proposal requires a finite current value", ErrInvalidDecision)
	}
	checkIn, err := normalizeOptionalCheckIn(model.CheckIn)
	if err != nil {
		return nil, err
	}
	if err := validateCheckInGrounding(checkIn, validated.checkInSources); err != nil {
		return nil, err
	}
	value := *model.CurrentValue
	return &KeyResultAction{
		Target:       snapshot(target),
		CurrentValue: &value,
		CheckIn:      checkIn,
	}, nil
}

func resolveStoryAction(validated validatedRequest, model ModelStoryAction) (*StoryAction, error) {
	target, err := resolveTarget(validated.targets, model.TargetReference, TargetStory)
	if err != nil {
		return nil, err
	}
	if model.DueDate == nil && model.Status == nil && model.Assignee == nil {
		return nil, fmt.Errorf("%w: story proposal has no changes", ErrInvalidDecision)
	}
	action := &StoryAction{Target: snapshot(target)}
	if model.DueDate != nil {
		dueDate, err := resolveDateChange(*model.DueDate)
		if err != nil {
			return nil, err
		}
		action.DueDate = &dueDate
	}
	if model.Status != nil {
		choice, err := resolveChoice(validated.choices, model.Status.ChoiceReference, ChoiceStoryStatus, target.TeamID)
		if err != nil {
			return nil, err
		}
		action.Status = &StatusChange{StatusID: choice.ID, StatusName: choice.DisplayName}
	}
	if model.Assignee != nil {
		assignee, err := resolveAssigneeChange(validated.choices, target.TeamID, *model.Assignee)
		if err != nil {
			return nil, err
		}
		action.Assignee = &assignee
	}
	return action, nil
}

func resolveFeedbackStatusAction(validated validatedRequest, model ModelFeedbackStatusAction) (*FeedbackStatusAction, error) {
	target, err := resolveTarget(validated.targets, model.TargetReference, TargetFeedback)
	if err != nil {
		return nil, err
	}
	if !validFeedbackStatus(model.Status) {
		return nil, fmt.Errorf("%w: unsupported feedback status %q", ErrInvalidDecision, model.Status)
	}
	return &FeedbackStatusAction{Target: snapshot(target), Status: model.Status}, nil
}

func resolveTarget(targets map[string]AuthorizedTarget, reference string, kind TargetKind) (AuthorizedTarget, error) {
	target, ok := targets[strings.TrimSpace(reference)]
	if !ok {
		return AuthorizedTarget{}, fmt.Errorf("%w: proposal cites unknown target %q", ErrInvalidDecision, reference)
	}
	if target.Kind != kind {
		return AuthorizedTarget{}, fmt.Errorf("%w: target %q is %q, not %q", ErrInvalidDecision, reference, target.Kind, kind)
	}
	return target, nil
}

func resolveChoice(choices map[string]AuthorizedChoice, reference string, kind ChoiceKind, targetTeamID uuid.UUID) (AuthorizedChoice, error) {
	choice, ok := choices[strings.TrimSpace(reference)]
	if !ok {
		return AuthorizedChoice{}, fmt.Errorf("%w: proposal cites unknown choice %q", ErrInvalidDecision, reference)
	}
	if choice.Kind != kind {
		return AuthorizedChoice{}, fmt.Errorf("%w: choice %q is %q, not %q", ErrInvalidDecision, reference, choice.Kind, kind)
	}
	if choice.TeamID != targetTeamID {
		return AuthorizedChoice{}, fmt.Errorf("%w: choice %q does not belong to the target team", ErrInvalidDecision, reference)
	}
	return choice, nil
}

func resolveDateChange(model ModelDateChange) (DateChange, error) {
	model.Date = strings.TrimSpace(model.Date)
	switch model.Operation {
	case DateClear:
		if model.Date != "" {
			return DateChange{}, fmt.Errorf("%w: clear due-date operation cannot include a date", ErrInvalidDecision)
		}
		return DateChange{Operation: DateClear}, nil
	case DateSet:
		date, err := time.Parse(dateOnlyLayout, model.Date)
		if err != nil || date.Format(dateOnlyLayout) != model.Date {
			return DateChange{}, fmt.Errorf("%w: due date must be a valid YYYY-MM-DD value", ErrInvalidDecision)
		}
		return DateChange{Operation: DateSet, Date: model.Date}, nil
	default:
		return DateChange{}, fmt.Errorf("%w: unsupported due-date operation %q", ErrInvalidDecision, model.Operation)
	}
}

func resolveAssigneeChange(choices map[string]AuthorizedChoice, teamID uuid.UUID, model ModelAssigneeChange) (AssigneeChange, error) {
	model.ChoiceReference = strings.TrimSpace(model.ChoiceReference)
	switch model.Operation {
	case AssigneeUnassign:
		if model.ChoiceReference != "" {
			return AssigneeChange{}, fmt.Errorf("%w: unassign cannot include an assignee choice", ErrInvalidDecision)
		}
		return AssigneeChange{Operation: AssigneeUnassign}, nil
	case AssigneeAssign:
		choice, err := resolveChoice(choices, model.ChoiceReference, ChoiceStoryAssignee, teamID)
		if err != nil {
			return AssigneeChange{}, err
		}
		assigneeID := choice.ID
		return AssigneeChange{
			Operation:    AssigneeAssign,
			AssigneeID:   &assigneeID,
			AssigneeName: choice.DisplayName,
		}, nil
	default:
		return AssigneeChange{}, fmt.Errorf("%w: unsupported assignee operation %q", ErrInvalidDecision, model.Operation)
	}
}

func normalizeOptionalCheckIn(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil, fmt.Errorf("%w: check-in cannot be empty", ErrInvalidDecision)
	}
	if utf8.RuneCountInString(normalized) > maxCheckInRunes {
		return nil, fmt.Errorf("%w: check-in exceeds %d runes", ErrInvalidDecision, maxCheckInRunes)
	}
	if strings.ContainsRune(normalized, '\x00') {
		return nil, fmt.Errorf("%w: check-in contains a null byte", ErrInvalidDecision)
	}
	return &normalized, nil
}

func snapshot(target AuthorizedTarget) TargetSnapshot {
	return TargetSnapshot{
		ID:                target.ID,
		TeamID:            target.TeamID,
		DisplayName:       target.DisplayName,
		ExpectedUpdatedAt: target.ExpectedUpdatedAt.UTC(),
	}
}

func payloadCount(model *ModelActionProposal) int {
	count := 0
	if model.Objective != nil {
		count++
	}
	if model.KeyResult != nil {
		count++
	}
	if model.Story != nil {
		count++
	}
	if model.Feedback != nil {
		count++
	}
	return count
}

func mismatchedPayload(kind ActionKind) error {
	return fmt.Errorf("%w: proposal kind %q does not match its action payload", ErrInvalidDecision, kind)
}

func validObjectiveHealth(health ObjectiveHealth) bool {
	switch health {
	case ObjectiveHealthAtRisk, ObjectiveHealthOnTrack, ObjectiveHealthOffTrack:
		return true
	default:
		return false
	}
}

func validFeedbackStatus(status FeedbackStatus) bool {
	switch status {
	case FeedbackStatusPending,
		FeedbackStatusReviewing,
		FeedbackStatusPlanned,
		FeedbackStatusInProgress,
		FeedbackStatusCompleted,
		FeedbackStatusClosed:
		return true
	default:
		return false
	}
}

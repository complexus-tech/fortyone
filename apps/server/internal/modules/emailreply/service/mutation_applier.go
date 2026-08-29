package emailreply

import (
	"context"
	"errors"
	"fmt"
	"time"

	feedbackdomain "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
)

// ProposalVersionReader re-reads the current entity version immediately
// before the domain compare-and-swap call. This distinguishes a user-facing
// stale preview from transient database/service failures without weakening the
// domain service's atomic CAS.
type ProposalVersionReader interface {
	CurrentVersion(ctx context.Context, proposal ActionProposal) (time.Time, error)
}

// DomainMutationApplier maps confirmed inert proposals onto the existing
// domain-service CAS methods. No repository write bypasses domain side effects.
type DomainMutationApplier struct {
	objectives objectiveMutationPort
	keyResults keyResultMutationPort
	stories    storyMutationPort
	feedback   feedbackMutationPort
	versions   ProposalVersionReader
}

func NewDomainMutationApplier(
	objectiveService objectiveMutationBackend,
	keyResultService keyResultMutationBackend,
	storyService storyMutationBackend,
	feedbackService feedbackMutationBackend,
	versionReader ProposalVersionReader,
) (*DomainMutationApplier, error) {
	if objectiveService == nil || keyResultService == nil || storyService == nil || feedbackService == nil || versionReader == nil {
		return nil, errors.New("all email reply mutation services are required")
	}
	return &DomainMutationApplier{
		objectives: objectiveMutationAdapter{backend: objectiveService},
		keyResults: keyResultMutationAdapter{backend: keyResultService},
		stories:    storyMutationAdapter{backend: storyService},
		feedback:   feedbackMutationAdapter{backend: feedbackService},
		versions:   versionReader,
	}, nil
}

func (applier *DomainMutationApplier) Apply(ctx context.Context, proposal ActionProposal) error {
	if applier == nil {
		return errors.New("email reply mutation applier is not configured")
	}
	currentVersion, err := applier.versions.CurrentVersion(ctx, proposal)
	if err != nil {
		return normalizeMutationBackendError(err)
	}
	// A version mismatch is still passed to the domain CAS method so its
	// desired-state retry check can recognize a write that committed before a
	// previous proposal-completion attempt crashed. The processor performs a
	// second, entity-specific reconciliation for domain services without that
	// built-in behavior.
	_ = currentVersion
	err = nil
	switch proposal.Kind {
	case ActionObjectiveUpdate:
		if proposal.Objective == nil || proposal.Objective.Health == nil {
			return errors.New("objective proposal is incomplete")
		}
		comment := optionalText(proposal.Objective.CheckIn)
		err = applier.objectives.ApplyObjectiveHealth(ctx, objectiveHealthCommand{
			ObjectiveID: proposal.Objective.Target.ID, WorkspaceID: proposal.WorkspaceID,
			ActorID: proposal.ActorID, ExpectedUpdatedAt: proposal.Objective.Target.ExpectedUpdatedAt,
			Health: string(*proposal.Objective.Health), CheckIn: comment,
		})
	case ActionKeyResultUpdate:
		if proposal.KeyResult == nil || proposal.KeyResult.CurrentValue == nil {
			return errors.New("key result proposal is incomplete")
		}
		err = applier.keyResults.ApplyKeyResultValue(ctx, keyResultValueCommand{
			KeyResultID: proposal.KeyResult.Target.ID, WorkspaceID: proposal.WorkspaceID,
			ActorID: proposal.ActorID, ExpectedUpdatedAt: proposal.KeyResult.Target.ExpectedUpdatedAt,
			CurrentValue: *proposal.KeyResult.CurrentValue, CheckIn: optionalText(proposal.KeyResult.CheckIn),
		})
	case ActionStoryUpdate:
		if proposal.Story == nil {
			return errors.New("story proposal is incomplete")
		}
		changes, updateErr := storyProposalChanges(*proposal.Story)
		if updateErr != nil {
			return updateErr
		}
		err = applier.stories.ApplyStoryMutation(ctx, storyMutationCommand{
			StoryID: proposal.Story.Target.ID, WorkspaceID: proposal.WorkspaceID,
			ActorID: proposal.ActorID, ExpectedUpdatedAt: proposal.Story.Target.ExpectedUpdatedAt,
			Changes: changes,
		})
	case ActionFeedbackStatus:
		if proposal.Feedback == nil {
			return errors.New("feedback proposal is incomplete")
		}
		err = applier.feedback.ApplyFeedbackStatus(ctx, feedbackStatusCommand{
			ItemID: proposal.Feedback.Target.ID, WorkspaceID: proposal.WorkspaceID,
			ActorID: proposal.ActorID, ExpectedUpdatedAt: proposal.Feedback.Target.ExpectedUpdatedAt,
			Status: string(proposal.Feedback.Status),
		})
	default:
		return fmt.Errorf("unsupported email action kind %q", proposal.Kind)
	}
	return err
}

func storyProposalChanges(action StoryAction) (storyMutationChanges, error) {
	changes := storyMutationChanges{}
	if action.DueDate != nil {
		changes.DueDateSet = true
		switch action.DueDate.Operation {
		case DateClear:
			changes.DueDate = nil
		case DateSet:
			value, err := time.Parse("2006-01-02", action.DueDate.Date)
			if err != nil {
				return storyMutationChanges{}, fmt.Errorf("parse confirmed story due date: %w", err)
			}
			value = value.UTC()
			changes.DueDate = &value
		default:
			return storyMutationChanges{}, errors.New("unsupported story due-date operation")
		}
	}
	if action.Status != nil {
		changes.StatusSet = true
		changes.StatusID = action.Status.StatusID
	}
	if action.Assignee != nil {
		changes.AssigneeSet = true
		switch action.Assignee.Operation {
		case AssigneeUnassign:
			changes.AssigneeID = nil
		case AssigneeAssign:
			if action.Assignee.AssigneeID == nil {
				return storyMutationChanges{}, errors.New("confirmed story assignee is missing")
			}
			assigneeID := *action.Assignee.AssigneeID
			changes.AssigneeID = &assigneeID
		default:
			return storyMutationChanges{}, errors.New("unsupported story assignee operation")
		}
	}
	if changes.empty() {
		return storyMutationChanges{}, errors.New("confirmed story proposal has no changes")
	}
	return changes, nil
}

func optionalText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func normalizeMutationBackendError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, objectivesdomain.ErrVersionConflict) ||
		errors.Is(err, keyresultsdomain.ErrVersionConflict) ||
		errors.Is(err, storydomain.ErrStoryChanged) ||
		errors.Is(err, feedbackdomain.ErrVersionConflict) {
		return errors.Join(ErrActionConflict, err)
	}
	if errors.Is(err, objectivesdomain.ErrNotFound) ||
		errors.Is(err, keyresultsdomain.ErrNotFound) ||
		errors.Is(err, storydomain.ErrNotFound) ||
		errors.Is(err, feedbackdomain.ErrNotFound) {
		return errors.Join(ErrActionUnauthorized, err)
	}
	return err
}

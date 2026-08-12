package emailreply

import (
	"context"
	"errors"
	"fmt"
	"time"

	emailagent "github.com/complexus-tech/projects-api/internal/modules/emailagent/service"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

type objectiveMutationService interface {
	UpdateExternalUserActionIfUnchanged(
		ctx context.Context,
		id, workspaceID, userID uuid.UUID,
		expectedUpdatedAt time.Time,
		comment string,
		updates map[string]any,
	) error
}

type keyResultMutationService interface {
	UpdateExternalUserActionIfUnchanged(
		ctx context.Context,
		id, workspaceID, userID uuid.UUID,
		expectedUpdatedAt time.Time,
		updates map[string]any,
		comment string,
	) error
}

type storyMutationService interface {
	UpdateExternalUserActionIfUnchanged(
		ctx context.Context,
		actorID, storyID, workspaceID uuid.UUID,
		expectedUpdatedAt time.Time,
		updates map[string]any,
	) error
}

type feedbackMutationService interface {
	UpdateItemStatusIfUnchanged(
		ctx context.Context,
		workspaceID, itemID uuid.UUID,
		expectedUpdatedAt time.Time,
		input feedback.CoreUpdateItemStatusInput,
	) (feedback.CoreItem, error)
}

// ProposalVersionReader re-reads the current entity version immediately
// before the domain compare-and-swap call. This distinguishes a user-facing
// stale preview from transient database/service failures without weakening the
// domain service's atomic CAS.
type ProposalVersionReader interface {
	CurrentVersion(ctx context.Context, proposal emailagent.ActionProposal) (time.Time, error)
}

// DomainMutationApplier maps confirmed inert proposals onto the existing
// domain-service CAS methods. No repository write bypasses domain side effects.
type DomainMutationApplier struct {
	objectives objectiveMutationService
	keyResults keyResultMutationService
	stories    storyMutationService
	feedback   feedbackMutationService
	versions   ProposalVersionReader
}

func NewDomainMutationApplier(
	objectiveService objectiveMutationService,
	keyResultService keyResultMutationService,
	storyService storyMutationService,
	feedbackService feedbackMutationService,
	versionReader ProposalVersionReader,
) (*DomainMutationApplier, error) {
	if objectiveService == nil || keyResultService == nil || storyService == nil || feedbackService == nil || versionReader == nil {
		return nil, errors.New("all email reply mutation services are required")
	}
	return &DomainMutationApplier{
		objectives: objectiveService,
		keyResults: keyResultService,
		stories:    storyService,
		feedback:   feedbackService,
		versions:   versionReader,
	}, nil
}

func (applier *DomainMutationApplier) Apply(ctx context.Context, proposal emailagent.ActionProposal) error {
	if applier == nil {
		return errors.New("email reply mutation applier is not configured")
	}
	currentVersion, err := applier.versions.CurrentVersion(ctx, proposal)
	if err != nil {
		return normalizeMutationError(err)
	}
	// A version mismatch is still passed to the domain CAS method so its
	// desired-state retry check can recognize a write that committed before a
	// previous proposal-completion attempt crashed. The processor performs a
	// second, entity-specific reconciliation for domain services without that
	// built-in behavior.
	_ = currentVersion
	err = nil
	switch proposal.Kind {
	case emailagent.ActionObjectiveUpdate:
		if proposal.Objective == nil || proposal.Objective.Health == nil {
			return errors.New("objective proposal is incomplete")
		}
		comment := optionalText(proposal.Objective.CheckIn)
		err = applier.objectives.UpdateExternalUserActionIfUnchanged(
			ctx,
			proposal.Objective.Target.ID,
			proposal.WorkspaceID,
			proposal.ActorID,
			proposal.Objective.Target.ExpectedUpdatedAt,
			comment,
			map[string]any{"health": string(*proposal.Objective.Health)},
		)
	case emailagent.ActionKeyResultUpdate:
		if proposal.KeyResult == nil || proposal.KeyResult.CurrentValue == nil {
			return errors.New("key result proposal is incomplete")
		}
		err = applier.keyResults.UpdateExternalUserActionIfUnchanged(
			ctx,
			proposal.KeyResult.Target.ID,
			proposal.WorkspaceID,
			proposal.ActorID,
			proposal.KeyResult.Target.ExpectedUpdatedAt,
			map[string]any{"current_value": *proposal.KeyResult.CurrentValue},
			optionalText(proposal.KeyResult.CheckIn),
		)
	case emailagent.ActionStoryUpdate:
		if proposal.Story == nil {
			return errors.New("story proposal is incomplete")
		}
		updates, updateErr := storyProposalUpdates(*proposal.Story)
		if updateErr != nil {
			return updateErr
		}
		err = applier.stories.UpdateExternalUserActionIfUnchanged(
			ctx,
			proposal.ActorID,
			proposal.Story.Target.ID,
			proposal.WorkspaceID,
			proposal.Story.Target.ExpectedUpdatedAt,
			updates,
		)
	case emailagent.ActionFeedbackStatus:
		if proposal.Feedback == nil {
			return errors.New("feedback proposal is incomplete")
		}
		_, err = applier.feedback.UpdateItemStatusIfUnchanged(
			ctx,
			proposal.WorkspaceID,
			proposal.Feedback.Target.ID,
			proposal.Feedback.Target.ExpectedUpdatedAt,
			feedback.CoreUpdateItemStatusInput{
				Status: string(proposal.Feedback.Status), ActorID: proposal.ActorID,
			},
		)
	default:
		return fmt.Errorf("unsupported email action kind %q", proposal.Kind)
	}
	return normalizeMutationError(err)
}

func storyProposalUpdates(action emailagent.StoryAction) (map[string]any, error) {
	updates := make(map[string]any, 3)
	if action.DueDate != nil {
		switch action.DueDate.Operation {
		case emailagent.DateClear:
			updates["end_date"] = nil
		case emailagent.DateSet:
			value, err := time.Parse("2006-01-02", action.DueDate.Date)
			if err != nil {
				return nil, fmt.Errorf("parse confirmed story due date: %w", err)
			}
			updates["end_date"] = value.UTC()
		default:
			return nil, errors.New("unsupported story due-date operation")
		}
	}
	if action.Status != nil {
		updates["status_id"] = action.Status.StatusID
	}
	if action.Assignee != nil {
		switch action.Assignee.Operation {
		case emailagent.AssigneeUnassign:
			updates["assignee_id"] = nil
		case emailagent.AssigneeAssign:
			if action.Assignee.AssigneeID == nil {
				return nil, errors.New("confirmed story assignee is missing")
			}
			updates["assignee_id"] = *action.Assignee.AssigneeID
		default:
			return nil, errors.New("unsupported story assignee operation")
		}
	}
	if len(updates) == 0 {
		return nil, errors.New("confirmed story proposal has no changes")
	}
	return updates, nil
}

func normalizeMutationError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, objectives.ErrVersionConflict) ||
		errors.Is(err, keyresults.ErrVersionConflict) ||
		errors.Is(err, stories.ErrStoryChanged) ||
		errors.Is(err, feedback.ErrVersionConflict) {
		return fmt.Errorf("%w: %v", ErrActionConflict, err)
	}
	if errors.Is(err, objectives.ErrNotFound) ||
		errors.Is(err, keyresults.ErrNotFound) ||
		errors.Is(err, stories.ErrNotFound) ||
		errors.Is(err, feedback.ErrNotFound) {
		return fmt.Errorf("%w: %v", ErrActionUnauthorized, err)
	}
	return err
}

func optionalText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

package maya

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

func (s *Service) applyActions(ctx context.Context, actions []CoreAction) ([]CoreAction, error) {
	return s.applyActionsWithOptions(ctx, actions, actionApplicationOptions{})
}

func (s *Service) applyActionsWithOptions(ctx context.Context, actions []CoreAction, options actionApplicationOptions) ([]CoreAction, error) {
	appliedAt := s.clock.Now().UTC()
	applied := make([]CoreAction, len(actions))
	copy(applied, actions)
	var persistenceErr error
	markApplied := func(action *CoreAction) {
		persistenceErr = errors.Join(persistenceErr, s.markActionApplied(ctx, action, appliedAt))
	}
	markFailed := func(action *CoreAction, message string) {
		persistenceErr = errors.Join(persistenceErr, s.markActionFailed(ctx, action, message))
	}
	scheduleIndexes := make([]int, 0, len(applied))
	hasAssignmentAction := false
	for index := range applied {
		if applied[index].Type == ActionTypeScheduleWorkBlock {
			scheduleIndexes = append(scheduleIndexes, index)
		}
		if applied[index].Type == ActionTypeAssignStory {
			hasAssignmentAction = true
		}
	}
	scheduleApplied := len(scheduleIndexes) == 0
	var scheduleState *appliedScheduleState
	if len(scheduleIndexes) > 0 {
		scheduleActions := make([]CoreAction, 0, len(scheduleIndexes))
		for _, index := range scheduleIndexes {
			scheduleActions = append(scheduleActions, applied[index])
		}
		state, err := s.applyScheduleActionsAtomically(ctx, scheduleActions)
		if err != nil {
			for _, index := range scheduleIndexes {
				markFailed(&applied[index], err.Error())
			}
		} else {
			scheduleApplied = true
			scheduleState = &state
		}
	}
	if len(options.StoryUpdates) > 0 && !hasAssignmentAction {
		if len(scheduleIndexes) == 0 {
			if err := s.applyStoredPlanStoryUpdates(ctx, actions, options); err != nil {
				return applied, err
			}
		} else if scheduleApplied {
			if err := s.applyStoredPlanStoryUpdates(ctx, actions, options); err != nil {
				scheduleApplied = false
				message := "schedule was restored because the scheduling preferences could not be saved"
				if scheduleState != nil {
					if restoreErr := s.restoreScheduleState(ctx, *scheduleState); restoreErr != nil {
						message = errors.Join(errors.New(message), restoreErr).Error()
					}
				}
				for _, index := range scheduleIndexes {
					markFailed(&applied[index], message+": "+err.Error())
				}
			} else if scheduleState != nil {
				if err := s.refreshAppliedScheduleAfterAssignment(ctx, *scheduleState); err != nil {
					scheduleApplied = false
					message := "schedule was restored because its story-version ownership could not be refreshed"
					if restoreErr := s.restoreScheduleState(ctx, *scheduleState); restoreErr != nil {
						message = errors.Join(errors.New(message), restoreErr).Error()
					}
					for _, index := range scheduleIndexes {
						markFailed(&applied[index], message+": "+err.Error())
					}
				}
			}
		}
	}
	dispatchSchedule := false
	for i, action := range applied {
		if action.Type == ActionTypeScheduleWorkBlock {
			continue
		}
		if action.Type == ActionTypeAssignStory && !scheduleApplied {
			markFailed(&applied[i], "assignment was not applied because the schedule could not be committed")
			continue
		}
		additionalUpdates := map[string]any(nil)
		if action.Type == ActionTypeAssignStory {
			additionalUpdates = options.StoryUpdates
		}
		if err := s.applyActionWithUpdates(ctx, action, additionalUpdates); err != nil {
			markFailed(&applied[i], err.Error())
			if action.Type == ActionTypeAssignStory && scheduleState != nil {
				rollbackErr := s.restoreScheduleState(ctx, *scheduleState)
				scheduleApplied = false
				message := "schedule was restored because the dependent story assignment failed"
				if rollbackErr != nil {
					message = errors.Join(errors.New(message), rollbackErr).Error()
				} else {
					dispatchSchedule = true
				}
				for _, index := range scheduleIndexes {
					markFailed(&applied[index], message)
				}
			}
			continue
		}
		if action.Type == ActionTypeAssignStory && scheduleState != nil {
			if err := s.refreshAppliedScheduleAfterAssignment(ctx, *scheduleState); err != nil {
				// The assignment is already committed and must never be reported as
				// rolled back. Leave provider delivery fenced by the stale ownership
				// watermark; the assignment event/recovery sweep will retry the exact
				// schedule against the current story version.
				markApplied(&applied[i])
				scheduleApplied = false
				for _, index := range scheduleIndexes {
					markFailed(&applied[index], "schedule ownership refresh failed after assignment: "+err.Error())
				}
				continue
			}
		}
		markApplied(&applied[i])
	}
	if scheduleApplied {
		for _, index := range scheduleIndexes {
			markApplied(&applied[index])
		}
		dispatchSchedule = scheduleState != nil
	}
	if dispatchSchedule && scheduleState != nil {
		// The database outbox is the delivery contract. A transient provider
		// failure must not change the result of the durable local mutation.
		_ = scheduleState.calendar.DispatchScheduleEventOutbox(ctx, scheduleState.userID)
	}
	return applied, persistenceErr
}

func (s *Service) applyStoredPlanStoryUpdates(ctx context.Context, actions []CoreAction, options actionApplicationOptions) error {
	if len(options.StoryUpdates) == 0 {
		return nil
	}
	storyID := options.StoryID
	workspaceID := options.WorkspaceID
	if len(actions) > 0 {
		storyID = actions[0].StoryID
		workspaceID = actions[0].WorkspaceID
	}
	if storyID == uuid.Nil || workspaceID == uuid.Nil {
		return fmt.Errorf("%w: stored work plan has no story identity", ErrInvalidPlanInput)
	}
	return s.stories.UpdateAutomationIfUnchanged(
		ctx,
		s.mayaActorID,
		storyID,
		workspaceID,
		options.ExpectedStoryUpdatedAt,
		options.StoryUpdates,
		"Maya saved the approved scheduling preferences for this work plan.",
	)
}

type appliedScheduleState struct {
	calendar       ScheduleCalendarService
	workspaceID    uuid.UUID
	userID         uuid.UUID
	storyID        uuid.UUID
	previousOwners []scheduleOwnerState
	segments       []ScheduleSegmentInput
}

type scheduleOwnerState struct {
	userID        uuid.UUID
	segments      []ScheduleSegmentInput
	keepOwnership bool
}

func (s *Service) applyScheduleActionsAtomically(ctx context.Context, actions []CoreAction) (appliedScheduleState, error) {
	if len(actions) == 0 {
		return appliedScheduleState{}, nil
	}
	first := actions[0]
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return appliedScheduleState{}, err
	}
	if first.Payload.ScheduleBlock == nil {
		return appliedScheduleState{}, fmt.Errorf("missing schedule block payload")
	}
	userID := first.Payload.ScheduleBlock.UserID
	expectedStoryUpdatedAt := first.Payload.ScheduleBlock.ExpectedStoryUpdatedAt
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return appliedScheduleState{}, err
	}
	ownerIDs, err := scheduleRepo.ListMayaScheduleOwners(ctx, first.WorkspaceID, first.StoryID)
	if err != nil {
		return appliedScheduleState{}, err
	}
	if !slices.Contains(ownerIDs, userID) {
		ownerIDs = append(ownerIDs, userID)
	}
	previousOwners := make([]scheduleOwnerState, 0, len(ownerIDs))
	for _, ownerID := range ownerIDs {
		currentBlocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, first.WorkspaceID, ownerID, first.StoryID)
		if err != nil {
			return appliedScheduleState{}, err
		}
		previousOwnership, err := scheduleCalendar.MayaScheduleOwnershipExists(ctx, first.WorkspaceID, ownerID, first.StoryID)
		if err != nil {
			return appliedScheduleState{}, err
		}
		previousOwners = append(previousOwners, scheduleOwnerState{
			userID: ownerID, segments: mayaSegmentsFromBlocks(currentBlocks), keepOwnership: previousOwnership,
		})
	}
	segments := make([]ScheduleSegmentInput, 0, len(actions))
	preemptBlockIDs := []uuid.UUID(nil)
	for _, action := range actions {
		payload := action.Payload.ScheduleBlock
		if payload == nil || action.WorkspaceID != first.WorkspaceID || action.StoryID != first.StoryID || payload.UserID != userID {
			return appliedScheduleState{}, fmt.Errorf("schedule actions do not belong to one story and user")
		}
		if !payload.ExpectedStoryUpdatedAt.Equal(expectedStoryUpdatedAt) {
			return appliedScheduleState{}, fmt.Errorf("schedule actions were planned from different story versions")
		}
		if len(payload.PreemptBlockIDs) > 0 {
			if len(preemptBlockIDs) == 0 {
				preemptBlockIDs = append([]uuid.UUID(nil), payload.PreemptBlockIDs...)
			} else if !slices.Equal(preemptBlockIDs, payload.PreemptBlockIDs) {
				return appliedScheduleState{}, fmt.Errorf("schedule actions were planned with different preemption sets")
			}
		}
		if payload.Operation == ScheduleBlockOperationRetain {
			continue
		}
		if payload.Operation != "" && payload.Operation != ScheduleBlockOperationUpsert {
			return appliedScheduleState{}, fmt.Errorf("unsupported initial schedule operation %q", payload.Operation)
		}
		segments = append(segments, ScheduleSegmentInput{
			SegmentIndex: payload.SegmentIndex,
			Title:        payload.Title,
			StartAt:      payload.StartAt,
			EndAt:        payload.EndAt,
		})
	}
	if _, err := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
		WorkspaceID: first.WorkspaceID, UserID: userID, StoryID: first.StoryID,
		ExpectedStoryUpdatedAt: &expectedStoryUpdatedAt, Segments: segments, PreemptBlockIDs: preemptBlockIDs, KeepOwnership: true,
	}); err != nil {
		return appliedScheduleState{}, err
	}
	state := appliedScheduleState{
		calendar: scheduleCalendar, workspaceID: first.WorkspaceID, userID: userID,
		storyID: first.StoryID, previousOwners: previousOwners, segments: segments,
	}
	for _, previousOwner := range previousOwners {
		if previousOwner.userID == userID {
			continue
		}
		if _, err := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
			WorkspaceID: first.WorkspaceID, UserID: previousOwner.userID, StoryID: first.StoryID,
		}); err != nil {
			return appliedScheduleState{}, errors.Join(err, s.restoreScheduleState(ctx, state))
		}
	}
	return state, nil
}

func (s *Service) refreshAppliedScheduleAfterAssignment(ctx context.Context, state appliedScheduleState) error {
	story, err := s.stories.Get(ctx, state.storyID, state.workspaceID)
	if err != nil {
		return err
	}
	if !story.AutoSchedulingEnabled || story.Assignee == nil || *story.Assignee != state.userID {
		return errors.New("story assignment no longer matches the committed Maya schedule")
	}
	_, err = state.calendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
		WorkspaceID:            state.workspaceID,
		UserID:                 state.userID,
		StoryID:                state.storyID,
		ExpectedStoryUpdatedAt: &story.UpdatedAt,
		Segments:               state.segments,
		KeepOwnership:          true,
		Locked:                 story.AutoSchedulingLocked,
	})
	return err
}

func (s *Service) restoreScheduleState(ctx context.Context, state appliedScheduleState) error {
	previousByOwner := make(map[uuid.UUID]scheduleOwnerState, len(state.previousOwners))
	for _, previous := range state.previousOwners {
		previousByOwner[previous.userID] = previous
	}
	selected, existed := previousByOwner[state.userID]
	if !existed {
		selected = scheduleOwnerState{userID: state.userID}
	}
	var restoreErr error
	if _, err := state.calendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
		WorkspaceID: state.workspaceID, UserID: selected.userID, StoryID: state.storyID,
		Segments: selected.segments, KeepOwnership: selected.keepOwnership,
	}); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}
	for _, previous := range state.previousOwners {
		if previous.userID == state.userID {
			continue
		}
		if _, err := state.calendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
			WorkspaceID: state.workspaceID, UserID: previous.userID, StoryID: state.storyID,
			Segments: previous.segments, KeepOwnership: previous.keepOwnership,
		}); err != nil {
			restoreErr = errors.Join(restoreErr, err)
		}
	}
	return restoreErr
}

func mayaSegmentsFromBlocks(blocks []ScheduleBlock) []ScheduleSegmentInput {
	segments := make([]ScheduleSegmentInput, 0, len(blocks))
	for _, block := range blocks {
		segments = append(segments, ScheduleSegmentInput{
			SegmentIndex: block.SegmentIndex, Title: block.Title, StartAt: block.StartAt, EndAt: block.EndAt,
		})
	}
	return segments
}

func (s *Service) markActionApplied(ctx context.Context, action *CoreAction, appliedAt time.Time) error {
	action.Status = ActionStatusApplied
	action.AppliedAt = &appliedAt
	action.Error = nil
	if err := s.repo.MarkActionApplied(ctx, action.ID); err != nil {
		return fmt.Errorf("mark Maya action %s applied: %w", action.ID, err)
	}
	return nil
}

func (s *Service) markActionFailed(ctx context.Context, action *CoreAction, message string) error {
	action.Status = ActionStatusFailed
	action.Error = &message
	if err := s.repo.MarkActionFailed(ctx, action.ID, message); err != nil {
		return fmt.Errorf("mark Maya action %s failed: %w", action.ID, err)
	}
	return nil
}

func (s *Service) applyActionWithUpdates(ctx context.Context, action CoreAction, additionalUpdates map[string]any) error {
	switch action.Type {
	case ActionTypeAssignStory:
		if action.Payload.AssignStory == nil {
			return fmt.Errorf("missing assign story payload")
		}
		updates := make(map[string]any, len(additionalUpdates)+1)
		for key, value := range additionalUpdates {
			updates[key] = value
		}
		updates["assignee_id"] = action.Payload.AssignStory.AssigneeID
		return s.stories.UpdateAutomationIfUnchanged(ctx, s.mayaActorID, action.StoryID, action.WorkspaceID, action.Payload.AssignStory.ExpectedUpdatedAt, updates, action.Reason)
	case ActionTypeScheduleWorkBlock:
		return fmt.Errorf("schedule work blocks must be applied as one atomic set")
	case ActionTypeFlagScheduleRisk:
		return nil
	default:
		return fmt.Errorf("unsupported maya action type: %s", action.Type)
	}
}

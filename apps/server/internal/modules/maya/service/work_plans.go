package maya

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateWorkPlan(ctx context.Context, input CreateWorkPlanInput) (WorkPlan, error) {
	asOf := s.clock.Now().UTC()
	if !input.AutoApply {
		return s.createWorkPlan(ctx, input, asOf)
	}
	if err := s.validate(); err != nil {
		return WorkPlan{}, err
	}
	if input.WorkspaceID == uuid.Nil || input.StoryID == uuid.Nil {
		return WorkPlan{}, fmt.Errorf("%w: story and workspace are required", ErrInvalidPlanInput)
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return WorkPlan{}, err
	}
	var plan WorkPlan
	err = scheduleRepo.WithScheduleStoryLock(ctx, input.WorkspaceID, input.StoryID, func() error {
		var planErr error
		plan, planErr = s.createWorkPlan(ctx, input, asOf)
		return planErr
	})
	return plan, err
}

func (s *Service) ApplyWorkPlan(ctx context.Context, input ApplyWorkPlanInput) (WorkPlan, error) {
	ctx, span := mayaServiceTracer.Start(ctx, "business.core.maya.ApplyWorkPlan")
	defer span.End()

	if err := s.validate(); err != nil {
		return WorkPlan{}, err
	}
	if input.WorkspaceID == uuid.Nil || input.RunID == uuid.Nil || input.TriggeredBy == uuid.Nil {
		return WorkPlan{}, fmt.Errorf("%w: work plan, workspace, and user are required", ErrInvalidPlanInput)
	}

	plan, err := s.repo.GetWorkPlan(ctx, input.RunID, input.WorkspaceID, input.TriggeredBy)
	if err != nil {
		return WorkPlan{}, err
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return WorkPlan{}, err
	}

	var appliedPlan WorkPlan
	err = scheduleRepo.WithScheduleStoryLock(ctx, input.WorkspaceID, plan.Run.StoryID, func() error {
		currentPlan, err := s.repo.GetWorkPlan(ctx, input.RunID, input.WorkspaceID, input.TriggeredBy)
		if err != nil {
			return err
		}
		appliedPlan = currentPlan
		if currentPlan.Run.Status != RunStatusSucceeded {
			return fmt.Errorf("%w: only a completed work-plan preview can be applied", ErrInvalidPlanInput)
		}
		if workPlanHasActionStatus(currentPlan.Actions, ActionStatusFailed) || !workPlanHasActionStatus(currentPlan.Actions, ActionStatusProposed) {
			return nil
		}

		story, err := s.stories.Get(ctx, currentPlan.Run.StoryID, input.WorkspaceID)
		if err != nil {
			return fmt.Errorf("get story for stored Maya plan: %w", err)
		}
		hasAccess, err := scheduleRepo.WorkspaceCanUseMaya(ctx, input.WorkspaceID)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrMayaAccessDenied
		}
		active, err := scheduleRepo.StoryIsActiveForAutoScheduling(ctx, input.WorkspaceID, story.ID)
		if err != nil {
			return err
		}
		if !active {
			return fmt.Errorf("%w: story is not active for auto-scheduling", ErrInvalidPlanInput)
		}
		if story.AutoSchedulingLocked {
			return ErrAutoSchedulingOwnerLocked
		}

		planContext, err := decodePersistedWorkPlanContext(currentPlan.Run.Context)
		if err != nil {
			return err
		}
		if planContext.StoryUpdatedAt.IsZero() || !story.UpdatedAt.Equal(planContext.StoryUpdatedAt) {
			return ErrStoryChanged
		}

		result := storedPlanResult(currentPlan, story, planContext)
		previousBlocks, err := s.listPreviousWorkPlanBlocks(ctx, story, result.SelectedUserID)
		if err != nil {
			return err
		}
		proposedActions := actionsWithStatus(currentPlan.Actions, ActionStatusProposed)
		updatedActions, err := s.applyActionsWithOptions(ctx, proposedActions, actionApplicationOptions{
			ExpectedStoryUpdatedAt: planContext.StoryUpdatedAt,
			StoryUpdates:           storedPlanStoryUpdates(story, planContext.DurationMinutes),
			StoryID:                story.ID,
			WorkspaceID:            story.Workspace,
		})
		appliedPlan.Actions = mergeWorkPlanActions(currentPlan.Actions, updatedActions)
		if err != nil {
			return err
		}
		if workPlanHasActionStatus(appliedPlan.Actions, ActionStatusFailed) {
			return nil
		}
		if err := s.finalizeAppliedWorkPlan(ctx, story, result, previousBlocks); err != nil {
			return fmt.Errorf("finalize stored Maya work plan: %w", err)
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		return appliedPlan, err
	}
	return appliedPlan, nil
}

func decodePersistedWorkPlanContext(payload json.RawMessage) (persistedWorkPlanContext, error) {
	var planContext persistedWorkPlanContext
	if len(payload) == 0 {
		return planContext, fmt.Errorf("%w: work-plan preview context is missing", ErrInvalidPlanInput)
	}
	if err := json.Unmarshal(payload, &planContext); err != nil {
		return planContext, fmt.Errorf("%w: decode work-plan preview context: %v", ErrInvalidPlanInput, err)
	}
	return planContext, nil
}

func workPlanHasActionStatus(actions []CoreAction, status ActionStatus) bool {
	for _, action := range actions {
		if action.Status == status {
			return true
		}
	}
	return false
}

func actionsWithStatus(actions []CoreAction, status ActionStatus) []CoreAction {
	filtered := make([]CoreAction, 0, len(actions))
	for _, action := range actions {
		if action.Status == status {
			filtered = append(filtered, action)
		}
	}
	return filtered
}

func mergeWorkPlanActions(existing, updated []CoreAction) []CoreAction {
	updatedByID := make(map[uuid.UUID]CoreAction, len(updated))
	for _, action := range updated {
		updatedByID[action.ID] = action
	}
	merged := make([]CoreAction, len(existing))
	for index, action := range existing {
		if updatedAction, ok := updatedByID[action.ID]; ok {
			merged[index] = updatedAction
			continue
		}
		merged[index] = action
	}
	return merged
}

func storedPlanStoryUpdates(story Story, durationMinutes int) map[string]any {
	updates := make(map[string]any)
	if !story.AutoSchedulingEnabled {
		updates["auto_scheduling_enabled"] = true
		updates["auto_scheduling_locked"] = false
	}
	if durationMinutes > 0 && (story.EstimatedDurationMinutes == nil || *story.EstimatedDurationMinutes != durationMinutes) {
		updates["estimated_duration_minutes"] = durationMinutes
		if story.MinimumFocusBlockMinutes != nil && *story.MinimumFocusBlockMinutes > durationMinutes {
			updates["minimum_focus_block_minutes"] = durationMinutes
		}
	}
	return updates
}

func storedPlanResult(plan WorkPlan, story Story, planContext persistedWorkPlanContext) PlanResult {
	result := PlanResult{Summary: plan.Run.Summary, DurationMinutes: planContext.DurationMinutes}
	preemptedIDs := make(map[uuid.UUID]struct{})
	for _, action := range plan.Actions {
		if payload := action.Payload.AssignStory; payload != nil {
			selectedUserID := payload.AssigneeID
			result.SelectedUserID = &selectedUserID
		}
		if payload := action.Payload.ScheduleBlock; payload != nil {
			selectedUserID := payload.UserID
			result.SelectedUserID = &selectedUserID
			if result.Timezone == "" {
				result.Timezone = planContext.CandidateTimezones[payload.UserID.String()]
			}
			if payload.Operation == "" || payload.Operation == ScheduleBlockOperationUpsert {
				result.ScheduledMinutes += int(payload.EndAt.Sub(payload.StartAt).Minutes())
			}
			for _, blockID := range payload.PreemptBlockIDs {
				preemptedIDs[blockID] = struct{}{}
			}
		}
		if payload := action.Payload.Risk; payload != nil && payload.RemainingMinutes > result.RemainingMinutes {
			result.RemainingMinutes = payload.RemainingMinutes
		}
	}
	if result.SelectedUserID == nil && story.Assignee != nil {
		selectedUserID := *story.Assignee
		result.SelectedUserID = &selectedUserID
		result.Timezone = planContext.CandidateTimezones[selectedUserID.String()]
	}
	if result.DurationMinutes == 0 {
		result.DurationMinutes = result.ScheduledMinutes + result.RemainingMinutes
	}
	result.PreemptedBlockIDs = make([]uuid.UUID, 0, len(preemptedIDs))
	for blockID := range preemptedIDs {
		result.PreemptedBlockIDs = append(result.PreemptedBlockIDs, blockID)
	}
	return result
}

func (s *Service) listPreviousWorkPlanBlocks(ctx context.Context, story Story, selectedUserID *uuid.UUID) ([]ScheduleBlock, error) {
	if selectedUserID == nil || *selectedUserID == uuid.Nil {
		return nil, nil
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return nil, err
	}
	ownerIDs, err := scheduleRepo.ListMayaScheduleOwners(ctx, story.Workspace, story.ID)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(ownerIDs, *selectedUserID) {
		ownerIDs = append(ownerIDs, *selectedUserID)
	}
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return nil, err
	}
	blocks := make([]ScheduleBlock, 0)
	for _, ownerID := range ownerIDs {
		ownerBlocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, story.Workspace, ownerID, story.ID)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, ownerBlocks...)
	}
	return blocks, nil
}

func (s *Service) createWorkPlan(ctx context.Context, input CreateWorkPlanInput, asOf time.Time) (WorkPlan, error) {
	ctx, span := mayaServiceTracer.Start(ctx, "business.core.maya.CreateWorkPlan")
	defer span.End()

	if err := s.validate(); err != nil {
		span.RecordError(err)
		return WorkPlan{}, err
	}
	if input.Trigger == "" {
		input.Trigger = RunTriggerManual
	}

	story, err := s.stories.Get(ctx, input.StoryID, input.WorkspaceID)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("get story for maya plan: %w", err)
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return WorkPlan{}, err
	}
	hasAccess, err := scheduleRepo.WorkspaceCanUseMaya(ctx, input.WorkspaceID)
	if err != nil {
		return WorkPlan{}, err
	}
	if !hasAccess {
		return WorkPlan{}, ErrMayaAccessDenied
	}
	if input.AutoApply {
		active, activeErr := scheduleRepo.StoryIsActiveForAutoScheduling(ctx, input.WorkspaceID, input.StoryID)
		if activeErr != nil {
			return WorkPlan{}, activeErr
		}
		if !active {
			return WorkPlan{}, fmt.Errorf("%w: story is not active for auto-scheduling", ErrInvalidPlanInput)
		}
	}
	if input.AutoApply && story.AutoSchedulingLocked {
		return WorkPlan{}, ErrAutoSchedulingOwnerLocked
	}
	prePlanUpdates := make(map[string]any)
	if input.AutoApply && !story.AutoSchedulingEnabled {
		prePlanUpdates["auto_scheduling_enabled"] = true
		prePlanUpdates["auto_scheduling_locked"] = false
	}
	if input.AutoApply && input.DurationMinutes > 0 && (story.EstimatedDurationMinutes == nil || *story.EstimatedDurationMinutes != input.DurationMinutes) {
		prePlanUpdates["estimated_duration_minutes"] = input.DurationMinutes
		if story.MinimumFocusBlockMinutes != nil && *story.MinimumFocusBlockMinutes > input.DurationMinutes {
			prePlanUpdates["minimum_focus_block_minutes"] = input.DurationMinutes
		}
	}
	if len(prePlanUpdates) > 0 {
		if err := s.stories.UpdateAutomationIfUnchanged(
			ctx,
			s.mayaActorID,
			story.ID,
			story.Workspace,
			story.UpdatedAt,
			prePlanUpdates,
			"Maya saved the confirmed scheduling preferences for this work plan.",
		); err != nil {
			return WorkPlan{}, fmt.Errorf("enable Maya auto-scheduling: %w", err)
		}
		story, err = s.stories.Get(ctx, input.StoryID, input.WorkspaceID)
		if err != nil {
			return WorkPlan{}, fmt.Errorf("reload story after enabling Maya auto-scheduling: %w", err)
		}
	}
	candidates, contextPayload, err := s.buildCandidates(ctx, input, story)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, err
	}

	run, err := s.repo.CreateRun(ctx, CreateRunInput{
		WorkspaceID: input.WorkspaceID,
		StoryID:     input.StoryID,
		TriggeredBy: input.TriggeredBy,
		Trigger:     input.Trigger,
		Context:     contextPayload,
	})
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("create maya run: %w", err)
	}

	result, err := s.planner.Plan(PlanInput{
		Context:          ctx,
		AsOf:             asOf,
		WorkspaceID:      input.WorkspaceID,
		Story:            story,
		DurationMinutes:  input.DurationMinutes,
		WindowStart:      input.WindowStart,
		WindowEnd:        input.WindowEnd,
		Candidates:       candidates,
		AssignmentReason: input.AssignmentReason,
	})
	if err != nil {
		message := err.Error()
		completed, completeErr := s.repo.CompleteRun(ctx, run.ID, RunStatusFailed, "", &message)
		if completeErr != nil {
			return WorkPlan{}, fmt.Errorf("complete failed maya run: %w", completeErr)
		}
		return WorkPlan{Run: completed}, err
	}
	if result.SelectedUserID != nil {
		for _, candidate := range candidates {
			if candidate.Member.UserID == *result.SelectedUserID {
				result.Timezone = candidate.Timezone
				break
			}
		}
	}
	previousBlocks := []ScheduleBlock{}
	if input.AutoApply && result.SelectedUserID != nil {
		ownerIDs, ownerErr := scheduleRepo.ListMayaScheduleOwners(ctx, story.Workspace, story.ID)
		if ownerErr != nil {
			return WorkPlan{}, ownerErr
		}
		if !slices.Contains(ownerIDs, *result.SelectedUserID) {
			ownerIDs = append(ownerIDs, *result.SelectedUserID)
		}
		scheduleCalendar, calendarErr := s.scheduleCalendarService()
		if calendarErr != nil {
			return WorkPlan{}, calendarErr
		}
		for _, ownerID := range ownerIDs {
			blocks, listErr := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, story.Workspace, ownerID, story.ID)
			if listErr != nil {
				return WorkPlan{}, listErr
			}
			previousBlocks = append(previousBlocks, blocks...)
		}
	}

	for i := range result.Actions {
		result.Actions[i].RunID = run.ID
		result.Actions[i].WorkspaceID = input.WorkspaceID
		result.Actions[i].StoryID = input.StoryID
	}
	actions, err := s.repo.CreateActions(ctx, result.Actions)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("create maya actions: %w", err)
	}
	if input.AutoApply {
		actions, err = s.applyActions(ctx, actions)
		if err != nil {
			span.RecordError(err)
			return WorkPlan{Run: run, Actions: actions}, fmt.Errorf("persist Maya action outcome: %w", err)
		}
	}

	runStatus := RunStatusSucceeded
	var runError *string
	for _, action := range actions {
		if action.Status == ActionStatusFailed {
			runStatus = RunStatusFailed
			if action.Error != nil {
				message := *action.Error
				runError = &message
			}
			break
		}
	}
	completed, err := s.repo.CompleteRun(ctx, run.ID, runStatus, result.Summary, runError)
	if err != nil {
		span.RecordError(err)
		return WorkPlan{}, fmt.Errorf("complete maya run: %w", err)
	}
	if input.AutoApply && runStatus == RunStatusSucceeded {
		if err := s.finalizeAppliedWorkPlan(ctx, story, result, previousBlocks); err != nil {
			return WorkPlan{Run: completed, Actions: actions}, fmt.Errorf("finalize Maya auto-scheduling state: %w", err)
		}
	}
	return WorkPlan{Run: completed, Actions: actions}, nil
}

func (s *Service) finalizeAppliedWorkPlan(ctx context.Context, plannedStory Story, result PlanResult, previousBlocks []ScheduleBlock) error {
	story, err := s.stories.Get(ctx, plannedStory.ID, plannedStory.Workspace)
	if err != nil {
		return err
	}
	if !story.AutoSchedulingEnabled {
		return nil
	}
	if result.SelectedUserID == nil || *result.SelectedUserID == uuid.Nil {
		reason := "Maya could not select an eligible teammate yet."
		return s.stories.UpdateAutomationStateIfUnchanged(
			ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
			AutoSchedulingStatusNeedsOwner, &reason, nil, nil,
		)
	}
	ownerID := *result.SelectedUserID
	if story.Assignee == nil || *story.Assignee != ownerID {
		return errors.New("applied schedule owner does not match the current story assignee")
	}
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return err
	}
	blocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, story.Workspace, ownerID, story.ID)
	if err != nil {
		return err
	}
	segments := mayaSegmentsFromBlocks(blocks)
	status, reason := autoSchedulingOutcome(result, segments)
	if story.AutoSchedulingLocked && len(segments) > 0 {
		status = AutoSchedulingStatusLocked
		reason = "Maya retained the locked schedule without moving its time."
	}
	reason = refineScheduleOutcomeReason(previousBlocks, segments, status, reason)
	transition := buildStoryScheduleTransition(story, ownerID, previousBlocks, segments, result.Timezone, status, reason)
	return s.stories.UpdateAutomationStateIfUnchanged(
		ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
		status, &reason, nil, transition,
	)
}

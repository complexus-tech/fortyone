package maya

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/workschedule"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func (s *Service) ReconcileSchedule(ctx context.Context, input ReconcileScheduleInput) error {
	if err := s.validate(); err != nil {
		return err
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return err
	}
	refs := []ScheduleStoryRef{}
	if input.UserID == nil && input.StoryID != nil && *input.StoryID != uuid.Nil {
		if input.WorkspaceID == nil || *input.WorkspaceID == uuid.Nil {
			return fmt.Errorf("%w: story reconciliation requires a workspace", ErrInvalidPlanInput)
		}
		story, storyErr := s.stories.Get(ctx, *input.StoryID, *input.WorkspaceID)
		if storyErr != nil && !errors.Is(storyErr, ErrStoryNotFound) {
			return storyErr
		}
		if storyErr == nil && story.AutoSchedulingEnabled && story.Assignee != nil && *story.Assignee != uuid.Nil && *story.Assignee != s.mayaActorID {
			// A story update can change the ordering of the owner's entire
			// future workload. Reconcile the owner as a batch so a newly higher
			// priority story can displace movable lower-priority blocks.
			input.UserID = story.Assignee
		}
	}
	if input.UserID != nil && *input.UserID != uuid.Nil {
		userRefs, err := scheduleRepo.ListScheduleStoryRefsForUser(ctx, *input.UserID)
		if err != nil {
			return err
		}
		refs = append(refs, userRefs...)
	}
	if input.StoryID != nil && *input.StoryID != uuid.Nil {
		if input.WorkspaceID == nil || *input.WorkspaceID == uuid.Nil {
			return fmt.Errorf("%w: story reconciliation requires a workspace", ErrInvalidPlanInput)
		}
		refs = append(refs, ScheduleStoryRef{WorkspaceID: *input.WorkspaceID, StoryID: *input.StoryID})
	}

	return s.reconcileScheduleRefs(ctx, refs, true, s.clock.Now().UTC())
}

// RecoverScheduleOwnerships periodically replays the durable ownership set so
// schedules converge even when a worker dies between schedule commit and story
// assignment, or when a settings/lifecycle change has no dedicated event.
func (s *Service) RecoverScheduleOwnerships(ctx context.Context, limit int) (int, error) {
	if err := s.validate(); err != nil {
		return 0, err
	}
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return 0, err
	}
	now := s.clock.Now().UTC()
	refs, err := scheduleRepo.ClaimScheduleRecoveryStoryRefs(
		ctx,
		limit,
		now.Add(-scheduleRecoveryRetryDelay),
		now.Add(-interruptedScheduleRunStaleness),
	)
	if err != nil {
		return 0, err
	}
	var recoveryErr error
	for _, recovery := range refs {
		if err := s.reconcileScheduleRefs(ctx, []ScheduleStoryRef{recovery.ScheduleStoryRef}, false, now); err != nil {
			recoveryErr = errors.Join(recoveryErr, err)
			continue
		}
		if recovery.InterruptedRunID != nil {
			const message = "The original Maya operation was interrupted; recovery reconciled the schedule to the story's current state without changing its assignee."
			if err := scheduleRepo.CompleteInterruptedScheduleRun(ctx, *recovery.InterruptedRunID, message); err != nil {
				recoveryErr = errors.Join(recoveryErr, err)
			}
		}
	}
	return len(refs), recoveryErr
}

func (s *Service) reconcileScheduleRefs(ctx context.Context, refs []ScheduleStoryRef, recordNoop bool, asOf time.Time) error {
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return err
	}
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return err
	}

	seen := make(map[ScheduleStoryRef]struct{}, len(refs))
	affectedUsers := make(map[uuid.UUID]struct{})
	var reconciliationErr error
	for _, ref := range refs {
		if ref.WorkspaceID == uuid.Nil || ref.StoryID == uuid.Nil {
			continue
		}
		if _, exists := seen[ref]; exists {
			continue
		}
		seen[ref] = struct{}{}
		var users []uuid.UUID
		err := scheduleRepo.WithScheduleStoryLock(ctx, ref.WorkspaceID, ref.StoryID, func() error {
			var reconcileErr error
			users, reconcileErr = s.reconcileStorySchedule(ctx, ref, recordNoop, asOf)
			return reconcileErr
		})
		for _, userID := range users {
			affectedUsers[userID] = struct{}{}
		}
		if err != nil {
			reconciliationErr = errors.Join(reconciliationErr, err)
		}
	}
	for userID := range affectedUsers {
		if err := scheduleCalendar.DispatchScheduleEventOutbox(ctx, userID); err != nil {
			reconciliationErr = errors.Join(reconciliationErr, err)
		}
	}
	return reconciliationErr
}

func (s *Service) reconcileStorySchedule(ctx context.Context, ref ScheduleStoryRef, recordNoop bool, asOf time.Time) ([]uuid.UUID, error) {
	scheduleRepo, err := s.scheduleRepository()
	if err != nil {
		return nil, err
	}
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return nil, err
	}
	owners, err := scheduleRepo.ListMayaScheduleOwners(ctx, ref.WorkspaceID, ref.StoryID)
	if err != nil {
		return nil, err
	}
	affectedUsers := append([]uuid.UUID(nil), owners...)
	story, err := s.stories.Get(ctx, ref.StoryID, ref.WorkspaceID)
	if err != nil {
		if !errors.Is(err, ErrStoryNotFound) {
			return affectedUsers, err
		}
		for _, ownerID := range owners {
			if _, cleanupErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
				WorkspaceID: ref.WorkspaceID, UserID: ownerID, StoryID: ref.StoryID,
			}); cleanupErr != nil {
				return affectedUsers, errors.Join(err, cleanupErr)
			}
		}
		return affectedUsers, nil
	}
	hasAccess, err := scheduleRepo.WorkspaceCanUseMaya(ctx, ref.WorkspaceID)
	if err != nil {
		return affectedUsers, err
	}
	active, err := scheduleRepo.StoryIsActiveForAutoScheduling(ctx, ref.WorkspaceID, ref.StoryID)
	if err != nil {
		return affectedUsers, err
	}
	if !story.AutoSchedulingEnabled || !hasAccess || !active {
		retiredBlocks := make([]ScheduleBlock, 0)
		for _, ownerID := range owners {
			blocks, listErr := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, ownerID, ref.StoryID)
			if listErr != nil {
				return affectedUsers, listErr
			}
			retiredBlocks = append(retiredBlocks, blocks...)
		}
		if story.AutoSchedulingEnabled {
			reason := "Auto-scheduling is paused while this story is complete, cancelled, archived, deleted, or unavailable to Maya."
			if !hasAccess {
				reason = "Auto-scheduling is paused because this workspace does not currently have Maya access."
			}
			var transition *events.StoryScheduleTransition
			if len(owners) > 0 {
				transition = buildStoryScheduleTransition(
					story, owners[0], retiredBlocks, nil, "UTC",
					AutoSchedulingStatusOff, reason,
				)
			}
			var locked *bool
			if story.AutoSchedulingLocked {
				unlocked := false
				locked = &unlocked
			}
			if err := s.stories.UpdateAutomationStateIfUnchanged(
				ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
				AutoSchedulingStatusOff, &reason, locked, transition,
			); err != nil {
				return affectedUsers, err
			}
		}
		for _, ownerID := range owners {
			if _, cleanupErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
				WorkspaceID: ref.WorkspaceID, UserID: ownerID, StoryID: ref.StoryID,
				ExpectedStoryUpdatedAt: &story.UpdatedAt,
			}); cleanupErr != nil {
				return affectedUsers, cleanupErr
			}
		}
		return affectedUsers, nil
	}
	if story.AutoSchedulingLocked {
		hasLockedBlock := false
		for _, ownerID := range owners {
			blocks, listErr := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, ownerID, ref.StoryID)
			if listErr != nil {
				return affectedUsers, listErr
			}
			if len(blocks) > 0 {
				hasLockedBlock = true
				break
			}
		}
		if !hasLockedBlock {
			status, reason := unlockedAutoSchedulingState(story, s.mayaActorID)
			unlocked := false
			if err := s.stories.UpdateAutomationStateIfUnchanged(
				ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
				status, &reason, &unlocked, nil,
			); err != nil {
				return affectedUsers, err
			}
			story.AutoSchedulingLocked = false
			story.AutoSchedulingStatus = status
			story.AutoSchedulingReason = &reason
		}
	}

	if story.Assignee != nil && *story.Assignee == s.mayaActorID {
		windowStart := asOf
		plan, planErr := s.createWorkPlan(ctx, CreateWorkPlanInput{
			WorkspaceID: story.Workspace,
			StoryID:     story.ID,
			TriggeredBy: s.mayaActorID,
			Trigger:     RunTriggerEvent,
			WindowStart: windowStart,
			WindowEnd:   windowStart.Add(14 * 24 * time.Hour),
			AutoApply:   true,
		}, asOf)
		if planErr != nil {
			return affectedUsers, planErr
		}
		for _, action := range plan.Actions {
			if action.Payload.AssignStory != nil {
				affectedUsers = appendUniqueUUID(affectedUsers, action.Payload.AssignStory.AssigneeID)
			}
		}
		return affectedUsers, nil
	}

	var run CoreRun
	ensureRun := func() error {
		if run.ID != uuid.Nil {
			return nil
		}
		contextPayload, _ := json.Marshal(map[string]any{
			"reason":    "calendar availability or story scheduling constraints changed",
			"automatic": true,
		})
		created, createErr := s.repo.CreateRun(ctx, CreateRunInput{
			WorkspaceID: ref.WorkspaceID, StoryID: ref.StoryID, TriggeredBy: s.mayaActorID,
			Trigger: RunTriggerEvent, Context: contextPayload,
		})
		if createErr != nil {
			return fmt.Errorf("create automatic Maya schedule run: %w", createErr)
		}
		run = created
		return nil
	}
	failRun := func(runErr error) error {
		if ensureErr := ensureRun(); ensureErr != nil {
			return errors.Join(runErr, ensureErr)
		}
		return s.failScheduleRun(ctx, run, runErr)
	}

	desiredOwner := uuid.Nil
	if story.Assignee != nil && *story.Assignee != uuid.Nil && *story.Assignee != s.mayaActorID {
		schedulable, eligibilityErr := scheduleRepo.StoryIsSchedulableForUser(ctx, ref.WorkspaceID, ref.StoryID, *story.Assignee)
		if eligibilityErr != nil {
			return affectedUsers, failRun(eligibilityErr)
		}
		if schedulable {
			desiredOwner = *story.Assignee
			affectedUsers = appendUniqueUUID(affectedUsers, desiredOwner)
		}
	}
	if desiredOwner == uuid.Nil {
		reason := "Choose an active teammate with access to this story so Maya can schedule it."
		var locked *bool
		if story.AutoSchedulingLocked {
			unlocked := false
			locked = &unlocked
		}
		if err := s.stories.UpdateAutomationStateIfUnchanged(
			ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
			AutoSchedulingStatusNeedsOwner, &reason, locked, nil,
		); err != nil {
			return affectedUsers, failRun(err)
		}
		for _, ownerID := range owners {
			if _, cleanupErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
				WorkspaceID: ref.WorkspaceID, UserID: ownerID, StoryID: ref.StoryID,
				ExpectedStoryUpdatedAt: &story.UpdatedAt,
			}); cleanupErr != nil {
				return affectedUsers, failRun(cleanupErr)
			}
		}
		return affectedUsers, nil
	}

	actions := []CoreAction{}
	summary := "Maya removed scheduled work because the story is no longer eligible for automatic scheduling."
	desiredSegments := []ScheduleSegmentInput{}
	previousBlocks := make([]ScheduleBlock, 0)
	preemptedBlockIDs := []uuid.UUID(nil)
	blocksByOwner := make(map[uuid.UUID][]ScheduleBlock, len(owners)+1)
	for _, ownerID := range owners {
		blocks, listErr := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, ownerID, ref.StoryID)
		if listErr != nil {
			return affectedUsers, failRun(listErr)
		}
		blocksByOwner[ownerID] = blocks
		previousBlocks = append(previousBlocks, blocks...)
	}
	if !uuidSliceContains(owners, desiredOwner) {
		blocks, listErr := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, desiredOwner, ref.StoryID)
		if listErr != nil {
			return affectedUsers, failRun(listErr)
		}
		blocksByOwner[desiredOwner] = blocks
		previousBlocks = append(previousBlocks, blocks...)
	}
	planTimezone := "UTC"
	outcomeStatus := AutoSchedulingStatusScheduled
	outcomeReason := "Maya scheduled this story around the assignee's availability."
	if story.AutoSchedulingLocked {
		lockedBlocks := append([]ScheduleBlock(nil), blocksByOwner[desiredOwner]...)
		if len(lockedBlocks) == 0 {
			for _, ownerID := range owners {
				if len(blocksByOwner[ownerID]) > 0 {
					lockedBlocks = append(lockedBlocks, blocksByOwner[ownerID]...)
					break
				}
			}
		}
		sort.SliceStable(lockedBlocks, func(i, j int) bool {
			if lockedBlocks[i].SegmentIndex != lockedBlocks[j].SegmentIndex {
				return lockedBlocks[i].SegmentIndex < lockedBlocks[j].SegmentIndex
			}
			return lockedBlocks[i].StartAt.Before(lockedBlocks[j].StartAt)
		})
		if len(lockedBlocks) == 0 {
			outcomeStatus = AutoSchedulingStatusAtRisk
			outcomeReason = "The story is marked locked, but there is no Maya schedule to retain."
		} else {
			timezoneView, timezoneErr := scheduleCalendar.ListSchedulingAvailability(
				ctx,
				story.Workspace,
				desiredOwner,
				lockedBlocks[0].StartAt,
				lockedBlocks[len(lockedBlocks)-1].EndAt,
			)
			if timezoneErr != nil {
				return affectedUsers, failRun(timezoneErr)
			}
			planTimezone = timezoneView.Timezone
			summary = "Maya retained the locked schedule without moving its time."
			for index, block := range lockedBlocks {
				desiredSegments = append(desiredSegments, ScheduleSegmentInput{
					SegmentIndex: index,
					Title:        story.Title,
					StartAt:      block.StartAt,
					EndAt:        block.EndAt,
				})
				actions = append(actions, CoreAction{
					WorkspaceID: ref.WorkspaceID, StoryID: ref.StoryID,
					Type: ActionTypeScheduleWorkBlock, Status: ActionStatusProposed,
					Reason: "The user locked this Maya schedule, so its current time is retained.",
					Payload: ActionPayload{ScheduleBlock: &ScheduleBlockPayload{
						UserID: desiredOwner, SegmentIndex: index, Title: story.Title,
						StartAt: block.StartAt, EndAt: block.EndAt, ExpectedStoryUpdatedAt: story.UpdatedAt,
						Operation: ScheduleBlockOperationUpsert,
					}},
				})
			}
			outcomeStatus = AutoSchedulingStatusLocked
			outcomeReason = "Maya retained the locked schedule without moving its time."
			if riskReason, atRisk := lockedScheduleRisk(story, lockedBlocks, planTimezone, asOf); atRisk {
				outcomeStatus = AutoSchedulingStatusAtRisk
				outcomeReason = riskReason
			}
		}
	} else {
		planResult, planErr := s.planAssignedStory(ctx, story, desiredOwner, asOf)
		if planErr != nil {
			return affectedUsers, failRun(planErr)
		}
		planTimezone = planResult.Timezone
		summary = planResult.Summary
		preemptedBlockIDs = append([]uuid.UUID(nil), planResult.PreemptedBlockIDs...)
		actions = append(actions, planResult.Actions...)
		for _, action := range planResult.Actions {
			if action.Type != ActionTypeScheduleWorkBlock || action.Payload.ScheduleBlock == nil || action.Payload.ScheduleBlock.Operation == ScheduleBlockOperationRetain {
				continue
			}
			action.Payload.ScheduleBlock.Operation = ScheduleBlockOperationUpsert
			desiredSegments = append(desiredSegments, ScheduleSegmentInput{
				SegmentIndex: action.Payload.ScheduleBlock.SegmentIndex,
				Title:        action.Payload.ScheduleBlock.Title,
				StartAt:      action.Payload.ScheduleBlock.StartAt,
				EndAt:        action.Payload.ScheduleBlock.EndAt,
			})
		}
		if len(desiredSegments) == 0 {
			existingBlocks, listErr := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, desiredOwner, ref.StoryID)
			if listErr != nil {
				return affectedUsers, failRun(listErr)
			}
			if len(existingBlocks) > 0 {
				actions = append(actions, CoreAction{
					WorkspaceID: ref.WorkspaceID, StoryID: ref.StoryID,
					Type: ActionTypeScheduleWorkBlock, Status: ActionStatusProposed,
					Reason: "Maya removed the generated segments because the current scheduling constraints do not permit a safe placement.",
					Payload: ActionPayload{ScheduleBlock: &ScheduleBlockPayload{
						UserID: desiredOwner, Operation: ScheduleBlockOperationDelete,
					}},
				})
			}
		}
		outcomeStatus, outcomeReason = autoSchedulingOutcome(planResult, desiredSegments)
	}
	outcomeReason = refineScheduleOutcomeReason(previousBlocks, desiredSegments, outcomeStatus, outcomeReason)
	if !recordNoop {
		unchanged, stateErr := s.recoveryScheduleStateMatches(ctx, scheduleCalendar, ref, owners, desiredOwner, desiredSegments, story.AutoSchedulingLocked)
		if stateErr != nil {
			return affectedUsers, stateErr
		}
		if unchanged {
			_, stateErr = scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
				WorkspaceID: ref.WorkspaceID, UserID: desiredOwner, StoryID: ref.StoryID,
				ExpectedStoryUpdatedAt: &story.UpdatedAt, Segments: desiredSegments, KeepOwnership: true,
				Locked: story.AutoSchedulingLocked,
			})
			if stateErr == nil {
				transition := buildStoryScheduleTransition(story, desiredOwner, previousBlocks, desiredSegments, planTimezone, outcomeStatus, outcomeReason)
				stateErr = s.stories.UpdateAutomationStateIfUnchanged(
					ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
					outcomeStatus, &outcomeReason, nil, transition,
				)
			}
			return affectedUsers, stateErr
		}
	}
	if err := ensureRun(); err != nil {
		return affectedUsers, err
	}
	for _, ownerID := range owners {
		if ownerID == desiredOwner {
			continue
		}
		actions = append(actions, CoreAction{
			WorkspaceID: ref.WorkspaceID, StoryID: ref.StoryID,
			Type: ActionTypeScheduleWorkBlock, Status: ActionStatusProposed,
			Reason:  "Maya removed this user's generated schedule because the story owner or status changed.",
			Payload: ActionPayload{ScheduleBlock: &ScheduleBlockPayload{UserID: ownerID, Operation: ScheduleBlockOperationDelete}},
		})
	}
	for index := range actions {
		actions[index].RunID = run.ID
		actions[index].WorkspaceID = ref.WorkspaceID
		actions[index].StoryID = ref.StoryID
	}
	persistedActions, err := s.repo.CreateActions(ctx, actions)
	if err != nil {
		return affectedUsers, s.failScheduleRun(ctx, run, err)
	}

	if desiredOwner != uuid.Nil {
		if _, err := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
			WorkspaceID: ref.WorkspaceID, UserID: desiredOwner, StoryID: ref.StoryID,
			ExpectedStoryUpdatedAt: &story.UpdatedAt, Segments: desiredSegments, KeepOwnership: true,
			PreemptBlockIDs: preemptedBlockIDs, Locked: story.AutoSchedulingLocked,
		}); err != nil {
			if markErr := s.markScheduleActionsFailed(ctx, persistedActions, err); markErr != nil {
				// Leave the run open so the interrupted-run recovery transaction can
				// terminalize the actions and run together once persistence recovers.
				return affectedUsers, errors.Join(err, markErr)
			}
			return affectedUsers, s.failScheduleRun(ctx, run, err)
		}
	}
	for _, ownerID := range owners {
		if ownerID == desiredOwner {
			continue
		}
		if _, err := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, ScheduleReconcileInput{
			WorkspaceID: ref.WorkspaceID, UserID: ownerID, StoryID: ref.StoryID,
			ExpectedStoryUpdatedAt: &story.UpdatedAt,
		}); err != nil {
			if markErr := s.markScheduleActionsFailed(ctx, persistedActions, err); markErr != nil {
				// Leave the run open so the interrupted-run recovery transaction can
				// terminalize the actions and run together once persistence recovers.
				return affectedUsers, errors.Join(err, markErr)
			}
			return affectedUsers, s.failScheduleRun(ctx, run, err)
		}
	}
	appliedAt := s.clock.Now().UTC()
	var markErr error
	for index := range persistedActions {
		markErr = errors.Join(markErr, s.markActionApplied(ctx, &persistedActions[index], appliedAt))
	}
	if markErr != nil {
		// Calendar mutations are already durable. Do not report or compensate them
		// as rolled back; leave the run open for interrupted-run recovery.
		return affectedUsers, markErr
	}
	if _, err := s.repo.CompleteRun(ctx, run.ID, RunStatusSucceeded, summary, nil); err != nil {
		return affectedUsers, err
	}
	transition := buildStoryScheduleTransition(story, desiredOwner, previousBlocks, desiredSegments, planTimezone, outcomeStatus, outcomeReason)
	if err := s.stories.UpdateAutomationStateIfUnchanged(
		ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
		outcomeStatus, &outcomeReason, nil, transition,
	); err != nil {
		return affectedUsers, err
	}
	return affectedUsers, nil
}

func unlockedAutoSchedulingState(story Story, mayaActorID uuid.UUID) (string, string) {
	if story.Assignee == nil || *story.Assignee == uuid.Nil {
		return AutoSchedulingStatusNeedsOwner, "Choose an assignee so Maya can schedule this story."
	}
	if *story.Assignee == mayaActorID {
		return AutoSchedulingStatusNeedsOwner, "Maya is selecting the best available teammate."
	}
	if story.EstimatedDurationMinutes == nil {
		return AutoSchedulingStatusNeedsTime, "Add time needed so Maya can reserve focused work."
	}
	return AutoSchedulingStatusPlanning, "Maya is checking availability and scheduling this story."
}

func lockedScheduleRisk(story Story, blocks []ScheduleBlock, timezone string, asOf time.Time) (string, bool) {
	for _, block := range blocks {
		if block.HasConflict {
			return "The locked time now conflicts with another calendar event. Unlock it so Maya can move the work.", true
		}
	}
	allElapsed := len(blocks) > 0
	for _, block := range blocks {
		if block.EndAt.After(asOf) {
			allElapsed = false
			break
		}
	}
	if allElapsed {
		return "The locked work time has passed, but this story is still incomplete. Unlock it so Maya can schedule the remaining work.", true
	}
	if story.EstimatedDurationMinutes == nil || *story.EstimatedDurationMinutes <= 0 {
		return "This locked schedule no longer has a valid time needed. Add time needed and unlock it so Maya can rebuild the work blocks.", true
	}

	expectedMinutes := *story.EstimatedDurationMinutes
	reservedMinutes := 0
	minimumFocusMinutes := 0
	if story.MinimumFocusBlockMinutes != nil && *story.MinimumFocusBlockMinutes > 0 {
		minimumFocusMinutes = *story.MinimumFocusBlockMinutes
	}
	if minimumFocusMinutes > expectedMinutes {
		minimumFocusMinutes = expectedMinutes
	}
	for _, block := range blocks {
		blockMinutes := int(block.EndAt.Sub(block.StartAt) / time.Minute)
		if minimumFocusMinutes > 0 && blockMinutes < minimumFocusMinutes {
			return fmt.Sprintf(
				"A locked work block is shorter than the %d-minute minimum focus block. Unlock it so Maya can rebuild the schedule.",
				minimumFocusMinutes,
			), true
		}
		reservedMinutes += blockMinutes
	}
	if reservedMinutes != expectedMinutes {
		return fmt.Sprintf(
			"The locked schedule reserves %d minutes, but this story now needs %d minutes. Unlock it so Maya can rebuild the work blocks.",
			reservedMinutes,
			expectedMinutes,
		), true
	}

	location := calendarLocation(timezone)
	for _, block := range blocks {
		if story.StartDate != nil && block.StartAt.Before(story.StartDate.UTC()) {
			return "A locked work block is before this story's start date. Unlock it so Maya can move the work into the valid window.", true
		}
		if story.EndDate != nil && block.EndAt.After(story.EndDate.UTC().Add(24*time.Hour)) {
			return "A locked work block is after this story's deadline. Unlock it so Maya can move the work into the valid window.", true
		}
		if story.SprintSummary != nil {
			defaultSchedule := workschedule.Default()
			sprintStart := workdayBoundary(story.SprintSummary.StartDate, defaultSchedule.StartMinute, location)
			sprintEnd := workdayBoundary(story.SprintSummary.EndDate, defaultSchedule.EndMinute, location)
			if block.StartAt.Before(sprintStart) || block.EndAt.After(sprintEnd) {
				return "A locked work block falls outside this story's sprint. Unlock it so Maya can move the work into the sprint window.", true
			}
		}
	}
	return "", false
}

package maya

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

type ReconcileScheduleInput struct {
	WorkspaceID *uuid.UUID
	UserID      *uuid.UUID
	StoryID     *uuid.UUID
}

const (
	scheduleRecoveryRetryDelay      = 5 * time.Minute
	interruptedScheduleRunStaleness = 10 * time.Minute
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
		if storyErr != nil && !errors.Is(storyErr, stories.ErrNotFound) {
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

	return s.reconcileScheduleRefs(ctx, refs, true)
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
	now := time.Now().UTC()
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
		if err := s.reconcileScheduleRefs(ctx, []ScheduleStoryRef{recovery.ScheduleStoryRef}, false); err != nil {
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

func (s *Service) reconcileScheduleRefs(ctx context.Context, refs []ScheduleStoryRef, recordNoop bool) error {
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
			users, reconcileErr = s.reconcileStorySchedule(ctx, ref, recordNoop)
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

func (s *Service) reconcileStorySchedule(ctx context.Context, ref ScheduleStoryRef, recordNoop bool) ([]uuid.UUID, error) {
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
		if !errors.Is(err, stories.ErrNotFound) {
			return affectedUsers, err
		}
		for _, ownerID := range owners {
			if _, cleanupErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
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
		retiredBlocks := make([]calendar.CoreScheduleBlock, 0)
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
					stories.AutoSchedulingStatusOff, reason,
				)
			}
			var locked *bool
			if story.AutoSchedulingLocked {
				unlocked := false
				locked = &unlocked
			}
			if err := s.stories.UpdateAutomationStateIfUnchanged(
				ctx, s.mayaActorID, story.ID, story.Workspace, story.UpdatedAt,
				stories.AutoSchedulingStatusOff, &reason, locked, transition,
			); err != nil {
				return affectedUsers, err
			}
		}
		for _, ownerID := range owners {
			if _, cleanupErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
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
		windowStart := time.Now().UTC()
		plan, planErr := s.createWorkPlan(ctx, CreateWorkPlanInput{
			WorkspaceID: story.Workspace,
			StoryID:     story.ID,
			TriggeredBy: s.mayaActorID,
			Trigger:     RunTriggerEvent,
			WindowStart: windowStart,
			WindowEnd:   windowStart.Add(14 * 24 * time.Hour),
			AutoApply:   true,
		})
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
			stories.AutoSchedulingStatusNeedsOwner, &reason, locked, nil,
		); err != nil {
			return affectedUsers, failRun(err)
		}
		for _, ownerID := range owners {
			if _, cleanupErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
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
	desiredSegments := []calendar.MayaScheduleSegmentInput{}
	previousBlocks := make([]calendar.CoreScheduleBlock, 0)
	preemptedBlockIDs := []uuid.UUID(nil)
	blocksByOwner := make(map[uuid.UUID][]calendar.CoreScheduleBlock, len(owners)+1)
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
	outcomeStatus := stories.AutoSchedulingStatusScheduled
	outcomeReason := "Maya scheduled this story around the assignee's availability."
	if story.AutoSchedulingLocked {
		lockedBlocks := append([]calendar.CoreScheduleBlock(nil), blocksByOwner[desiredOwner]...)
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
			outcomeStatus = stories.AutoSchedulingStatusAtRisk
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
				desiredSegments = append(desiredSegments, calendar.MayaScheduleSegmentInput{
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
			outcomeStatus = stories.AutoSchedulingStatusLocked
			outcomeReason = "Maya retained the locked schedule without moving its time."
			if riskReason, atRisk := lockedScheduleRisk(story, lockedBlocks, planTimezone); atRisk {
				outcomeStatus = stories.AutoSchedulingStatusAtRisk
				outcomeReason = riskReason
			}
		}
	} else {
		planResult, planErr := s.planAssignedStory(ctx, story, desiredOwner)
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
			desiredSegments = append(desiredSegments, calendar.MayaScheduleSegmentInput{
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
			_, stateErr = scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
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
		if _, err := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
			WorkspaceID: ref.WorkspaceID, UserID: desiredOwner, StoryID: ref.StoryID,
			ExpectedStoryUpdatedAt: &story.UpdatedAt, Segments: desiredSegments, KeepOwnership: true,
			PreemptBlockIDs: preemptedBlockIDs, Locked: story.AutoSchedulingLocked,
		}); err != nil {
			s.markScheduleActionsFailed(ctx, persistedActions, err)
			return affectedUsers, s.failScheduleRun(ctx, run, err)
		}
	}
	for _, ownerID := range owners {
		if ownerID == desiredOwner {
			continue
		}
		if _, err := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
			WorkspaceID: ref.WorkspaceID, UserID: ownerID, StoryID: ref.StoryID,
			ExpectedStoryUpdatedAt: &story.UpdatedAt,
		}); err != nil {
			s.markScheduleActionsFailed(ctx, persistedActions, err)
			return affectedUsers, s.failScheduleRun(ctx, run, err)
		}
	}
	for _, action := range persistedActions {
		_ = s.repo.MarkActionApplied(ctx, action.ID)
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

func unlockedAutoSchedulingState(story stories.CoreSingleStory, mayaActorID uuid.UUID) (string, string) {
	if story.Assignee == nil || *story.Assignee == uuid.Nil {
		return stories.AutoSchedulingStatusNeedsOwner, "Choose an assignee so Maya can schedule this story."
	}
	if *story.Assignee == mayaActorID {
		return stories.AutoSchedulingStatusNeedsOwner, "Maya is selecting the best available teammate."
	}
	if story.EstimatedDurationMinutes == nil {
		return stories.AutoSchedulingStatusNeedsTime, "Add time needed so Maya can reserve focused work."
	}
	return stories.AutoSchedulingStatusPlanning, "Maya is checking availability and scheduling this story."
}

func lockedScheduleRisk(story stories.CoreSingleStory, blocks []calendar.CoreScheduleBlock, timezone string) (string, bool) {
	for _, block := range blocks {
		if block.HasConflict {
			return "The locked time now conflicts with another calendar event. Unlock it so Maya can move the work.", true
		}
	}
	allElapsed := len(blocks) > 0
	now := time.Now().UTC()
	for _, block := range blocks {
		if block.EndAt.After(now) {
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
			sprintStart := sprintWorkdayStart(story.SprintSummary.StartDate, location)
			sprintEnd := sprintWorkdayEnd(story.SprintSummary.EndDate, location)
			if block.StartAt.Before(sprintStart) || block.EndAt.After(sprintEnd) {
				return "A locked work block falls outside this story's sprint. Unlock it so Maya can move the work into the sprint window.", true
			}
		}
	}
	return "", false
}

func (s *Service) recoveryScheduleStateMatches(
	ctx context.Context,
	scheduleCalendar ScheduleCalendarService,
	ref ScheduleStoryRef,
	owners []uuid.UUID,
	desiredOwner uuid.UUID,
	desiredSegments []calendar.MayaScheduleSegmentInput,
	locked bool,
) (bool, error) {
	if desiredOwner != uuid.Nil {
		if len(owners) != 1 || owners[0] != desiredOwner {
			return false, nil
		}
		blocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, desiredOwner, ref.StoryID)
		if err != nil {
			return false, err
		}
		return mayaScheduleSegmentsMatchBlocks(desiredSegments, blocks, locked), nil
	}
	for _, ownerID := range owners {
		blocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, ownerID, ref.StoryID)
		if err != nil {
			return false, err
		}
		if len(blocks) > 0 {
			return false, nil
		}
	}
	return true, nil
}

func mayaScheduleSegmentsMatchBlocks(segments []calendar.MayaScheduleSegmentInput, blocks []calendar.CoreScheduleBlock, locked bool) bool {
	if len(segments) != len(blocks) {
		return false
	}
	blocksByIndex := make(map[int]calendar.CoreScheduleBlock, len(blocks))
	for _, block := range blocks {
		blocksByIndex[block.SegmentIndex] = block
	}
	for _, segment := range segments {
		block, ok := blocksByIndex[segment.SegmentIndex]
		if !ok || block.Title != segment.Title || !block.StartAt.Equal(segment.StartAt) || !block.EndAt.Equal(segment.EndAt) || block.IsLocked != locked {
			return false
		}
	}
	return true
}

func (s *Service) planAssignedStory(ctx context.Context, story stories.CoreSingleStory, userID uuid.UUID) (PlanResult, error) {
	windowStart := time.Now().UTC()
	if story.StartDate != nil && story.StartDate.After(windowStart) {
		windowStart = story.StartDate.UTC()
	}
	windowEnd := windowStart.Add(90 * 24 * time.Hour)
	if story.EndDate != nil {
		deadlineEnd := story.EndDate.UTC().Add(24 * time.Hour)
		if !deadlineEnd.After(windowStart) {
			return PlanResult{
				Summary:        "Maya could not place this work because its deadline has already passed.",
				SelectedUserID: &userID,
				Actions: []CoreAction{
					scheduleOwnershipRetentionAction(story.Workspace, story, userID, "Maya will keep watching this overdue work for a changed deadline."),
					{
						WorkspaceID: story.Workspace,
						StoryID:     story.ID,
						Type:        ActionTypeFlagScheduleRisk,
						Status:      ActionStatusProposed,
						Reason:      "The story deadline passed before enough work time could be reserved.",
						Payload: ActionPayload{Risk: &RiskPayload{
							Code:    "deadline_passed",
							Message: "Move the deadline or adjust the time needed before Maya can place this work.",
						}},
					},
				},
			}, nil
		}
		if deadlineEnd.Before(windowEnd) {
			windowEnd = deadlineEnd
		}
	}
	scheduleCalendar, err := s.scheduleCalendarService()
	if err != nil {
		return PlanResult{}, err
	}
	schedule, err := scheduleCalendar.ListSchedulingAvailability(ctx, story.Workspace, userID, windowStart, windowEnd)
	if err != nil {
		return PlanResult{}, err
	}
	blocks := make([]calendar.CoreScheduleBlock, 0, len(schedule.Blocks))
	var activeBlock *calendar.CoreScheduleBlock
	elapsedMinutes := 0
	hasUnfinishedScheduledTime := false
	now := time.Now().UTC()
	for _, block := range schedule.Blocks {
		if block.WorkspaceID == story.Workspace && block.StoryID != nil && *block.StoryID == story.ID && block.Source == calendar.ScheduleBlockSourceMaya {
			elapsedMinutes += elapsedScheduleMinutes(block, now)
			hasUnfinishedScheduledTime = hasUnfinishedScheduledTime || block.EndAt.After(now)
			if !block.IsLocked && block.StartAt.Before(now) && block.EndAt.After(now) {
				current := block
				activeBlock = &current
			}
			continue
		}
		blocks = append(blocks, block)
	}
	if activeBlock != nil {
		// Keep the in-progress block as an anonymous occupied interval. The
		// planner must not move it, while the reconciliation layer will restore
		// its story metadata and segment index in the desired schedule below.
		current := *activeBlock
		current.StoryID = nil
		current.StoryTitle = nil
		current.StoryCode = nil
		blocks = append(blocks, current)
	}
	workingDays, err := s.getWorkingDays(ctx, story.Team, story.Workspace)
	if err != nil {
		return PlanResult{}, err
	}
	if !hasUnfinishedScheduledTime {
		// A fully elapsed unlocked schedule is treated as unfinished work by
		// the existing recovery contract; reserve the full estimate again.
		elapsedMinutes = 0
	}
	durationMinutes := effectiveRemainingDurationMinutes(story, elapsedMinutes)
	candidate := CandidateSchedule{
		Member: reports.CoreMemberWorkload{UserID: userID}, Timezone: schedule.Timezone,
		BusyWindows: schedule.BusyWindows, Blocks: blocks,
	}
	if feedbackService, ok := s.calendar.(ScheduleFeedbackService); ok {
		preference, preferenceErr := feedbackService.ListManualSchedulePreference(ctx, story.Workspace, userID)
		if preferenceErr != nil {
			return PlanResult{}, fmt.Errorf("list calendar schedule preference for %s: %w", userID, preferenceErr)
		}
		if preference.Confidence > 0 {
			candidate.PreferredStartMinute = preference.PreferredStartMinute
		}
	}
	result, err := s.planner.Plan(PlanInput{
		Context: ctx, WorkspaceID: story.Workspace, Story: story,
		DurationMinutes: durationMinutes,
		WindowStart:     windowStart, WindowEnd: windowEnd, WorkingDays: workingDays,
		MinimumFocusBlockMinutes: valueOrZero(story.MinimumFocusBlockMinutes),
		Candidates:               []CandidateSchedule{candidate},
	})
	if activeBlock != nil {
		result = retainActiveScheduleBlock(result, story, userID, *activeBlock)
	}
	result.Timezone = schedule.Timezone
	return result, err
}

func elapsedScheduleMinutes(block calendar.CoreScheduleBlock, now time.Time) int {
	if !block.StartAt.Before(now) {
		return 0
	}
	endAt := block.EndAt
	if endAt.After(now) {
		endAt = now
	}
	return max(0, int(endAt.Sub(block.StartAt)/time.Minute))
}

func effectiveRemainingDurationMinutes(story stories.CoreSingleStory, elapsedMinutes int) int {
	if story.EstimatedDurationMinutes == nil || *story.EstimatedDurationMinutes <= 0 {
		return 0
	}
	remaining := *story.EstimatedDurationMinutes - elapsedMinutes
	if remaining < 0 {
		return 0
	}
	return remaining
}

func retainActiveScheduleBlock(result PlanResult, story stories.CoreSingleStory, userID uuid.UUID, block calendar.CoreScheduleBlock) PlanResult {
	activeAction := CoreAction{
		WorkspaceID: story.Workspace,
		StoryID:     story.ID,
		Type:        ActionTypeScheduleWorkBlock,
		Status:      ActionStatusProposed,
		Reason:      "Maya retained the work already in progress while rebalancing future focus blocks.",
		Payload: ActionPayload{ScheduleBlock: &ScheduleBlockPayload{
			UserID: userID, SegmentIndex: 0, Title: story.Title,
			StartAt: block.StartAt, EndAt: block.EndAt,
			PlannedStartAt: block.StartAt, PlannedEndAt: block.EndAt,
			ExpectedStoryUpdatedAt: story.UpdatedAt,
		}},
	}
	shifted := make([]CoreAction, 0, len(result.Actions)+1)
	shifted = append(shifted, activeAction)
	for _, action := range result.Actions {
		if action.Payload.ScheduleBlock != nil && action.Type == ActionTypeScheduleWorkBlock && action.Payload.ScheduleBlock.Operation != ScheduleBlockOperationRetain {
			action.Payload.ScheduleBlock.SegmentIndex++
		}
		shifted = append(shifted, action)
	}
	result.Actions = shifted
	return result
}

func autoSchedulingOutcome(result PlanResult, segments []calendar.MayaScheduleSegmentInput) (string, string) {
	if len(segments) > 0 {
		return stories.AutoSchedulingStatusScheduled, "Maya scheduled this story around the assignee's availability."
	}
	for _, action := range result.Actions {
		if action.Payload.Risk == nil {
			continue
		}
		switch action.Payload.Risk.Code {
		case "missing_duration":
			return stories.AutoSchedulingStatusNeedsTime, "Add time needed so Maya can reserve focused work."
		case "deadline_passed":
			return stories.AutoSchedulingStatusAtRisk, "The deadline has passed before Maya could reserve enough focused time."
		case "no_available_slot":
			return stories.AutoSchedulingStatusCannotFit, "Maya could not fit the required focus time into the current planning window."
		}
	}
	return stories.AutoSchedulingStatusCannotFit, "Maya could not place this work safely in the current planning window."
}

func refineScheduleOutcomeReason(previousBlocks []calendar.CoreScheduleBlock, segments []calendar.MayaScheduleSegmentInput, status, fallback string) string {
	if status != stories.AutoSchedulingStatusScheduled || len(previousBlocks) == 0 {
		return fallback
	}
	if scheduleSegmentsChanged(previousBlocks, segments) {
		return "The assignee's availability or this story's scheduling constraints changed, so Maya moved it to the next safe slot."
	}
	return fallback
}

func scheduleSegmentsChanged(previousBlocks []calendar.CoreScheduleBlock, segments []calendar.MayaScheduleSegmentInput) bool {
	if len(previousBlocks) != len(segments) {
		return true
	}
	previousBySegment := make(map[int]calendar.CoreScheduleBlock, len(previousBlocks))
	for _, block := range previousBlocks {
		previousBySegment[block.SegmentIndex] = block
	}
	for _, segment := range segments {
		previous, exists := previousBySegment[segment.SegmentIndex]
		if !exists || !previous.StartAt.Equal(segment.StartAt) || !previous.EndAt.Equal(segment.EndAt) {
			return true
		}
	}
	return false
}

func buildStoryScheduleTransition(
	story stories.CoreSingleStory,
	userID uuid.UUID,
	previousBlocks []calendar.CoreScheduleBlock,
	segments []calendar.MayaScheduleSegmentInput,
	timezone string,
	status string,
	reason string,
) *events.StoryScheduleTransition {
	if userID == uuid.Nil {
		return nil
	}
	for _, block := range previousBlocks {
		if block.UserID != userID {
			// Assignment/reassignment has its own reason-aware StoryUpdated event.
			// Suppress a second schedule-move notification for the same decision.
			return nil
		}
	}
	previousStart, previousEnd := scheduleBlockBounds(previousBlocks)
	startAt, endAt := scheduleSegmentBounds(segments)
	previousState := events.StoryScheduleState(story.AutoSchedulingStatus)
	state := events.StoryScheduleState(status)
	kind := events.StoryScheduleTransitionKind("")
	wasLocked := false
	for _, block := range previousBlocks {
		wasLocked = wasLocked || block.IsLocked
	}

	location := time.UTC
	if parsed, err := time.LoadLocation(timezone); err == nil {
		location = parsed
	} else {
		timezone = "UTC"
	}
	previousLocalDate := localDate(previousStart, location)
	localDateValue := localDate(startAt, location)
	shiftMinutes := 0
	if change, ok := selectMeaningfulScheduleChange(previousBlocks, segments, location); ok {
		kind = change.Kind
		previousStart = &change.PreviousStartAt
		previousEnd = &change.PreviousEndAt
		startAt = &change.StartAt
		endAt = &change.EndAt
		previousLocalDate = change.PreviousLocalDate
		localDateValue = change.LocalDate
		shiftMinutes = change.ShiftMinutes
	}
	if kind == "" && story.AutoSchedulingLocked && !wasLocked && len(segments) > 0 {
		kind = events.StoryScheduleTransitionLocked
	} else if kind == "" && !story.AutoSchedulingLocked && wasLocked {
		kind = events.StoryScheduleTransitionUnlocked
	}
	if kind == "" && len(segments) > 0 && len(previousBlocks) == 0 {
		kind = events.StoryScheduleTransitionFirstSchedule
	}
	if kind == "" && previousState != state {
		kind = events.StoryScheduleTransitionStateChanged
	}
	if kind == "" && equalScheduleReasonChanged(story.AutoSchedulingReason, reason) &&
		(status == stories.AutoSchedulingStatusNeedsTime || status == stories.AutoSchedulingStatusCannotFit || status == stories.AutoSchedulingStatusAtRisk) {
		kind = events.StoryScheduleTransitionStateChanged
	}
	if kind == "" {
		return nil
	}
	return &events.StoryScheduleTransition{
		Kind:              kind,
		UserID:            userID,
		PreviousState:     previousState,
		State:             state,
		PreviousStartAt:   previousStart,
		StartAt:           startAt,
		PreviousEndAt:     previousEnd,
		EndAt:             endAt,
		Timezone:          timezone,
		PreviousLocalDate: previousLocalDate,
		LocalDate:         localDateValue,
		ShiftMinutes:      shiftMinutes,
	}
}

type meaningfulScheduleChange struct {
	Kind              events.StoryScheduleTransitionKind
	PreviousStartAt   time.Time
	StartAt           time.Time
	PreviousEndAt     time.Time
	EndAt             time.Time
	PreviousLocalDate string
	LocalDate         string
	ShiftMinutes      int
}

func selectMeaningfulScheduleChange(
	previousBlocks []calendar.CoreScheduleBlock,
	segments []calendar.MayaScheduleSegmentInput,
	location *time.Location,
) (meaningfulScheduleChange, bool) {
	if scheduleTimesEqual(previousBlocks, segments) {
		return meaningfulScheduleChange{}, false
	}
	previousBySegment := make(map[int]calendar.CoreScheduleBlock, len(previousBlocks))
	for _, block := range previousBlocks {
		previousBySegment[block.SegmentIndex] = block
	}
	segmentsByIndex := make(map[int]calendar.MayaScheduleSegmentInput, len(segments))
	for _, segment := range segments {
		segmentsByIndex[segment.SegmentIndex] = segment
	}

	var selected meaningfulScheduleChange
	selectedSegmentIndex := 0
	selectedFound := false
	selectedDayChanged := false
	consider := func(previous calendar.CoreScheduleBlock, segment calendar.MayaScheduleSegmentInput) {
		previousLocalDate := previous.StartAt.In(location).Format(time.DateOnly)
		localDateValue := segment.StartAt.In(location).Format(time.DateOnly)
		startShiftMinutes := int(segment.StartAt.Sub(previous.StartAt).Minutes())
		endShiftMinutes := int(segment.EndAt.Sub(previous.EndAt).Minutes())
		dayChanged := previousLocalDate != localDateValue
		shiftMinutes := startShiftMinutes
		if !dayChanged && absoluteInt(startShiftMinutes) < 60 {
			if absoluteInt(endShiftMinutes) < 60 {
				return
			}
			shiftMinutes = endShiftMinutes
		}

		shouldSelect := !selectedFound ||
			(dayChanged && !selectedDayChanged) ||
			(dayChanged == selectedDayChanged && absoluteInt(shiftMinutes) > absoluteInt(selected.ShiftMinutes)) ||
			(dayChanged == selectedDayChanged && absoluteInt(shiftMinutes) == absoluteInt(selected.ShiftMinutes) && segment.SegmentIndex < selectedSegmentIndex)
		if !shouldSelect {
			return
		}

		kind := events.StoryScheduleTransitionMoved
		if dayChanged {
			kind = events.StoryScheduleTransitionDayChanged
		}
		selected = meaningfulScheduleChange{
			Kind:              kind,
			PreviousStartAt:   previous.StartAt,
			StartAt:           segment.StartAt,
			PreviousEndAt:     previous.EndAt,
			EndAt:             segment.EndAt,
			PreviousLocalDate: previousLocalDate,
			LocalDate:         localDateValue,
			ShiftMinutes:      shiftMinutes,
		}
		selectedSegmentIndex = segment.SegmentIndex
		selectedFound = true
		selectedDayChanged = dayChanged
	}

	for _, segment := range segments {
		previous, exists := previousBySegment[segment.SegmentIndex]
		if exists {
			consider(previous, segment)
		}
	}
	for _, segment := range segments {
		if _, matched := previousBySegment[segment.SegmentIndex]; matched || len(previousBlocks) == 0 {
			continue
		}
		consider(nearestScheduleBlock(previousBlocks, segment.SegmentIndex), segment)
	}
	for _, previous := range previousBlocks {
		if _, matched := segmentsByIndex[previous.SegmentIndex]; matched || len(segments) == 0 {
			continue
		}
		consider(previous, nearestScheduleSegment(segments, previous.SegmentIndex))
	}

	return selected, selectedFound
}

func scheduleTimesEqual(previousBlocks []calendar.CoreScheduleBlock, segments []calendar.MayaScheduleSegmentInput) bool {
	if len(previousBlocks) != len(segments) {
		return false
	}
	previousTimes := make([]calendar.CoreScheduleBlock, len(previousBlocks))
	copy(previousTimes, previousBlocks)
	segmentTimes := make([]calendar.MayaScheduleSegmentInput, len(segments))
	copy(segmentTimes, segments)
	sort.Slice(previousTimes, func(i, j int) bool {
		if !previousTimes[i].StartAt.Equal(previousTimes[j].StartAt) {
			return previousTimes[i].StartAt.Before(previousTimes[j].StartAt)
		}
		return previousTimes[i].EndAt.Before(previousTimes[j].EndAt)
	})
	sort.Slice(segmentTimes, func(i, j int) bool {
		if !segmentTimes[i].StartAt.Equal(segmentTimes[j].StartAt) {
			return segmentTimes[i].StartAt.Before(segmentTimes[j].StartAt)
		}
		return segmentTimes[i].EndAt.Before(segmentTimes[j].EndAt)
	})
	for index := range previousTimes {
		if !previousTimes[index].StartAt.Equal(segmentTimes[index].StartAt) || !previousTimes[index].EndAt.Equal(segmentTimes[index].EndAt) {
			return false
		}
	}
	return true
}

func nearestScheduleBlock(blocks []calendar.CoreScheduleBlock, segmentIndex int) calendar.CoreScheduleBlock {
	nearest := blocks[0]
	nearestDistance := absoluteInt(nearest.SegmentIndex - segmentIndex)
	for _, block := range blocks[1:] {
		distance := absoluteInt(block.SegmentIndex - segmentIndex)
		if distance < nearestDistance || (distance == nearestDistance && block.SegmentIndex < nearest.SegmentIndex) {
			nearest = block
			nearestDistance = distance
		}
	}
	return nearest
}

func nearestScheduleSegment(segments []calendar.MayaScheduleSegmentInput, segmentIndex int) calendar.MayaScheduleSegmentInput {
	nearest := segments[0]
	nearestDistance := absoluteInt(nearest.SegmentIndex - segmentIndex)
	for _, segment := range segments[1:] {
		distance := absoluteInt(segment.SegmentIndex - segmentIndex)
		if distance < nearestDistance || (distance == nearestDistance && segment.SegmentIndex < nearest.SegmentIndex) {
			nearest = segment
			nearestDistance = distance
		}
	}
	return nearest
}

func equalScheduleReasonChanged(previous *string, next string) bool {
	return previous == nil || *previous != next
}

func scheduleBlockBounds(blocks []calendar.CoreScheduleBlock) (*time.Time, *time.Time) {
	if len(blocks) == 0 {
		return nil, nil
	}
	startAt := blocks[0].StartAt
	endAt := blocks[0].EndAt
	for _, block := range blocks[1:] {
		if block.StartAt.Before(startAt) {
			startAt = block.StartAt
		}
		if block.EndAt.After(endAt) {
			endAt = block.EndAt
		}
	}
	return &startAt, &endAt
}

func scheduleSegmentBounds(segments []calendar.MayaScheduleSegmentInput) (*time.Time, *time.Time) {
	if len(segments) == 0 {
		return nil, nil
	}
	startAt := segments[0].StartAt
	endAt := segments[0].EndAt
	for _, segment := range segments[1:] {
		if segment.StartAt.Before(startAt) {
			startAt = segment.StartAt
		}
		if segment.EndAt.After(endAt) {
			endAt = segment.EndAt
		}
	}
	return &startAt, &endAt
}

func localDate(value *time.Time, location *time.Location) string {
	if value == nil {
		return ""
	}
	return value.In(location).Format(time.DateOnly)
}

func absoluteInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func uuidSliceContains(values []uuid.UUID, target uuid.UUID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (s *Service) failScheduleRun(ctx context.Context, run CoreRun, runErr error) error {
	message := runErr.Error()
	_, completeErr := s.repo.CompleteRun(ctx, run.ID, RunStatusFailed, "Automatic schedule reconciliation failed.", &message)
	return errors.Join(runErr, completeErr)
}

func (s *Service) markScheduleActionsFailed(ctx context.Context, actions []CoreAction, actionErr error) {
	for _, action := range actions {
		_ = s.repo.MarkActionFailed(ctx, action.ID, actionErr.Error())
	}
}

func appendUniqueUUID(values []uuid.UUID, value uuid.UUID) []uuid.UUID {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func (s *Service) scheduleRepository() (ScheduleRepository, error) {
	repo, ok := s.repo.(ScheduleRepository)
	if !ok {
		return nil, ErrNotConfigured
	}
	return repo, nil
}

func (s *Service) scheduleCalendarService() (ScheduleCalendarService, error) {
	calendarService, ok := s.calendar.(ScheduleCalendarService)
	if !ok {
		return nil, ErrNotConfigured
	}
	return calendarService, nil
}

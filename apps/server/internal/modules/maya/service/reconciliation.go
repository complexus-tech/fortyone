package maya

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
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
	// Story events maintain schedules Maya already owns. They must not opt every
	// ordinary assigned story into autonomous scheduling without a product-level
	// auto-schedule preference.
	if len(owners) == 0 {
		return nil, nil
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
	if story.Assignee != nil && *story.Assignee != uuid.Nil {
		schedulable, eligibilityErr := scheduleRepo.StoryIsSchedulableForUser(ctx, ref.WorkspaceID, ref.StoryID, *story.Assignee)
		if eligibilityErr != nil {
			return affectedUsers, failRun(eligibilityErr)
		}
		if schedulable {
			desiredOwner = *story.Assignee
			affectedUsers = appendUniqueUUID(affectedUsers, desiredOwner)
		}
	}
	keepOwnershipByOwner := make(map[uuid.UUID]bool, len(owners))
	if desiredOwner != uuid.Nil {
		keepOwnershipByOwner[desiredOwner] = true
	} else {
		for _, ownerID := range owners {
			retainable, retentionErr := scheduleRepo.StoryScheduleOwnershipIsRetainable(ctx, ref.WorkspaceID, ref.StoryID, ownerID)
			if retentionErr != nil {
				return affectedUsers, failRun(retentionErr)
			}
			keepOwnershipByOwner[ownerID] = retainable
		}
	}

	actions := []CoreAction{}
	summary := "Maya removed scheduled work because the story is no longer eligible for automatic scheduling."
	desiredSegments := []calendar.MayaScheduleSegmentInput{}
	if desiredOwner != uuid.Nil {
		planResult, planErr := s.planAssignedStory(ctx, story, desiredOwner)
		if planErr != nil {
			return affectedUsers, failRun(planErr)
		}
		summary = planResult.Summary
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
	}
	if !recordNoop {
		unchanged, stateErr := s.recoveryScheduleStateMatches(ctx, scheduleCalendar, ref, owners, desiredOwner, desiredSegments)
		if stateErr != nil {
			return affectedUsers, stateErr
		}
		if unchanged {
			if desiredOwner != uuid.Nil {
				_, stateErr = scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
					WorkspaceID: ref.WorkspaceID, UserID: desiredOwner, StoryID: ref.StoryID,
					ExpectedStoryUpdatedAt: &story.UpdatedAt, Segments: desiredSegments, KeepOwnership: true,
				})
			} else {
				for _, ownerID := range owners {
					if _, reconcileErr := scheduleCalendar.ReconcileMayaScheduleBlocks(ctx, calendar.MayaScheduleReconcileInput{
						WorkspaceID: ref.WorkspaceID, UserID: ownerID, StoryID: ref.StoryID,
						ExpectedStoryUpdatedAt: &story.UpdatedAt, KeepOwnership: keepOwnershipByOwner[ownerID],
					}); reconcileErr != nil {
						stateErr = errors.Join(stateErr, reconcileErr)
					}
				}
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
			ExpectedStoryUpdatedAt: &story.UpdatedAt, KeepOwnership: keepOwnershipByOwner[ownerID],
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
	return affectedUsers, nil
}

func (s *Service) recoveryScheduleStateMatches(
	ctx context.Context,
	scheduleCalendar ScheduleCalendarService,
	ref ScheduleStoryRef,
	owners []uuid.UUID,
	desiredOwner uuid.UUID,
	desiredSegments []calendar.MayaScheduleSegmentInput,
) (bool, error) {
	if desiredOwner != uuid.Nil {
		if len(owners) != 1 || owners[0] != desiredOwner {
			return false, nil
		}
		blocks, err := scheduleCalendar.ListMayaScheduleBlocksForStory(ctx, ref.WorkspaceID, desiredOwner, ref.StoryID)
		if err != nil {
			return false, err
		}
		return mayaScheduleSegmentsMatchBlocks(desiredSegments, blocks), nil
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

func mayaScheduleSegmentsMatchBlocks(segments []calendar.MayaScheduleSegmentInput, blocks []calendar.CoreScheduleBlock) bool {
	if len(segments) != len(blocks) {
		return false
	}
	blocksByIndex := make(map[int]calendar.CoreScheduleBlock, len(blocks))
	for _, block := range blocks {
		blocksByIndex[block.SegmentIndex] = block
	}
	for _, segment := range segments {
		block, ok := blocksByIndex[segment.SegmentIndex]
		if !ok || block.Title != segment.Title || !block.StartAt.Equal(segment.StartAt) || !block.EndAt.Equal(segment.EndAt) {
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
	for _, block := range schedule.Blocks {
		if block.WorkspaceID == story.Workspace && block.StoryID != nil && *block.StoryID == story.ID && block.Source == calendar.ScheduleBlockSourceMaya {
			continue
		}
		blocks = append(blocks, block)
	}
	workingDays, err := s.getWorkingDays(ctx, story.Team, story.Workspace)
	if err != nil {
		return PlanResult{}, err
	}
	return s.planner.Plan(PlanInput{
		Context: ctx, WorkspaceID: story.Workspace, Story: story,
		WindowStart: windowStart, WindowEnd: windowEnd, WorkingDays: workingDays,
		MinimumFocusBlockMinutes: valueOrZero(story.MinimumFocusBlockMinutes),
		Candidates: []CandidateSchedule{{
			Member: reports.CoreMemberWorkload{UserID: userID}, Timezone: schedule.Timezone,
			BusyWindows: schedule.BusyWindows, Blocks: blocks,
		}},
	})
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

package maya

import (
	"context"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func (f *fakeMayaRepository) CreateRun(_ context.Context, input CreateRunInput) (CoreRun, error) {
	f.createRunCalls++
	f.run = CoreRun{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		StoryID:     input.StoryID,
		TriggeredBy: input.TriggeredBy,
		Trigger:     input.Trigger,
		Status:      RunStatusRunning,
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Context:     input.Context,
	}
	return f.run, nil
}

func (f *fakeMayaRepository) CompleteRun(_ context.Context, runID uuid.UUID, status RunStatus, summary string, message *string) (CoreRun, error) {
	f.completeRunCalls++
	f.run.ID = runID
	f.run.Status = status
	f.run.Summary = summary
	f.run.Error = message
	f.run.UpdatedAt = time.Now()
	completedAt := time.Now()
	f.run.CompletedAt = &completedAt
	return f.run, nil
}

func (f *fakeMayaRepository) CreateActions(_ context.Context, actions []CoreAction) ([]CoreAction, error) {
	for i := range actions {
		actions[i].ID = uuid.New()
	}
	f.actions = actions
	return actions, nil
}

func (f *fakeMayaRepository) GetWorkPlan(_ context.Context, runID, workspaceID, triggeredBy uuid.UUID) (WorkPlan, error) {
	if f.run.ID != runID || f.run.WorkspaceID != workspaceID || f.run.TriggeredBy != triggeredBy {
		return WorkPlan{}, ErrPlanNotFound
	}
	return WorkPlan{Run: f.run, Actions: append([]CoreAction(nil), f.actions...)}, nil
}

func (f *fakeMayaRepository) MarkActionApplied(_ context.Context, actionID uuid.UUID) error {
	f.appliedActionIDs = append(f.appliedActionIDs, actionID)
	if f.markActionAppliedErr != nil {
		return f.markActionAppliedErr
	}
	for index := range f.actions {
		if f.actions[index].ID == actionID && f.actions[index].Status == ActionStatusProposed {
			f.actions[index].Status = ActionStatusApplied
		}
	}
	return nil
}

func (f *fakeMayaRepository) MarkActionFailed(_ context.Context, actionID uuid.UUID, message string) error {
	f.failedActionIDs = append(f.failedActionIDs, actionID)
	f.failedActionMessages = append(f.failedActionMessages, message)
	if f.markActionFailedErr != nil {
		return f.markActionFailedErr
	}
	for index := range f.actions {
		if f.actions[index].ID == actionID && f.actions[index].Status == ActionStatusProposed {
			f.actions[index].Status = ActionStatusFailed
			f.actions[index].Error = &message
		}
	}
	return nil
}

func (f *fakeMayaRepository) ListScheduleStoryRefsForUser(_ context.Context, _ uuid.UUID) ([]ScheduleStoryRef, error) {
	return append([]ScheduleStoryRef(nil), f.storyRefs...), nil
}

func (f *fakeMayaRepository) ClaimScheduleRecoveryStoryRefs(_ context.Context, limit int, retryBefore, interruptedRunBefore time.Time) ([]ScheduleRecoveryRef, error) {
	f.recoveryClaimLimit = limit
	f.recoveryRetryBefore = retryBefore
	f.recoveryInterruptedRunBefore = interruptedRunBefore
	if f.recoveryRequiresOwnership && len(f.scheduleOwners) == 0 {
		return nil, nil
	}
	refs := append([]ScheduleRecoveryRef(nil), f.recoveryRefs...)
	if limit > 0 && len(refs) > limit {
		refs = refs[:limit]
	}
	return refs, nil
}

func (f *fakeMayaRepository) CompleteInterruptedScheduleRun(_ context.Context, runID uuid.UUID, _ string) error {
	f.completedInterruptedRunIDs = append(f.completedInterruptedRunIDs, runID)
	return nil
}

func (f *fakeMayaRepository) ListMayaScheduleOwners(_ context.Context, _, _ uuid.UUID) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), f.scheduleOwners...), nil
}

func (f *fakeMayaRepository) StoryIsSchedulableForUser(_ context.Context, _, _, _ uuid.UUID) (bool, error) {
	return f.schedulable, nil
}

func (f *fakeMayaRepository) WorkspaceCanUseMaya(_ context.Context, _ uuid.UUID) (bool, error) {
	return !f.workspaceAccessDenied, nil
}

func (f *fakeMayaRepository) StoryIsActiveForAutoScheduling(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return !f.storyInactive, nil
}

func (f *fakeMayaRepository) StoryScheduleOwnershipIsRetainable(_ context.Context, _, _, _ uuid.UUID) (bool, error) {
	return f.ownershipRetainable, nil
}

func (f *fakeMayaRepository) WithScheduleStoryLock(_ context.Context, _, _ uuid.UUID, reconcile func() error) error {
	f.scheduleLock.Lock()
	defer f.scheduleLock.Unlock()
	f.scheduleLockCalls++
	return reconcile()
}

type fakeMayaStories struct {
	story                       stories.CoreSingleStory
	actorID                     uuid.UUID
	updatedAssignee             *uuid.UUID
	updatedStartDate            *time.Time
	updatedEndDate              *time.Time
	lastUpdates                 map[string]any
	updateReasons               []string
	assignmentUpdateErr         error
	assignmentExpectedUpdatedAt time.Time
	automationStates            []string
	scheduleTransitions         []*events.StoryScheduleTransition
}

func (f *fakeMayaStories) Get(_ context.Context, storyID, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	f.story.ID = storyID
	f.story.Workspace = workspaceID
	return f.story, nil
}

func (f *fakeMayaStories) UpdateExternal(_ context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any) error {
	return f.UpdateExternalWithReason(context.Background(), actorID, storyID, workspaceID, updates, "")
}

func (f *fakeMayaStories) UpdateExternalWithReason(_ context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any, reason string) error {
	f.actorID = actorID
	f.lastUpdates = updates
	f.updateReasons = append(f.updateReasons, reason)
	if _, isAssignmentUpdate := updates["assignee_id"]; isAssignmentUpdate && f.assignmentUpdateErr != nil {
		return f.assignmentUpdateErr
	}
	if value, ok := updates["assignee_id"].(uuid.UUID); ok {
		f.updatedAssignee = &value
		f.story.Assignee = &value
	}
	if value, ok := updates["start_date"].(time.Time); ok {
		f.updatedStartDate = &value
	}
	if value, ok := updates["end_date"].(time.Time); ok {
		f.updatedEndDate = &value
	}
	return nil
}

func (f *fakeMayaStories) UpdateExternalWithReasonIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any, reason string) error {
	f.assignmentExpectedUpdatedAt = expectedUpdatedAt
	if !f.story.UpdatedAt.IsZero() && !f.story.UpdatedAt.Equal(expectedUpdatedAt) {
		return stories.ErrStoryChanged
	}
	return f.UpdateExternalWithReason(ctx, actorID, storyID, workspaceID, updates, reason)
}

func (f *fakeMayaStories) UpdateAutomationIfUnchanged(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any, reason string) error {
	if _, isAssignment := updates["assignee_id"]; isAssignment {
		f.assignmentExpectedUpdatedAt = expectedUpdatedAt
	}
	if !f.story.UpdatedAt.IsZero() && !f.story.UpdatedAt.Equal(expectedUpdatedAt) {
		return stories.ErrStoryChanged
	}
	if value, ok := updates["auto_scheduling_enabled"].(bool); ok {
		f.story.AutoSchedulingEnabled = value
	}
	if value, ok := updates["auto_scheduling_locked"].(bool); ok {
		f.story.AutoSchedulingLocked = value
	}
	if value, ok := updates["estimated_duration_minutes"].(int); ok {
		f.story.EstimatedDurationMinutes = &value
	}
	if value, ok := updates["minimum_focus_block_minutes"].(int); ok {
		f.story.MinimumFocusBlockMinutes = &value
	}
	if err := f.UpdateExternalWithReason(ctx, actorID, storyID, workspaceID, updates, reason); err != nil {
		return err
	}
	f.story.UpdatedAt = time.Now().UTC()
	return nil
}

func (f *fakeMayaStories) UpdateAutomationStateIfUnchanged(_ context.Context, _, _, _ uuid.UUID, expectedUpdatedAt time.Time, status string, reason *string, locked *bool, schedule *events.StoryScheduleTransition) error {
	if !f.story.UpdatedAt.Equal(expectedUpdatedAt) {
		return stories.ErrStoryChanged
	}
	if locked != nil {
		f.story.AutoSchedulingLocked = *locked
	}
	f.story.AutoSchedulingStatus = status
	f.story.AutoSchedulingReason = reason
	f.automationStates = append(f.automationStates, status)
	f.scheduleTransitions = append(f.scheduleTransitions, schedule)
	return nil
}

type fakeMayaReports struct {
	analysis reports.CoreWorkloadAnalysis
}

func (f *fakeMayaReports) GetWorkloadAnalysis(_ context.Context, _ uuid.UUID, _ reports.ReportFilters) (reports.CoreWorkloadAnalysis, error) {
	return f.analysis, nil
}

type fakeMayaCalendar struct {
	createdBlock          calendar.CoreScheduleBlockInput
	reconciled            calendar.MayaScheduleReconcileInput
	reconciliations       []calendar.MayaScheduleReconcileInput
	schedulingView        calendar.CoreSchedule
	mayaBlocks            []calendar.CoreScheduleBlock
	reconcileErr          error
	dispatchErr           error
	dispatchedUsers       []uuid.UUID
	listScheduleCalls     int
	hasOwnership          bool
	ownerships            map[uuid.UUID]bool
	ownerRepo             *fakeMayaRepository
	onReconcile           func(calendar.MayaScheduleReconcileInput)
	currentStoryUpdatedAt *time.Time
	providerGate          func(calendar.MayaScheduleReconcileInput) string
	providerOperations    []string
}

func (f *fakeMayaCalendar) ListSchedulingAvailability(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (calendar.CoreSchedule, error) {
	f.listScheduleCalls++
	view := f.schedulingView
	view.StartAt = startAt
	view.EndAt = endAt
	return view, nil
}

func (f *fakeMayaCalendar) ListMayaScheduleBlocksForStory(_ context.Context, workspaceID, userID, storyID uuid.UUID) ([]calendar.CoreScheduleBlock, error) {
	blocks := make([]calendar.CoreScheduleBlock, 0)
	available := f.mayaBlocks
	if available == nil {
		available = f.schedulingView.Blocks
	}
	for _, block := range available {
		if block.WorkspaceID == workspaceID && block.UserID == userID && block.StoryID != nil && *block.StoryID == storyID && block.Source == calendar.ScheduleBlockSourceMaya {
			blocks = append(blocks, block)
		}
	}
	return blocks, nil
}

func (f *fakeMayaCalendar) MayaScheduleOwnershipExists(_ context.Context, _, userID, _ uuid.UUID) (bool, error) {
	if f.ownerships != nil {
		return f.ownerships[userID], nil
	}
	return f.hasOwnership, nil
}

func (f *fakeMayaCalendar) ReconcileMayaScheduleBlocks(_ context.Context, input calendar.MayaScheduleReconcileInput) (calendar.CoreScheduleReconcileResult, error) {
	if input.ExpectedStoryUpdatedAt != nil && f.currentStoryUpdatedAt != nil && !input.ExpectedStoryUpdatedAt.Equal(*f.currentStoryUpdatedAt) {
		return calendar.CoreScheduleReconcileResult{}, calendar.ErrCalendarScheduleStalePlan
	}
	f.reconciled = input
	f.reconciliations = append(f.reconciliations, input)
	if f.onReconcile != nil {
		f.onReconcile(input)
	}
	if f.reconcileErr != nil {
		return calendar.CoreScheduleReconcileResult{}, f.reconcileErr
	}
	owned := input.KeepOwnership || len(input.Segments) > 0
	f.hasOwnership = owned
	if f.ownerships == nil {
		f.ownerships = make(map[uuid.UUID]bool)
	}
	f.ownerships[input.UserID] = owned
	if f.ownerRepo != nil {
		if owned {
			f.ownerRepo.scheduleOwners = []uuid.UUID{input.UserID}
		} else {
			owners := make([]uuid.UUID, 0, len(f.ownerRepo.scheduleOwners))
			for _, ownerID := range f.ownerRepo.scheduleOwners {
				if ownerID != input.UserID {
					owners = append(owners, ownerID)
				}
			}
			f.ownerRepo.scheduleOwners = owners
		}
	}
	blocks := make([]calendar.CoreScheduleBlock, len(input.Segments))
	for index, segment := range input.Segments {
		storyID := input.StoryID
		blocks[index] = calendar.CoreScheduleBlock{
			ID: uuid.New(), WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: &storyID,
			Title: segment.Title, StartAt: segment.StartAt, EndAt: segment.EndAt,
			Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: segment.SegmentIndex, IsLocked: input.Locked,
		}
	}
	retainedBlocks := make([]calendar.CoreScheduleBlock, 0, len(f.mayaBlocks)+len(blocks))
	for _, block := range f.mayaBlocks {
		if block.WorkspaceID == input.WorkspaceID && block.UserID == input.UserID && block.StoryID != nil && *block.StoryID == input.StoryID {
			continue
		}
		retainedBlocks = append(retainedBlocks, block)
	}
	f.mayaBlocks = append(retainedBlocks, blocks...)
	if len(input.Segments) > 0 {
		first := input.Segments[0]
		storyID := input.StoryID
		f.createdBlock = calendar.CoreScheduleBlockInput{
			WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: &storyID,
			Title: first.Title, StartAt: first.StartAt, EndAt: first.EndAt,
			Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: first.SegmentIndex,
		}
	}
	return calendar.CoreScheduleReconcileResult{Blocks: blocks}, nil
}

func (f *fakeMayaCalendar) DispatchScheduleEventOutbox(_ context.Context, userID uuid.UUID) error {
	f.dispatchedUsers = append(f.dispatchedUsers, userID)
	if f.providerGate != nil {
		for index := len(f.reconciliations) - 1; index >= 0; index-- {
			if f.reconciliations[index].UserID == userID {
				f.providerOperations = append(f.providerOperations, f.providerGate(f.reconciliations[index]))
				break
			}
		}
	}
	return f.dispatchErr
}

func (f *fakeMayaCalendar) ListSchedule(_ context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (calendar.CoreSchedule, error) {
	f.listScheduleCalls++
	return calendar.CoreSchedule{StartAt: startAt, EndAt: endAt}, nil
}

func (f *fakeMayaCalendar) CreateScheduleBlock(_ context.Context, input calendar.CoreScheduleBlockInput) (calendar.CoreScheduleBlock, error) {
	f.createdBlock = input
	return calendar.CoreScheduleBlock{ID: uuid.New(), WorkspaceID: input.WorkspaceID, UserID: input.UserID, StoryID: input.StoryID, StartAt: input.StartAt, EndAt: input.EndAt, Source: input.Source}, nil
}

type fakeMayaUsers struct {
	members    []users.CoreUser
	lastFilter users.CoreListUsersFilter
}

func (f *fakeMayaUsers) List(_ context.Context, _ uuid.UUID, filter users.CoreListUsersFilter) ([]users.CoreUser, error) {
	f.lastFilter = filter
	if filter.Limit > 0 && len(f.members) > filter.Limit {
		return f.members[:filter.Limit], nil
	}
	return f.members, nil
}

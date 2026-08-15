package maya

import (
	"context"
	"sync"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func TestCreateWorkPlanPersistsAndAppliesActions(t *testing.T) {
	ctx := context.Background()
	workspaceID := uuid.New()
	storyID := uuid.New()
	requestedBy := uuid.New()
	mayaActorID := uuid.New()
	teamID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)
	duration := 60

	repo := &fakeMayaRepository{}
	storiesSvc := &fakeMayaStories{
		story: stories.CoreSingleStory{
			ID:                       storyID,
			Workspace:                workspaceID,
			Team:                     teamID,
			Title:                    "Schedule me",
			EstimatedDurationMinutes: &duration,
			AutoSchedulingEnabled:    true,
		},
	}
	calendarSvc := &fakeMayaCalendar{}
	service := New(Dependencies{
		Repository: repo,
		Stories:    storiesSvc,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: userID, FullName: "Ada", OpenStories: 1}},
		}},
		Calendar: calendarSvc,
		Users: &fakeMayaUsers{members: []users.CoreUser{
			{ID: userID, FullName: "Ada"},
		}},
		Planner:     NewPlanner(),
		MayaActorID: mayaActorID,
	})

	plan, err := service.CreateWorkPlan(ctx, CreateWorkPlanInput{
		WorkspaceID:     workspaceID,
		StoryID:         storyID,
		TriggeredBy:     requestedBy,
		Trigger:         RunTriggerManual,
		WindowStart:     startAt,
		WindowEnd:       endAt,
		DurationMinutes: duration,
		AutoApply:       true,
	})

	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if plan.Run.Status != RunStatusSucceeded {
		t.Fatalf("expected run status %q, got %q", RunStatusSucceeded, plan.Run.Status)
	}
	if len(plan.Actions) != 2 {
		t.Fatalf("expected two actions, got %d", len(plan.Actions))
	}
	for _, action := range plan.Actions {
		if action.Status != ActionStatusApplied {
			t.Fatalf("expected returned action %s to be applied, got %q", action.ID, action.Status)
		}
	}
	if storiesSvc.updatedAssignee == nil || *storiesSvc.updatedAssignee != userID {
		t.Fatalf("expected story assignee update to %s, got %v", userID, storiesSvc.updatedAssignee)
	}
	if storiesSvc.actorID != mayaActorID {
		t.Fatalf("expected Maya actor %s, got %s", mayaActorID, storiesSvc.actorID)
	}
	if calendarSvc.createdBlock.Source != calendar.ScheduleBlockSourceMaya {
		t.Fatalf("expected Maya schedule block source, got %q", calendarSvc.createdBlock.Source)
	}
	if storiesSvc.updatedStartDate != nil || storiesSvc.updatedEndDate != nil {
		t.Fatalf("generated schedule must not mutate story constraint dates: start=%v end=%v", storiesSvc.updatedStartDate, storiesSvc.updatedEndDate)
	}
	if len(repo.appliedActionIDs) != 2 {
		t.Fatalf("expected two applied action marks, got %d", len(repo.appliedActionIDs))
	}
	if len(storiesSvc.updateReasons) != 1 {
		t.Fatalf("expected one reason-aware assignment update, got %d", len(storiesSvc.updateReasons))
	}
	for _, reason := range storiesSvc.updateReasons {
		if reason == "" {
			t.Fatal("expected Maya story update reason to be recorded")
		}
	}
}

func TestCreateWorkPlanUsesBoundedDefaultCandidatePool(t *testing.T) {
	ctx := context.Background()
	workspaceID := uuid.New()
	storyID := uuid.New()
	requestedBy := uuid.New()
	teamID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 6, 15, 17, 0, 0, 0, time.UTC)
	members := make([]users.CoreUser, defaultCandidateLimit+20)
	for i := range members {
		members[i] = users.CoreUser{ID: uuid.New(), FullName: "Candidate"}
	}

	usersSvc := &fakeMayaUsers{members: members}
	calendarSvc := &fakeMayaCalendar{}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{},
		Stories: &fakeMayaStories{story: stories.CoreSingleStory{
			ID:        storyID,
			Workspace: workspaceID,
			Team:      teamID,
			Title:     "Bounded plan",
		}},
		Reports:     &fakeMayaReports{},
		Calendar:    calendarSvc,
		Users:       usersSvc,
		Planner:     NewPlanner(),
		MayaActorID: uuid.New(),
	})

	if _, err := service.CreateWorkPlan(ctx, CreateWorkPlanInput{
		WorkspaceID:     workspaceID,
		StoryID:         storyID,
		TriggeredBy:     requestedBy,
		WindowStart:     startAt,
		WindowEnd:       endAt,
		DurationMinutes: 60,
	}); err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}

	if usersSvc.lastFilter.Limit != defaultCandidateLimit {
		t.Fatalf("expected users list limit %d, got %d", defaultCandidateLimit, usersSvc.lastFilter.Limit)
	}
	if calendarSvc.listScheduleCalls != defaultCandidateLimit {
		t.Fatalf("expected %d schedule lookups, got %d", defaultCandidateLimit, calendarSvc.listScheduleCalls)
	}
}

func TestCreateWorkPlanUsesAccountWideAvailabilityTimezoneAndMinimumFocus(t *testing.T) {
	workspaceID := uuid.New()
	otherWorkspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	duration := 150
	minimumFocus := 60
	windowStart := time.Date(2026, 6, 15, 6, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	calendarSvc := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{
		Timezone: "Africa/Harare",
		Blocks: []calendar.CoreScheduleBlock{{
			WorkspaceID: otherWorkspaceID, UserID: userID, Title: "Busy",
			StartAt: time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC),
			EndAt:   time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC),
			Source:  calendar.ScheduleBlockSourceUser, IsLocked: true,
		}},
	}}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{},
		Stories: &fakeMayaStories{story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Account-wide plan",
			EstimatedDurationMinutes: &duration, MinimumFocusBlockMinutes: &minimumFocus,
		}},
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: userID}}}},
		Calendar: calendarSvc, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: userID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: windowStart, WindowEnd: windowEnd,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if len(plan.Actions) != 3 {
		t.Fatalf("expected assignment plus two focus segments: %#v", plan.Actions)
	}
	firstSegment := plan.Actions[1].Payload.ScheduleBlock
	secondSegment := plan.Actions[2].Payload.ScheduleBlock
	if firstSegment == nil || secondSegment == nil {
		t.Fatalf("expected two schedule payloads: %#v", plan.Actions)
	}
	if expected := time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC); !firstSegment.StartAt.Equal(expected) {
		t.Fatalf("expected other-workspace block to defer local 09:00 start until %s, got %s", expected, firstSegment.StartAt)
	}
	if firstSegment.EndAt.Sub(firstSegment.StartAt) < time.Hour || secondSegment.EndAt.Sub(secondSegment.StartAt) < time.Hour {
		t.Fatalf("expected story minimum focus to reach planner: %#v %#v", firstSegment, secondSegment)
	}
	if firstSegment.EndAt.Sub(firstSegment.StartAt)+secondSegment.EndAt.Sub(secondSegment.StartAt) != 150*time.Minute {
		t.Fatalf("expected exact task duration across segments: %#v %#v", firstSegment, secondSegment)
	}
}

func TestCreateWorkPlanPreservesStoryConstraintDates(t *testing.T) {
	ctx := context.Background()
	workspaceID := uuid.New()
	storyID := uuid.New()
	requestedBy := uuid.New()
	mayaActorID := uuid.New()
	teamID := uuid.New()
	userID := uuid.New()
	existingStartDate := time.Date(2026, 6, 20, 9, 0, 0, 0, time.UTC)
	existingEndDate := time.Date(2026, 6, 24, 17, 0, 0, 0, time.UTC)
	windowStart := time.Date(2026, 6, 16, 9, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 6, 30, 17, 0, 0, 0, time.UTC)

	storiesSvc := &fakeMayaStories{
		story: stories.CoreSingleStory{
			ID:        storyID,
			Workspace: workspaceID,
			Team:      teamID,
			Title:     "Keep explicit dates",
			StartDate: &existingStartDate,
			EndDate:   &existingEndDate,
		},
	}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{},
		Stories:    storiesSvc,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: userID, FullName: "Ada", OpenStories: 1}},
		}},
		Calendar: &fakeMayaCalendar{},
		Users: &fakeMayaUsers{members: []users.CoreUser{
			{ID: userID, FullName: "Ada"},
		}},
		Planner:     NewPlanner(),
		MayaActorID: mayaActorID,
	})

	if _, err := service.CreateWorkPlan(ctx, CreateWorkPlanInput{
		WorkspaceID:     workspaceID,
		StoryID:         storyID,
		TriggeredBy:     requestedBy,
		Trigger:         RunTriggerManual,
		WindowStart:     windowStart,
		WindowEnd:       windowEnd,
		DurationMinutes: 60,
		AutoApply:       true,
	}); err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}

	if storiesSvc.updatedStartDate != nil || storiesSvc.updatedEndDate != nil {
		t.Fatalf("generated calendar placement must not overwrite story constraints: start=%v end=%v", storiesSvc.updatedStartDate, storiesSvc.updatedEndDate)
	}
}

func TestShouldIncludeCandidateExcludesMayaActor(t *testing.T) {
	mayaActorID := uuid.New()
	humanUserID := uuid.New()

	if shouldIncludeCandidate(mayaActorID, nil, mayaActorID) {
		t.Fatal("expected Maya actor to be excluded from assignment candidates")
	}
	if !shouldIncludeCandidate(humanUserID, nil, mayaActorID) {
		t.Fatal("expected human user to be included in assignment candidates")
	}
}

type fakeMayaRepository struct {
	actions                      []CoreAction
	createRunCalls               int
	appliedActionIDs             []uuid.UUID
	storyRefs                    []ScheduleStoryRef
	recoveryRefs                 []ScheduleRecoveryRef
	recoveryClaimLimit           int
	recoveryRetryBefore          time.Time
	recoveryInterruptedRunBefore time.Time
	recoveryRequiresOwnership    bool
	completedInterruptedRunIDs   []uuid.UUID
	scheduleOwners               []uuid.UUID
	schedulable                  bool
	ownershipRetainable          bool
	workspaceAccessDenied        bool
	storyInactive                bool
	scheduleLock                 sync.Mutex
	scheduleLockCalls            int
}

func (f *fakeMayaRepository) CreateRun(_ context.Context, input CreateRunInput) (CoreRun, error) {
	f.createRunCalls++
	return CoreRun{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		StoryID:     input.StoryID,
		TriggeredBy: input.TriggeredBy,
		Trigger:     input.Trigger,
		Status:      RunStatusRunning,
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (f *fakeMayaRepository) CompleteRun(_ context.Context, runID uuid.UUID, status RunStatus, summary string, message *string) (CoreRun, error) {
	return CoreRun{
		ID:        runID,
		Status:    status,
		Summary:   summary,
		Error:     message,
		UpdatedAt: time.Now(),
	}, nil
}

func (f *fakeMayaRepository) CreateActions(_ context.Context, actions []CoreAction) ([]CoreAction, error) {
	for i := range actions {
		actions[i].ID = uuid.New()
	}
	f.actions = actions
	return actions, nil
}

func (f *fakeMayaRepository) MarkActionApplied(_ context.Context, actionID uuid.UUID) error {
	f.appliedActionIDs = append(f.appliedActionIDs, actionID)
	return nil
}

func (f *fakeMayaRepository) MarkActionFailed(_ context.Context, _ uuid.UUID, _ string) error {
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

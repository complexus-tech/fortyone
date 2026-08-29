package maya

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
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

func TestPreviewThenApplyStoredWorkPlanExactlyOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workspaceID := uuid.New()
	storyID := uuid.New()
	requestedBy := uuid.New()
	mayaActorID := uuid.New()
	teamID := uuid.New()
	userID := uuid.New()
	storyUpdatedAt := time.Date(2026, 8, 27, 8, 0, 0, 0, time.UTC)
	windowStart := time.Date(2026, 8, 28, 7, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(8 * time.Hour)
	duration := 60

	repo := &fakeMayaRepository{}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID:        storyID,
		Workspace: workspaceID,
		Team:      teamID,
		Title:     "Prepare launch brief",
		UpdatedAt: storyUpdatedAt,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "Africa/Harare"}}
	service := New(Dependencies{
		Repository: repo,
		Stories:    storiesService,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: userID, FullName: "Ada"}},
		}},
		Calendar: calendarService,
		Users:    &fakeMayaUsers{members: []users.CoreUser{{ID: userID, FullName: "Ada"}}},
		Planner:  NewPlanner(), MayaActorID: mayaActorID,
	})

	preview, err := service.CreateWorkPlan(ctx, CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: requestedBy,
		Trigger: RunTriggerManual, WindowStart: windowStart, WindowEnd: windowEnd,
		DurationMinutes: duration, AutoApply: false,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan preview returned error: %v", err)
	}
	if len(preview.Actions) != 2 {
		t.Fatalf("expected assignment and schedule previews, got %#v", preview.Actions)
	}
	for _, action := range preview.Actions {
		if action.Status != ActionStatusProposed {
			t.Fatalf("preview action %s was mutated: %q", action.ID, action.Status)
		}
	}
	if storiesService.updatedAssignee != nil || storiesService.story.AutoSchedulingEnabled {
		t.Fatalf("preview changed the story: %#v", storiesService.story)
	}
	if len(repo.appliedActionIDs) != 0 {
		t.Fatalf("preview applied actions: %v", repo.appliedActionIDs)
	}

	applied, err := service.ApplyWorkPlan(ctx, ApplyWorkPlanInput{
		WorkspaceID: workspaceID,
		RunID:       preview.Run.ID,
		TriggeredBy: requestedBy,
	})
	if err != nil {
		t.Fatalf("ApplyWorkPlan returned error: %v", err)
	}
	for _, action := range applied.Actions {
		if action.Status != ActionStatusApplied {
			t.Fatalf("expected applied action %s, got %q", action.ID, action.Status)
		}
	}
	if repo.createRunCalls != 1 {
		t.Fatalf("apply recomputed the plan and created %d runs", repo.createRunCalls)
	}
	if storiesService.updatedAssignee == nil || *storiesService.updatedAssignee != userID {
		t.Fatalf("expected stored assignee %s, got %v", userID, storiesService.updatedAssignee)
	}
	if !storiesService.story.AutoSchedulingEnabled || storiesService.story.EstimatedDurationMinutes == nil || *storiesService.story.EstimatedDurationMinutes != duration {
		t.Fatalf("approved scheduling preferences were not saved: %#v", storiesService.story)
	}
	if len(repo.appliedActionIDs) != 2 {
		t.Fatalf("expected two durable applied actions, got %v", repo.appliedActionIDs)
	}

	if _, err := service.ApplyWorkPlan(ctx, ApplyWorkPlanInput{
		WorkspaceID: workspaceID,
		RunID:       preview.Run.ID,
		TriggeredBy: requestedBy,
	}); err != nil {
		t.Fatalf("idempotent ApplyWorkPlan retry returned error: %v", err)
	}
	if len(repo.appliedActionIDs) != 2 {
		t.Fatalf("idempotent retry reapplied actions: %v", repo.appliedActionIDs)
	}
	if _, err := service.ApplyWorkPlan(ctx, ApplyWorkPlanInput{
		WorkspaceID: workspaceID,
		RunID:       preview.Run.ID,
		TriggeredBy: uuid.New(),
	}); !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf("expected user-scoped plan lookup to fail, got %v", err)
	}
}

func TestApplyStoredWorkPlanRefreshesScheduleAfterSavingPreferences(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	storyID := uuid.New()
	requestedBy := uuid.New()
	userID := uuid.New()
	updatedAt := time.Date(2026, 8, 27, 9, 0, 0, 0, time.UTC)
	windowStart := time.Date(2026, 8, 28, 7, 0, 0, 0, time.UTC)
	repo := &fakeMayaRepository{}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Review launch",
		Assignee: &userID, UpdatedAt: updatedAt,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "Africa/Harare"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService,
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: userID}}}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: userID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	preview, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: requestedBy,
		WindowStart: windowStart, WindowEnd: windowStart.Add(8 * time.Hour),
		DurationMinutes: 60,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan preview returned error: %v", err)
	}
	if len(preview.Actions) != 1 || preview.Actions[0].Type != ActionTypeScheduleWorkBlock {
		t.Fatalf("expected an exact schedule-only preview, got %#v", preview.Actions)
	}
	if _, err := service.ApplyWorkPlan(context.Background(), ApplyWorkPlanInput{
		WorkspaceID: workspaceID, RunID: preview.Run.ID, TriggeredBy: requestedBy,
	}); err != nil {
		t.Fatalf("ApplyWorkPlan returned error: %v", err)
	}
	if len(calendarService.reconciliations) < 2 {
		t.Fatalf("expected schedule commit and ownership refresh, got %#v", calendarService.reconciliations)
	}
	refreshed := calendarService.reconciliations[len(calendarService.reconciliations)-1]
	if refreshed.ExpectedStoryUpdatedAt == nil || !refreshed.ExpectedStoryUpdatedAt.Equal(storiesService.story.UpdatedAt) {
		t.Fatalf("schedule ownership was not refreshed to story version %s: %#v", storiesService.story.UpdatedAt, refreshed)
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

func TestManuallyScheduleStoryLocksExplicitConflictingTime(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	mayaActorID := uuid.New()
	duration := 90
	updatedAt := time.Date(2026, 8, 20, 8, 0, 0, 0, time.UTC)
	startAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Minute)
	repo := &fakeMayaRepository{schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID:                       storyID,
		Workspace:                workspaceID,
		Title:                    "Prepare launch brief",
		Assignee:                 &userID,
		EstimatedDurationMinutes: &duration,
		AutoSchedulingEnabled:    true,
		AutoSchedulingStatus:     stories.AutoSchedulingStatusCannotFit,
		UpdatedAt:                updatedAt,
	}}
	calendarService := &fakeMayaCalendar{hasOwnership: true}
	service := New(Dependencies{
		Repository:  repo,
		Stories:     storiesService,
		Reports:     &fakeMayaReports{},
		Calendar:    calendarService,
		Users:       &fakeMayaUsers{},
		Planner:     NewPlanner(),
		MayaActorID: mayaActorID,
	})

	block, err := service.ManuallyScheduleStory(context.Background(), ManualScheduleStoryInput{
		WorkspaceID: workspaceID,
		StoryID:     storyID,
		UserID:      userID,
		StartAt:     startAt,
		Timezone:    "Africa/Harare",
	})
	if err != nil {
		t.Fatalf("ManuallyScheduleStory returned error: %v", err)
	}
	if !block.StartAt.Equal(startAt) || !block.EndAt.Equal(startAt.Add(90*time.Minute)) || !block.IsLocked {
		t.Fatalf("unexpected manual block: %#v", block)
	}
	if !calendarService.reconciled.AllowConflicts || !calendarService.reconciled.Locked || !calendarService.reconciled.KeepOwnership {
		t.Fatalf("explicit placement did not preserve the override contract: %#v", calendarService.reconciled)
	}
	if !storiesService.story.AutoSchedulingLocked || storiesService.story.AutoSchedulingStatus != stories.AutoSchedulingStatusLocked {
		t.Fatalf("story was not locked after explicit placement: %#v", storiesService.story)
	}
	if len(storiesService.scheduleTransitions) != 1 || storiesService.scheduleTransitions[0] == nil || storiesService.scheduleTransitions[0].StartAt == nil || !storiesService.scheduleTransitions[0].StartAt.Equal(startAt) {
		t.Fatalf("expected one schedule transition for the chosen time: %#v", storiesService.scheduleTransitions)
	}
	if len(calendarService.dispatchedUsers) != 1 || calendarService.dispatchedUsers[0] != userID {
		t.Fatalf("expected provider outbox dispatch for %s, got %v", userID, calendarService.dispatchedUsers)
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
	run                          CoreRun
	createRunCalls               int
	completeRunCalls             int
	appliedActionIDs             []uuid.UUID
	failedActionIDs              []uuid.UUID
	failedActionMessages         []string
	markActionAppliedErr         error
	markActionFailedErr          error
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

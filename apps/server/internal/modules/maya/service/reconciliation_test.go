package maya

import (
	"context"
	"errors"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

func TestReconcileScheduleMovesGeneratedSegmentsToNewOwner(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	previousOwnerID := uuid.New()
	newOwnerID := uuid.New()
	duration := 60
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{previousOwnerID}, schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Move ownership",
		Assignee: &newOwnerID, EstimatedDurationMinutes: &duration,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID})
	if err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 2 {
		t.Fatalf("expected new-owner upsert and old-owner cleanup: %#v", calendarService.reconciliations)
	}
	if first := calendarService.reconciliations[0]; first.UserID != newOwnerID || len(first.Segments) != 1 {
		t.Fatalf("expected new owner to receive one segment first: %#v", first)
	}
	if second := calendarService.reconciliations[1]; second.UserID != previousOwnerID || len(second.Segments) != 0 {
		t.Fatalf("expected previous owner segments to be removed: %#v", second)
	}
	if storiesService.updatedAssignee != nil {
		t.Fatalf("automatic reconciliation must not mutate assignment: %v", storiesService.updatedAssignee)
	}
	if !containsUUID(calendarService.dispatchedUsers, previousOwnerID) || !containsUUID(calendarService.dispatchedUsers, newOwnerID) {
		t.Fatalf("expected both owners' Google outboxes to be dispatched: %v", calendarService.dispatchedUsers)
	}
}

func TestRecoverScheduleOwnershipsRepairsScheduleCommittedBeforeAssignment(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	stalePlannedOwnerID := uuid.New()
	actualOwnerID := uuid.New()
	duration := 60
	interruptedRunID := uuid.New()
	repo := &fakeMayaRepository{
		recoveryRefs: []ScheduleRecoveryRef{{
			ScheduleStoryRef: ScheduleStoryRef{WorkspaceID: workspaceID, StoryID: storyID},
			InterruptedRunID: &interruptedRunID,
		}},
		scheduleOwners: []uuid.UUID{stalePlannedOwnerID},
		schedulable:    true,
	}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Recover interrupted assignment",
		Assignee: &actualOwnerID, EstimatedDurationMinutes: &duration,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	startedAt := time.Now().UTC()
	processed, err := service.RecoverScheduleOwnerships(context.Background(), 25)
	if err != nil {
		t.Fatalf("RecoverScheduleOwnerships returned error: %v", err)
	}
	if processed != 1 || repo.recoveryClaimLimit != 25 {
		t.Fatalf("expected one bounded ownership claim, processed=%d limit=%d", processed, repo.recoveryClaimLimit)
	}
	wantRetryCutoff := startedAt.Add(-scheduleRecoveryRetryDelay)
	if repo.recoveryRetryBefore.Before(wantRetryCutoff.Add(-time.Second)) || repo.recoveryRetryBefore.After(time.Now().UTC().Add(-scheduleRecoveryRetryDelay).Add(time.Second)) {
		t.Fatalf("unexpected recovery retry cutoff: %v", repo.recoveryRetryBefore)
	}
	if len(calendarService.reconciliations) != 2 {
		t.Fatalf("expected actual-owner replan and stale-owner cleanup: %#v", calendarService.reconciliations)
	}
	if current := calendarService.reconciliations[0]; current.UserID != actualOwnerID || len(current.Segments) == 0 || !current.KeepOwnership {
		t.Fatalf("expected recovery to follow the story's actual assignee: %#v", current)
	}
	if stale := calendarService.reconciliations[1]; stale.UserID != stalePlannedOwnerID || len(stale.Segments) != 0 {
		t.Fatalf("expected interrupted planned owner to be cleaned up: %#v", stale)
	}
	if storiesService.updatedAssignee != nil {
		t.Fatalf("recovery must never overwrite the current story assignee: %v", storiesService.updatedAssignee)
	}
	if len(repo.completedInterruptedRunIDs) != 1 || repo.completedInterruptedRunIDs[0] != interruptedRunID {
		t.Fatalf("expected interrupted run to be terminalized after recovery: %v", repo.completedInterruptedRunIDs)
	}
}

func TestRecoverScheduleOwnershipsReplansWithoutAChangeEvent(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	repo := &fakeMayaRepository{
		recoveryRefs:   []ScheduleRecoveryRef{{ScheduleStoryRef: ScheduleStoryRef{WorkspaceID: workspaceID, StoryID: storyID}}},
		scheduleOwners: []uuid.UUID{ownerID},
		schedulable:    true,
	}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Replan after settings change",
		Assignee: &ownerID, EstimatedDurationMinutes: &duration,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	processed, err := service.RecoverScheduleOwnerships(context.Background(), 100)
	if err != nil {
		t.Fatalf("RecoverScheduleOwnerships returned error: %v", err)
	}
	if processed != 1 || len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) == 0 {
		t.Fatalf("expected the durable ownership sweep to replan without an event: processed=%d reconciliations=%#v", processed, calendarService.reconciliations)
	}
}

func TestRecoverScheduleOwnershipsDoesNotCreateAuditRunForUnchangedState(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	storyIDCopy := storyID
	existing := calendar.CoreScheduleBlock{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: ownerID, StoryID: &storyIDCopy,
		Title: "Already canonical", StartAt: startAt, EndAt: startAt.Add(time.Hour),
		Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: 0,
	}
	repo := &fakeMayaRepository{
		recoveryRefs:   []ScheduleRecoveryRef{{ScheduleStoryRef: ScheduleStoryRef{WorkspaceID: workspaceID, StoryID: storyID}}},
		scheduleOwners: []uuid.UUID{ownerID}, schedulable: true,
	}
	calendarService := &fakeMayaCalendar{
		schedulingView: calendar.CoreSchedule{Timezone: "UTC", Blocks: []calendar.CoreScheduleBlock{existing}},
		mayaBlocks:     []calendar.CoreScheduleBlock{existing},
	}
	service := New(Dependencies{
		Repository: repo,
		Stories: &fakeMayaStories{story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: existing.Title,
			Assignee: &ownerID, EstimatedDurationMinutes: &duration, StartDate: &startAt,
		}},
		Reports: &fakeMayaReports{}, Calendar: calendarService, Users: &fakeMayaUsers{},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	processed, err := service.RecoverScheduleOwnerships(context.Background(), 100)
	if err != nil {
		t.Fatalf("RecoverScheduleOwnerships returned error: %v", err)
	}
	if processed != 1 || repo.createRunCalls != 0 {
		t.Fatalf("unchanged recovery must refresh its watermark without audit churn: processed=%d runs=%d", processed, repo.createRunCalls)
	}
	if len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) != 1 {
		t.Fatalf("unchanged state should still be version-checked and watermark-refreshed: %#v", calendarService.reconciliations)
	}
}

func TestReconcileScheduleRemovesSegmentsForTerminalStory(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{ownerID}, schedulable: false}
	calendarService := &fakeMayaCalendar{}
	service := New(Dependencies{
		Repository: repo,
		Stories: &fakeMayaStories{story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Completed work",
			Assignee: &ownerID, EstimatedDurationMinutes: &duration, CompletedAt: timePointer(time.Now().UTC()),
		}},
		Reports: &fakeMayaReports{}, Calendar: calendarService, Users: &fakeMayaUsers{},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 1 || calendarService.reconciliations[0].UserID != ownerID || len(calendarService.reconciliations[0].Segments) != 0 || calendarService.reconciliations[0].KeepOwnership {
		t.Fatalf("expected terminal story cleanup: %#v", calendarService.reconciliations)
	}
}

func TestRecoveryReleasesTerminalOwnershipOnce(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	completedAt := time.Now().UTC()
	repo := &fakeMayaRepository{
		recoveryRefs: []ScheduleRecoveryRef{{ScheduleStoryRef: ScheduleStoryRef{
			WorkspaceID: workspaceID,
			StoryID:     storyID,
		}}},
		recoveryRequiresOwnership: true,
		scheduleOwners:            []uuid.UUID{ownerID},
		schedulable:               false,
		ownershipRetainable:       false,
	}
	calendarService := &fakeMayaCalendar{ownerRepo: repo, hasOwnership: true}
	service := New(Dependencies{
		Repository: repo,
		Stories: &fakeMayaStories{story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Retire terminal ownership",
			Assignee: &ownerID, CompletedAt: &completedAt,
		}},
		Reports: &fakeMayaReports{}, Calendar: calendarService, Users: &fakeMayaUsers{},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	processed, err := service.RecoverScheduleOwnerships(context.Background(), 100)
	if err != nil {
		t.Fatalf("first recovery returned error: %v", err)
	}
	if processed != 1 || len(calendarService.reconciliations) != 1 || calendarService.reconciliations[0].KeepOwnership {
		t.Fatalf("terminal recovery must clean and release ownership once: processed=%d reconciliations=%#v", processed, calendarService.reconciliations)
	}
	if len(repo.scheduleOwners) != 0 || calendarService.hasOwnership {
		t.Fatalf("terminal ownership survived cleanup: owners=%v calendarOwned=%t", repo.scheduleOwners, calendarService.hasOwnership)
	}

	processed, err = service.RecoverScheduleOwnerships(context.Background(), 100)
	if err != nil {
		t.Fatalf("second recovery returned error: %v", err)
	}
	if processed != 0 || len(calendarService.reconciliations) != 1 {
		t.Fatalf("released ownership must not be reclaimed: processed=%d reconciliations=%#v", processed, calendarService.reconciliations)
	}
}

func TestReconcileScheduleKeepsOwnershipAcrossTemporaryZeroSegmentPlan(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{ownerID}, schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Temporarily unschedulable",
		Assignee: &ownerID,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "UTC"}, hasOwnership: true}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	input := ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}
	if err := service.ReconcileSchedule(context.Background(), input); err != nil {
		t.Fatalf("zero-segment reconciliation returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) != 0 || !calendarService.reconciliations[0].KeepOwnership {
		t.Fatalf("expected missing duration to clear segments while retaining ownership: %#v", calendarService.reconciliations)
	}
	if !calendarService.hasOwnership {
		t.Fatal("temporary scheduling risk must remain discoverable for a later trigger")
	}

	duration := 60
	storiesService.story.EstimatedDurationMinutes = &duration
	if err := service.ReconcileSchedule(context.Background(), input); err != nil {
		t.Fatalf("follow-up reconciliation returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 2 || len(calendarService.reconciliations[1].Segments) == 0 || !calendarService.reconciliations[1].KeepOwnership {
		t.Fatalf("expected later duration update to rediscover and schedule the owned story: %#v", calendarService.reconciliations)
	}
	if repo.scheduleLockCalls != 2 {
		t.Fatalf("expected every queued trigger to serialize the complete story replan, got %d locks", repo.scheduleLockCalls)
	}
}

func TestInitialNoSlotEnrollmentSchedulesAfterAvailabilityChanges(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	windowStart := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	windowEnd := windowStart.Add(8 * time.Hour)
	repo := &fakeMayaRepository{schedulable: true}
	calendarService := &fakeMayaCalendar{
		ownerRepo: repo,
		schedulingView: calendar.CoreSchedule{
			Timezone: "UTC",
			BusyWindows: []calendar.CoreBusyWindow{{
				StartAt: windowStart,
				EndAt:   windowEnd,
			}},
		},
	}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Retry after meeting moves",
	}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService,
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: ownerID}}}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: windowStart, WindowEnd: windowEnd, DurationMinutes: duration, AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if len(repo.scheduleOwners) != 1 || repo.scheduleOwners[0] != ownerID || len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) != 0 || !calendarService.reconciliations[0].KeepOwnership {
		t.Fatalf("expected initial no-slot plan to retain zero-segment ownership: plan=%#v owners=%v reconciliations=%#v", plan.Actions, repo.scheduleOwners, calendarService.reconciliations)
	}
	if storiesService.story.Assignee == nil || *storiesService.story.Assignee != ownerID {
		t.Fatalf("expected selected owner to be assigned before later reconciliation: %v", storiesService.story.Assignee)
	}

	calendarService.schedulingView.BusyWindows = nil
	storiesService.story.EstimatedDurationMinutes = &duration
	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("availability-triggered reconciliation returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 2 || len(calendarService.reconciliations[1].Segments) == 0 {
		t.Fatalf("expected the retained story to schedule after availability changed: %#v", calendarService.reconciliations)
	}
}

func TestAssignedMissingDurationEnrollmentSchedulesAfterEstimateAdded(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	repo := &fakeMayaRepository{schedulable: true}
	calendarService := &fakeMayaCalendar{ownerRepo: repo, schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Estimate later", Assignee: &ownerID,
	}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService,
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: ownerID}}}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	windowStart := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: windowStart, WindowEnd: windowStart.Add(8 * time.Hour), AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if len(plan.Actions) != 2 || len(repo.scheduleOwners) != 1 || repo.scheduleOwners[0] != ownerID || len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) != 0 {
		t.Fatalf("expected assigned missing-duration work to retain ownership: plan=%#v owners=%v reconciliations=%#v", plan.Actions, repo.scheduleOwners, calendarService.reconciliations)
	}

	duration := 45
	storiesService.story.EstimatedDurationMinutes = &duration
	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("duration-triggered reconciliation returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 2 || len(calendarService.reconciliations[1].Segments) != 1 {
		t.Fatalf("expected duration update to schedule the retained story: %#v", calendarService.reconciliations)
	}
}

func TestCreateWorkPlanRestoresPriorScheduleWhenAssignmentFails(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	updateErr := errors.New("story assignment failed")
	calendarService := &fakeMayaCalendar{}
	storiesService := &fakeMayaStories{
		story:               stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Failure-safe plan"},
		assignmentUpdateErr: updateErr,
	}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{}, Stories: storiesService,
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: userID}}}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: userID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: 60, AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 2 || len(calendarService.reconciliations[0].Segments) != 1 || len(calendarService.reconciliations[1].Segments) != 0 {
		t.Fatalf("expected desired schedule followed by compensation: %#v", calendarService.reconciliations)
	}
	if storiesService.updatedAssignee != nil {
		t.Fatalf("assignment must remain unchanged after schedule compensation: %v", storiesService.updatedAssignee)
	}
	if len(plan.Actions) != 2 || plan.Actions[0].Status != ActionStatusFailed || plan.Actions[1].Status != ActionStatusFailed {
		t.Fatalf("expected schedule and dependent assignment to fail safely: %#v", plan.Actions)
	}
	if len(calendarService.dispatchedUsers) != 1 || calendarService.dispatchedUsers[0] != userID {
		t.Fatalf("expected compensated outbox state to be dispatched once: %v", calendarService.dispatchedUsers)
	}
}

func TestCreateWorkPlanDoesNotOverwriteConcurrentUserAssignment(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	plannedOwnerID := uuid.New()
	userSelectedOwnerID := uuid.New()
	inspectedAt := time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC)
	concurrentAt := inspectedAt.Add(time.Minute)
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Respect concurrent owner", UpdatedAt: inspectedAt,
	}}
	calendarService := &fakeMayaCalendar{}
	calendarService.onReconcile = func(input calendar.MayaScheduleReconcileInput) {
		if len(input.Segments) == 0 {
			return
		}
		storiesService.story.Assignee = &userSelectedOwnerID
		storiesService.story.UpdatedAt = concurrentAt
	}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{}, Stories: storiesService,
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: plannedOwnerID}}}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: plannedOwnerID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: 60, AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if storiesService.story.Assignee == nil || *storiesService.story.Assignee != userSelectedOwnerID {
		t.Fatalf("Maya must not overwrite the user's concurrent assignment: %v", storiesService.story.Assignee)
	}
	if !storiesService.assignmentExpectedUpdatedAt.Equal(inspectedAt) {
		t.Fatalf("assignment CAS used %v, want inspected story version %v", storiesService.assignmentExpectedUpdatedAt, inspectedAt)
	}
	if len(calendarService.reconciliations) != 2 || len(calendarService.reconciliations[0].Segments) != 1 || len(calendarService.reconciliations[1].Segments) != 0 {
		t.Fatalf("expected schedule commit followed by CAS compensation: %#v", calendarService.reconciliations)
	}
	if len(plan.Actions) != 2 || plan.Actions[0].Status != ActionStatusFailed || plan.Actions[1].Status != ActionStatusFailed {
		t.Fatalf("expected assignment and dependent schedule actions to fail safely: %#v", plan.Actions)
	}
}

func TestCreateWorkPlanRejectsStaleScheduleForExistingAssignee(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	inspectedAt := time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC)
	concurrentAt := inspectedAt.Add(time.Minute)
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Do not apply stale schedule",
		Assignee: &ownerID, UpdatedAt: inspectedAt,
	}}
	calendarService := &fakeMayaCalendar{currentStoryUpdatedAt: &concurrentAt}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{}, Stories: storiesService,
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: ownerID}}}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: 60, AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 0 {
		t.Fatalf("stale schedule must be rejected before local commit: %#v", calendarService.reconciliations)
	}
	if len(plan.Actions) != 1 || plan.Actions[0].Status != ActionStatusFailed || plan.Actions[0].Error == nil {
		t.Fatalf("expected stale schedule action to fail safely: %#v", plan.Actions)
	}
	if storiesService.updatedAssignee != nil {
		t.Fatalf("existing assignment must remain untouched: %v", storiesService.updatedAssignee)
	}
}

func TestCreateWorkPlanAssignsWhenGoogleDispatchFailsAfterDurableOutbox(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	calendarService := &fakeMayaCalendar{dispatchErr: errors.New("google unavailable")}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Durable delivery"}}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{}, Stories: storiesService,
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: userID}}}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: userID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: 60, AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if storiesService.updatedAssignee == nil || *storiesService.updatedAssignee != userID {
		t.Fatalf("expected assignment after durable local schedule commit, got %v", storiesService.updatedAssignee)
	}
	for _, action := range plan.Actions {
		if action.Status != ActionStatusApplied {
			t.Fatalf("expected locally committed action despite provider outage: %#v", action)
		}
	}
}

func TestCreateWorkPlanMovesExistingMayaScheduleBeforeReassignment(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	previousOwnerID := uuid.New()
	newOwnerID := uuid.New()
	storyIDCopy := storyID
	calendarService := &fakeMayaCalendar{
		ownerships:     map[uuid.UUID]bool{previousOwnerID: true},
		schedulingView: calendar.CoreSchedule{Timezone: "UTC"},
		mayaBlocks: []calendar.CoreScheduleBlock{{
			ID: uuid.New(), WorkspaceID: workspaceID, UserID: previousOwnerID, StoryID: &storyIDCopy,
			Title: "Prior segment", StartAt: time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC),
			EndAt: time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC), Source: calendar.ScheduleBlockSourceMaya,
		}},
	}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Move initial plan", Assignee: &previousOwnerID,
	}}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{scheduleOwners: []uuid.UUID{previousOwnerID}}, Stories: storiesService,
		Reports:  &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{Members: []reports.CoreMemberWorkload{{UserID: newOwnerID}}}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: newOwnerID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: 60, AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 2 || calendarService.reconciliations[0].UserID != newOwnerID || calendarService.reconciliations[1].UserID != previousOwnerID || len(calendarService.reconciliations[1].Segments) != 0 {
		t.Fatalf("expected new-owner schedule followed by prior-owner cleanup: %#v", calendarService.reconciliations)
	}
	if storiesService.updatedAssignee == nil || *storiesService.updatedAssignee != newOwnerID {
		t.Fatalf("expected reassignment only after both calendar states committed: %v", storiesService.updatedAssignee)
	}
	for _, action := range plan.Actions {
		if action.Status != ActionStatusApplied {
			t.Fatalf("expected reassignment plan to apply: %#v", plan.Actions)
		}
	}
}

func containsUUID(values []uuid.UUID, target uuid.UUID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func timePointer(value time.Time) *time.Time { return &value }

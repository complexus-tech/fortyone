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
	if len(repo.scheduleOwners) != 1 || repo.scheduleOwners[0] != ownerID || len(calendarService.reconciliations) != 2 || len(calendarService.reconciliations[0].Segments) != 0 || !calendarService.reconciliations[0].KeepOwnership || len(calendarService.reconciliations[1].Segments) != 0 || !calendarService.reconciliations[1].KeepOwnership {
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
	if len(calendarService.reconciliations) != 3 || len(calendarService.reconciliations[2].Segments) == 0 {
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
	duration := 60
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Respect concurrent owner", UpdatedAt: inspectedAt,
		EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
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
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: duration, AutoApply: true,
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

func TestCreateWorkPlanRefreshesOwnershipBeforeProviderUpsert(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	inspectedAt := time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC)
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Provider freshness",
		UpdatedAt: inspectedAt,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	calendarService.providerGate = func(input calendar.MayaScheduleReconcileInput) string {
		if storiesService.story.Assignee != nil && *storiesService.story.Assignee == input.UserID &&
			storiesService.story.AutoSchedulingEnabled && input.ExpectedStoryUpdatedAt != nil &&
			input.ExpectedStoryUpdatedAt.Equal(storiesService.story.UpdatedAt) && len(input.Segments) > 0 {
			return "upsert"
		}
		return "delete"
	}
	service := New(Dependencies{
		Repository: &fakeMayaRepository{}, Stories: storiesService,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: ownerID}},
		}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)

	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: duration, AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	for _, action := range plan.Actions {
		if action.Status != ActionStatusApplied {
			t.Fatalf("expected applied plan, got %#v", plan.Actions)
		}
	}
	if len(calendarService.reconciliations) != 2 {
		t.Fatalf("expected schedule commit and post-assignment refresh: %#v", calendarService.reconciliations)
	}
	refreshed := calendarService.reconciliations[1]
	if refreshed.ExpectedStoryUpdatedAt == nil || !refreshed.ExpectedStoryUpdatedAt.Equal(storiesService.story.UpdatedAt) {
		t.Fatalf("ownership watermark %v does not match assigned story version %v", refreshed.ExpectedStoryUpdatedAt, storiesService.story.UpdatedAt)
	}
	if len(calendarService.providerOperations) != 1 || calendarService.providerOperations[0] != "upsert" {
		t.Fatalf("provider gate must deliver the refreshed schedule as an upsert: %v", calendarService.providerOperations)
	}
}

func TestCreateWorkPlanPersistsShorterDurationAndCapsMinimumFocus(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 120
	minimumFocus := 60
	repo := &fakeMayaRepository{schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Shorter confirmed work",
		Assignee: &ownerID, UpdatedAt: time.Date(2026, 8, 17, 8, 0, 0, 0, time.UTC),
		EstimatedDurationMinutes: &duration, MinimumFocusBlockMinutes: &minimumFocus,
		AutoSchedulingEnabled: true, AutoSchedulingStatus: stories.AutoSchedulingStatusPlanning,
	}}
	calendarService := &fakeMayaCalendar{ownerRepo: repo, schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: ownerID}},
		}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)

	if _, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: 30, AutoApply: true,
	}); err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if storiesService.story.EstimatedDurationMinutes == nil || *storiesService.story.EstimatedDurationMinutes != 30 ||
		storiesService.story.MinimumFocusBlockMinutes == nil || *storiesService.story.MinimumFocusBlockMinutes != 30 {
		t.Fatalf("confirmed duration must persist with a valid minimum focus cap: duration=%v minimum=%v", storiesService.story.EstimatedDurationMinutes, storiesService.story.MinimumFocusBlockMinutes)
	}
	if len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) != 1 ||
		calendarService.reconciliations[0].Segments[0].EndAt.Sub(calendarService.reconciliations[0].Segments[0].StartAt) != 30*time.Minute {
		t.Fatalf("expected one persisted 30-minute schedule: %#v", calendarService.reconciliations)
	}

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("follow-up reconciliation returned error: %v", err)
	}
	if storiesService.story.AutoSchedulingStatus == stories.AutoSchedulingStatusNeedsTime || len(calendarService.reconciliations[len(calendarService.reconciliations)-1].Segments) == 0 {
		t.Fatalf("follow-up reconciliation must retain a schedulable persisted duration: status=%q reconciliations=%#v", storiesService.story.AutoSchedulingStatus, calendarService.reconciliations)
	}
}

func TestCreateWorkPlanRejectsInactiveAutoApplyBeforePreferenceMutation(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	repo := &fakeMayaRepository{storyInactive: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Archived story",
	}}
	calendarService := &fakeMayaCalendar{}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: ownerID}},
		}},
		Calendar: calendarService, Users: &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID}}},
		Planner: NewPlanner(), MayaActorID: uuid.New(),
	})
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)

	_, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: uuid.New(),
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: 60, AutoApply: true,
	})
	if !errors.Is(err, ErrInvalidPlanInput) {
		t.Fatalf("inactive auto-apply error = %v, want %v", err, ErrInvalidPlanInput)
	}
	if storiesService.story.AutoSchedulingEnabled || storiesService.lastUpdates != nil || repo.createRunCalls != 0 || len(calendarService.reconciliations) != 0 {
		t.Fatalf("inactive auto-apply must not mutate preferences or scheduling state: story=%#v updates=%#v runs=%d reconciliations=%#v", storiesService.story, storiesService.lastUpdates, repo.createRunCalls, calendarService.reconciliations)
	}
}

func TestCreateWorkPlanMovesExistingMayaScheduleBeforeReassignment(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	previousOwnerID := uuid.New()
	newOwnerID := uuid.New()
	duration := 60
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
		EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
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
		WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour), DurationMinutes: duration, AutoApply: true,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 3 || calendarService.reconciliations[0].UserID != newOwnerID || calendarService.reconciliations[1].UserID != previousOwnerID || len(calendarService.reconciliations[1].Segments) != 0 || calendarService.reconciliations[2].UserID != newOwnerID || len(calendarService.reconciliations[2].Segments) != 1 {
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

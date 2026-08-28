package maya

import (
	"context"
	"errors"
	"strings"
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
		Assignee: &newOwnerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
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

func TestReconcileSchedulePropagatesAppliedActionPersistenceErrorAfterScheduleCommit(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	terminalizeErr := errors.New("terminalize applied action")
	repo := &fakeMayaRepository{schedulable: true, markActionAppliedErr: terminalizeErr}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Recover applied outcome",
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
	}}
	calendarService := &fakeMayaCalendar{ownerRepo: repo, schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID})

	if !errors.Is(err, terminalizeErr) {
		t.Fatalf("ReconcileSchedule error = %v, want %v", err, terminalizeErr)
	}
	if len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) == 0 {
		t.Fatalf("durable schedule outcome was not preserved: %#v", calendarService.reconciliations)
	}
	if repo.completeRunCalls != 0 || repo.run.Status != RunStatusRunning {
		t.Fatalf("terminalization failure must leave the run recoverable: calls=%d run=%q", repo.completeRunCalls, repo.run.Status)
	}
}

func TestReconcileScheduleJoinsMutationAndFailedActionPersistenceErrors(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	mutationErr := errors.New("calendar mutation failed")
	terminalizeErr := errors.New("terminalize failed action")
	repo := &fakeMayaRepository{schedulable: true, markActionFailedErr: terminalizeErr}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Join persistence failure",
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
	}}
	calendarService := &fakeMayaCalendar{reconcileErr: mutationErr, schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID})

	if !errors.Is(err, mutationErr) || !errors.Is(err, terminalizeErr) {
		t.Fatalf("ReconcileSchedule error = %v, want joined %v and %v", err, mutationErr, terminalizeErr)
	}
	if len(repo.failedActionIDs) == 0 {
		t.Fatal("expected the failed action outcome to be persisted")
	}
	if repo.completeRunCalls != 0 || repo.run.Status != RunStatusRunning {
		t.Fatalf("partial terminalization must leave the run recoverable: calls=%d run=%q", repo.completeRunCalls, repo.run.Status)
	}
}

func TestReconcileScheduleEnrollsEnabledHumanAssignedStoryWithoutOwnership(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 45
	repo := &fakeMayaRepository{schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Enroll direct owner",
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingStatus: stories.AutoSchedulingStatusPlanning,
	}}
	calendarService := &fakeMayaCalendar{ownerRepo: repo, schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 1 {
		t.Fatalf("expected one initial enrollment reconciliation: %#v", calendarService.reconciliations)
	}
	input := calendarService.reconciliations[0]
	if input.UserID != ownerID || len(input.Segments) == 0 || !input.KeepOwnership {
		t.Fatalf("expected enabled direct owner to receive durable schedule ownership: %#v", input)
	}
	if len(storiesService.automationStates) != 1 || storiesService.automationStates[0] != stories.AutoSchedulingStatusScheduled {
		t.Fatalf("expected scheduled state after enrollment: %v", storiesService.automationStates)
	}
}

func TestReconcileScheduleMayaAssigneeCreatesOnlyInitialAssignmentRun(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	mayaActorID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	repo := &fakeMayaRepository{}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Assign once",
		Assignee: &mayaActorID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingStatus: stories.AutoSchedulingStatusNeedsOwner,
	}}
	calendarService := &fakeMayaCalendar{ownerRepo: repo, schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo,
		Stories:    storiesService,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: ownerID}},
		}},
		Calendar:    calendarService,
		Users:       &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID}}},
		Planner:     NewPlanner(),
		MayaActorID: mayaActorID,
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if repo.createRunCalls != 1 {
		t.Fatalf("Maya-assigned intake must create exactly one planning run, got %d", repo.createRunCalls)
	}
	if storiesService.story.Assignee == nil || *storiesService.story.Assignee != ownerID {
		t.Fatalf("expected Maya intake to assign the selected human, got %v", storiesService.story.Assignee)
	}
	if len(calendarService.reconciliations) != 2 {
		t.Fatalf("expected initial schedule plus post-assignment ownership refresh, got %#v", calendarService.reconciliations)
	}
}

func TestReconcileScheduleAccessLossCleansAndTurnsStateOff(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	storyIDCopy := storyID
	repo := &fakeMayaRepository{
		scheduleOwners:        []uuid.UUID{ownerID},
		workspaceAccessDenied: true,
		schedulable:           true,
	}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Entitlement ended",
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingLocked: true, AutoSchedulingStatus: stories.AutoSchedulingStatusLocked,
	}}
	calendarService := &fakeMayaCalendar{
		ownerRepo: repo,
		mayaBlocks: []calendar.CoreScheduleBlock{{
			ID: uuid.New(), WorkspaceID: workspaceID, UserID: ownerID, StoryID: &storyIDCopy,
			Title: "Entitlement ended", StartAt: startAt, EndAt: startAt.Add(time.Hour),
			Source: calendar.ScheduleBlockSourceMaya, IsLocked: true,
		}},
	}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) != 0 || calendarService.reconciliations[0].KeepOwnership {
		t.Fatalf("access loss must retire provider-backed ownership: %#v", calendarService.reconciliations)
	}
	if !storiesService.story.AutoSchedulingEnabled || storiesService.story.AutoSchedulingLocked || storiesService.story.AutoSchedulingStatus != stories.AutoSchedulingStatusOff {
		t.Fatalf("access loss must preserve user intent while pausing state: %#v", storiesService.story)
	}
	if storiesService.story.AutoSchedulingReason == nil || !strings.Contains(*storiesService.story.AutoSchedulingReason, "does not currently have Maya access") {
		t.Fatalf("expected deterministic access reason, got %v", storiesService.story.AutoSchedulingReason)
	}

	repo.workspaceAccessDenied = false
	calendarService.schedulingView.Timezone = "UTC"
	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("restoration reconciliation returned error: %v", err)
	}
	if storiesService.story.AutoSchedulingLocked || storiesService.story.AutoSchedulingStatus != stories.AutoSchedulingStatusScheduled {
		t.Fatalf("restored access must replan the preserved enabled intent: %#v", storiesService.story)
	}
	if len(calendarService.reconciliations) != 2 || len(calendarService.reconciliations[1].Segments) == 0 {
		t.Fatalf("restored access should create a new schedule: %#v", calendarService.reconciliations)
	}
}

func TestReconcileScheduleRestoreSelfHealsStaleLockWithoutBlocks(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	repo := &fakeMayaRepository{schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Restore after archive cascade",
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingLocked: true, AutoSchedulingStatus: stories.AutoSchedulingStatusOff,
	}}
	calendarService := &fakeMayaCalendar{ownerRepo: repo, schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if storiesService.story.AutoSchedulingLocked || storiesService.story.AutoSchedulingStatus != stories.AutoSchedulingStatusScheduled {
		t.Fatalf("restored story must clear its stale lock and schedule in the same run: %#v", storiesService.story)
	}
	if len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) == 0 || calendarService.reconciliations[0].Locked {
		t.Fatalf("expected an unlocked replacement schedule: %#v", calendarService.reconciliations)
	}
}

func TestReconcileScheduleIneligibleLockedOwnerClearsLockBeforeCleanup(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	startAt := time.Now().UTC().Add(24 * time.Hour)
	storyIDCopy := storyID
	block := calendar.CoreScheduleBlock{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: ownerID, StoryID: &storyIDCopy,
		Title: "Owner lost access", StartAt: startAt, EndAt: startAt.Add(time.Hour),
		Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: 0, IsLocked: true,
	}
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{ownerID}, schedulable: false}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: block.Title,
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingLocked: true, AutoSchedulingStatus: stories.AutoSchedulingStatusLocked,
	}}
	calendarService := &fakeMayaCalendar{
		ownerRepo: repo, mayaBlocks: []calendar.CoreScheduleBlock{block},
		schedulingView: calendar.CoreSchedule{Timezone: "UTC"},
	}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if storiesService.story.AutoSchedulingLocked || storiesService.story.AutoSchedulingStatus != stories.AutoSchedulingStatusNeedsOwner {
		t.Fatalf("ineligible locked owner must become unlocked needs-owner work: %#v", storiesService.story)
	}
	if len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) != 0 || calendarService.reconciliations[0].KeepOwnership {
		t.Fatalf("ineligible owner's locked blocks and ownership must be retired: %#v", calendarService.reconciliations)
	}
}

func TestReconcileScheduleKeepsConflictingLockedBlockAndMarksAtRisk(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	storyIDCopy := storyID
	block := calendar.CoreScheduleBlock{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: ownerID, StoryID: &storyIDCopy,
		Title: "Locked focus", StartAt: startAt, EndAt: startAt.Add(time.Hour),
		Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: 0, IsLocked: true, HasConflict: true,
	}
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{ownerID}, schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: block.Title,
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingLocked: true, AutoSchedulingStatus: stories.AutoSchedulingStatusLocked,
	}}
	calendarService := &fakeMayaCalendar{
		ownerRepo: repo, schedulingView: calendar.CoreSchedule{Timezone: "UTC"}, mayaBlocks: []calendar.CoreScheduleBlock{block},
	}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 1 {
		t.Fatalf("expected one exact locked retention write: %#v", calendarService.reconciliations)
	}
	input := calendarService.reconciliations[0]
	if !input.Locked || len(input.Segments) != 1 || !input.Segments[0].StartAt.Equal(startAt) || !input.Segments[0].EndAt.Equal(startAt.Add(time.Hour)) {
		t.Fatalf("conflicting lock must remain fixed until the user unlocks it: %#v", input)
	}
	if storiesService.story.AutoSchedulingStatus != stories.AutoSchedulingStatusAtRisk || storiesService.story.AutoSchedulingReason == nil || !strings.Contains(*storiesService.story.AutoSchedulingReason, "Unlock it so Maya can move the work") {
		t.Fatalf("expected visible locked-conflict risk, got status=%q reason=%v", storiesService.story.AutoSchedulingStatus, storiesService.story.AutoSchedulingReason)
	}
}

func TestReconcileScheduleMovesElapsedUnlockedWorkIntoFuture(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	pastStart := time.Now().UTC().Add(-2 * time.Hour)
	storyIDCopy := storyID
	pastBlock := calendar.CoreScheduleBlock{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: ownerID, StoryID: &storyIDCopy,
		Title: "Incomplete work", StartAt: pastStart, EndAt: pastStart.Add(time.Hour),
		Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: 0,
	}
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{ownerID}, schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: pastBlock.Title,
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingStatus: stories.AutoSchedulingStatusScheduled,
	}}
	calendarService := &fakeMayaCalendar{
		ownerRepo: repo, mayaBlocks: []calendar.CoreScheduleBlock{pastBlock},
		schedulingView: calendar.CoreSchedule{Timezone: "UTC"},
	}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	startedAt := time.Now().UTC()
	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 1 || len(calendarService.reconciliations[0].Segments) == 0 {
		t.Fatalf("elapsed unlocked work must be replanned: %#v", calendarService.reconciliations)
	}
	if !calendarService.reconciliations[0].Segments[0].EndAt.After(startedAt) {
		t.Fatalf("replacement work must end in the future: %#v", calendarService.reconciliations[0].Segments)
	}
}

func TestReconcileScheduleDoesNotDoubleAllocateInProgressWork(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 4 * 60
	activeStart := time.Now().UTC().Add(-30 * time.Minute)
	storyIDCopy := storyID
	activeBlock := calendar.CoreScheduleBlock{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: ownerID, StoryID: &storyIDCopy,
		Title: "Integration check", StartAt: activeStart, EndAt: activeStart.Add(4 * time.Hour),
		Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: 0,
	}
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{ownerID}, schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: activeBlock.Title,
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingStatus: stories.AutoSchedulingStatusScheduled,
	}}
	calendarService := &fakeMayaCalendar{
		ownerRepo: repo,
		schedulingView: calendar.CoreSchedule{
			Timezone: "UTC",
			Blocks:   []calendar.CoreScheduleBlock{activeBlock},
		},
	}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 1 {
		t.Fatalf("expected one reconciliation, got %#v", calendarService.reconciliations)
	}

	totalMinutes := 0
	for _, segment := range calendarService.reconciliations[0].Segments {
		totalMinutes += int(segment.EndAt.Sub(segment.StartAt) / time.Minute)
	}
	if totalMinutes != duration {
		t.Fatalf("in-progress work must count toward the %d-minute estimate, got %d scheduled minutes: %#v", duration, totalMinutes, calendarService.reconciliations[0].Segments)
	}
}

func TestReconcileScheduleKeepsElapsedLockedWorkAndMarksAtRisk(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	pastStart := time.Now().UTC().Add(-2 * time.Hour)
	storyIDCopy := storyID
	pastBlock := calendar.CoreScheduleBlock{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: ownerID, StoryID: &storyIDCopy,
		Title: "Elapsed lock", StartAt: pastStart, EndAt: pastStart.Add(time.Hour),
		Source: calendar.ScheduleBlockSourceMaya, SegmentIndex: 0, IsLocked: true,
	}
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{ownerID}, schedulable: true}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: pastBlock.Title,
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
		AutoSchedulingLocked: true, AutoSchedulingStatus: stories.AutoSchedulingStatusLocked,
	}}
	calendarService := &fakeMayaCalendar{
		ownerRepo: repo, mayaBlocks: []calendar.CoreScheduleBlock{pastBlock},
		schedulingView: calendar.CoreSchedule{Timezone: "UTC"},
	}
	service := New(Dependencies{
		Repository: repo, Stories: storiesService, Reports: &fakeMayaReports{}, Calendar: calendarService,
		Users: &fakeMayaUsers{}, Planner: NewPlanner(), MayaActorID: uuid.New(),
	})

	if err := service.ReconcileSchedule(context.Background(), ReconcileScheduleInput{WorkspaceID: &workspaceID, StoryID: &storyID}); err != nil {
		t.Fatalf("ReconcileSchedule returned error: %v", err)
	}
	if len(calendarService.reconciliations) != 1 || !calendarService.reconciliations[0].Locked || len(calendarService.reconciliations[0].Segments) != 1 ||
		!calendarService.reconciliations[0].Segments[0].StartAt.Equal(pastStart) {
		t.Fatalf("elapsed lock must retain its exact timestamp: %#v", calendarService.reconciliations)
	}
	if storiesService.story.AutoSchedulingStatus != stories.AutoSchedulingStatusAtRisk || storiesService.story.AutoSchedulingReason == nil ||
		!strings.Contains(*storiesService.story.AutoSchedulingReason, "locked work time has passed") {
		t.Fatalf("elapsed lock must surface one stable at-risk outcome: status=%q reason=%v", storiesService.story.AutoSchedulingStatus, storiesService.story.AutoSchedulingReason)
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
		Assignee: &actualOwnerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
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
		Assignee: &ownerID, EstimatedDurationMinutes: &duration, AutoSchedulingEnabled: true,
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
	asOf := startAt.Add(-time.Hour)
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
			Assignee: &ownerID, EstimatedDurationMinutes: &duration, StartDate: &startAt, AutoSchedulingEnabled: true,
		}},
		Reports: &fakeMayaReports{}, Calendar: calendarService, Users: &fakeMayaUsers{},
		Planner: NewPlanner(), Clock: fixedClock{now: asOf}, MayaActorID: uuid.New(),
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
	repo := &fakeMayaRepository{scheduleOwners: []uuid.UUID{ownerID}, schedulable: false, storyInactive: true}
	calendarService := &fakeMayaCalendar{}
	service := New(Dependencies{
		Repository: repo,
		Stories: &fakeMayaStories{story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Completed work",
			Assignee: &ownerID, EstimatedDurationMinutes: &duration, CompletedAt: timePointer(time.Now().UTC()), AutoSchedulingEnabled: true,
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
		storyInactive:             true,
		ownershipRetainable:       false,
	}
	calendarService := &fakeMayaCalendar{ownerRepo: repo, hasOwnership: true}
	service := New(Dependencies{
		Repository: repo,
		Stories: &fakeMayaStories{story: stories.CoreSingleStory{
			ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Retire terminal ownership",
			Assignee: &ownerID, CompletedAt: &completedAt, AutoSchedulingEnabled: true,
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
		Assignee: &ownerID, AutoSchedulingEnabled: true,
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

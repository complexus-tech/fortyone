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
	"github.com/complexus-tech/projects-api/pkg/events"
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

func TestBuildStoryScheduleTransitionUsesMostSignificantChangedSegment(t *testing.T) {
	userID := uuid.New()
	firstStart := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	secondStart := time.Date(2026, 8, 17, 13, 0, 0, 0, time.UTC)
	story := stories.CoreSingleStory{
		AutoSchedulingStatus: stories.AutoSchedulingStatusScheduled,
	}
	previousBlocks := []calendar.CoreScheduleBlock{
		{UserID: userID, SegmentIndex: 0, StartAt: firstStart, EndAt: firstStart.Add(time.Hour)},
		{UserID: userID, SegmentIndex: 1, StartAt: secondStart, EndAt: secondStart.Add(time.Hour)},
	}

	t.Run("later segment changes day while first segment stays fixed", func(t *testing.T) {
		movedStart := secondStart.Add(26 * time.Hour)
		transition := buildStoryScheduleTransition(
			story,
			userID,
			previousBlocks,
			[]calendar.MayaScheduleSegmentInput{
				{SegmentIndex: 0, StartAt: firstStart, EndAt: firstStart.Add(time.Hour)},
				{SegmentIndex: 1, StartAt: movedStart, EndAt: movedStart.Add(time.Hour)},
			},
			"UTC",
			stories.AutoSchedulingStatusScheduled,
			"Maya moved the work.",
		)

		if transition == nil || transition.Kind != events.StoryScheduleTransitionDayChanged {
			t.Fatalf("expected later segment day change, got %#v", transition)
		}
		if transition.PreviousStartAt == nil || !transition.PreviousStartAt.Equal(secondStart) ||
			transition.StartAt == nil || !transition.StartAt.Equal(movedStart) ||
			transition.ShiftMinutes != 26*60 {
			t.Fatalf("transition should describe the changed later segment: %#v", transition)
		}
	})

	t.Run("largest same-day shift wins", func(t *testing.T) {
		firstMovedStart := firstStart.Add(time.Hour)
		secondMovedStart := secondStart.Add(3 * time.Hour)
		transition := buildStoryScheduleTransition(
			story,
			userID,
			previousBlocks,
			[]calendar.MayaScheduleSegmentInput{
				{SegmentIndex: 0, StartAt: firstMovedStart, EndAt: firstMovedStart.Add(time.Hour)},
				{SegmentIndex: 1, StartAt: secondMovedStart, EndAt: secondMovedStart.Add(time.Hour)},
			},
			"UTC",
			stories.AutoSchedulingStatusScheduled,
			"Maya moved the work.",
		)

		if transition == nil || transition.Kind != events.StoryScheduleTransitionMoved {
			t.Fatalf("expected a same-day move, got %#v", transition)
		}
		if transition.PreviousStartAt == nil || !transition.PreviousStartAt.Equal(secondStart) ||
			transition.StartAt == nil || !transition.StartAt.Equal(secondMovedStart) ||
			transition.ShiftMinutes != 3*60 {
			t.Fatalf("transition should describe the largest changed segment: %#v", transition)
		}
	})

	t.Run("split onto another day is always meaningful", func(t *testing.T) {
		oldStart := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
		newDayStart := oldStart.Add(24 * time.Hour)
		transition := buildStoryScheduleTransition(
			story,
			userID,
			[]calendar.CoreScheduleBlock{{
				UserID: userID, SegmentIndex: 0, StartAt: oldStart, EndAt: oldStart.Add(2 * time.Hour),
			}},
			[]calendar.MayaScheduleSegmentInput{
				{SegmentIndex: 0, StartAt: oldStart, EndAt: oldStart.Add(time.Hour)},
				{SegmentIndex: 1, StartAt: newDayStart, EndAt: newDayStart.Add(time.Hour)},
			},
			"UTC",
			stories.AutoSchedulingStatusScheduled,
			"Maya moved the work.",
		)

		if transition == nil || transition.Kind != events.StoryScheduleTransitionDayChanged {
			t.Fatalf("a split onto another date must satisfy the always-meaningful day-change notification contract: %#v", transition)
		}
		if transition.StartAt == nil || !transition.StartAt.Equal(newDayStart) || transition.PreviousLocalDate == transition.LocalDate {
			t.Fatalf("transition should describe the newly scheduled day: %#v", transition)
		}
	})

	t.Run("material end-only change is moved", func(t *testing.T) {
		oldStart := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
		transition := buildStoryScheduleTransition(
			story,
			userID,
			[]calendar.CoreScheduleBlock{{
				UserID: userID, SegmentIndex: 0, StartAt: oldStart, EndAt: oldStart.Add(2 * time.Hour),
			}},
			[]calendar.MayaScheduleSegmentInput{{
				SegmentIndex: 0, StartAt: oldStart, EndAt: oldStart.Add(time.Hour),
			}},
			"UTC",
			stories.AutoSchedulingStatusScheduled,
			"Maya changed the work block.",
		)

		if transition == nil || transition.Kind != events.StoryScheduleTransitionMoved || transition.ShiftMinutes != -60 {
			t.Fatalf("a material end-only change must meet the 60-minute notification threshold: %#v", transition)
		}
	})
}

func TestBuildStoryScheduleTransitionDoesNotTreatPlanningReplanAsFirstSchedule(t *testing.T) {
	userID := uuid.New()
	startAt := time.Date(2026, 8, 17, 9, 0, 0, 0, time.UTC)
	transition := buildStoryScheduleTransition(
		stories.CoreSingleStory{AutoSchedulingStatus: stories.AutoSchedulingStatusPlanning},
		userID,
		[]calendar.CoreScheduleBlock{{UserID: userID, SegmentIndex: 0, StartAt: startAt, EndAt: startAt.Add(time.Hour)}},
		[]calendar.MayaScheduleSegmentInput{{SegmentIndex: 0, StartAt: startAt, EndAt: startAt.Add(time.Hour)}},
		"UTC",
		stories.AutoSchedulingStatusScheduled,
		"Maya kept the existing schedule.",
	)

	if transition == nil || transition.Kind != events.StoryScheduleTransitionStateChanged {
		t.Fatalf("an existing schedule leaving planning must be a state change, not first schedule: %#v", transition)
	}
}

func TestRefineScheduleOutcomeReasonKeepsFirstScheduleCopy(t *testing.T) {
	fallback := "Maya scheduled this story around the assignee's availability."
	startAt := time.Now().UTC().Add(24 * time.Hour)
	reason := refineScheduleOutcomeReason(
		nil,
		[]calendar.MayaScheduleSegmentInput{{SegmentIndex: 0, StartAt: startAt, EndAt: startAt.Add(time.Hour)}},
		stories.AutoSchedulingStatusScheduled,
		fallback,
	)
	if reason != fallback || strings.Contains(strings.ToLower(reason), "moved") || strings.Contains(strings.ToLower(reason), "changed") {
		t.Fatalf("first schedule must keep first-placement copy, got %q", reason)
	}
}

func TestLockedScheduleRisk(t *testing.T) {
	baseStart := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	storyID := uuid.New()
	ownerID := uuid.New()
	block := func(index int, start time.Time, minutes int) calendar.CoreScheduleBlock {
		return calendar.CoreScheduleBlock{
			ID: uuid.New(), UserID: ownerID, StoryID: &storyID, SegmentIndex: index,
			StartAt: start, EndAt: start.Add(time.Duration(minutes) * time.Minute), IsLocked: true,
		}
	}
	intPointer := func(value int) *int { return &value }
	timePointer := func(value time.Time) *time.Time { return &value }

	tests := []struct {
		name       string
		story      stories.CoreSingleStory
		blocks     []calendar.CoreScheduleBlock
		wantRisk   bool
		reasonPart string
	}{
		{
			name: "valid", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(60)},
			blocks: []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
		},
		{
			name: "busy conflict", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(60)},
			blocks: func() []calendar.CoreScheduleBlock {
				value := block(0, baseStart, 60)
				value.HasConflict = true
				return []calendar.CoreScheduleBlock{value}
			}(),
			wantRisk: true, reasonPart: "conflicts with another calendar event",
		},
		{
			name: "elapsed", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(60)},
			blocks:   []calendar.CoreScheduleBlock{block(0, time.Now().UTC().Add(-2*time.Hour), 60)},
			wantRisk: true, reasonPart: "locked work time has passed",
		},
		{
			name: "under allocated", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(90)},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "reserves 60 minutes",
		},
		{
			name: "over allocated", story: stories.CoreSingleStory{EstimatedDurationMinutes: intPointer(30)},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "now needs 30 minutes",
		},
		{
			name: "minimum focus violation", story: stories.CoreSingleStory{
				EstimatedDurationMinutes: intPointer(60), MinimumFocusBlockMinutes: intPointer(45),
			},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 30), block(1, baseStart.Add(time.Hour), 30)},
			wantRisk: true, reasonPart: "45-minute minimum focus block",
		},
		{
			name: "before story start", story: stories.CoreSingleStory{
				EstimatedDurationMinutes: intPointer(60), StartDate: timePointer(baseStart.Add(time.Hour)),
			},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "before this story's start date",
		},
		{
			name: "after story deadline", story: stories.CoreSingleStory{
				EstimatedDurationMinutes: intPointer(60), EndDate: timePointer(baseStart.Add(-48 * time.Hour)),
			},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "after this story's deadline",
		},
		{
			name: "outside sprint", story: stories.CoreSingleStory{
				EstimatedDurationMinutes: intPointer(60), SprintSummary: &stories.CoreSprintSummary{
					StartDate: baseStart.Add(24 * time.Hour), EndDate: baseStart.Add(48 * time.Hour),
				},
			},
			blocks:   []calendar.CoreScheduleBlock{block(0, baseStart, 60)},
			wantRisk: true, reasonPart: "outside this story's sprint",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reason, atRisk := lockedScheduleRisk(test.story, test.blocks, "UTC")
			if atRisk != test.wantRisk {
				t.Fatalf("lockedScheduleRisk() risk = %v, want %v; reason=%q", atRisk, test.wantRisk, reason)
			}
			if test.reasonPart != "" && !strings.Contains(reason, test.reasonPart) {
				t.Fatalf("lockedScheduleRisk() reason = %q, want substring %q", reason, test.reasonPart)
			}
		})
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

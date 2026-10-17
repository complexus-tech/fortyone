package maya

import (
	"context"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

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

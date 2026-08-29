package maya

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

func TestCreateWorkPlanPropagatesAppliedOutcomePersistenceConflict(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	requestedBy := uuid.New()
	mayaActorID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	startAt := time.Date(2026, 8, 28, 9, 0, 0, 0, time.UTC)
	repo := &fakeMayaRepository{markActionAppliedErr: mayadomain.ErrActionNotProposed}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID:                       storyID,
		Workspace:                workspaceID,
		Team:                     uuid.New(),
		Title:                    "Preserve committed outcome",
		EstimatedDurationMinutes: &duration,
		AutoSchedulingEnabled:    true,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo,
		Stories:    storiesService,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: ownerID, FullName: "Ada"}},
		}},
		Calendar:    calendarService,
		Users:       &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID, FullName: "Ada"}}},
		Planner:     NewPlanner(),
		Clock:       fixedClock{now: startAt.Add(-time.Hour)},
		MayaActorID: mayaActorID,
	})

	plan, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID:     workspaceID,
		StoryID:         storyID,
		TriggeredBy:     requestedBy,
		Trigger:         RunTriggerManual,
		WindowStart:     startAt,
		WindowEnd:       startAt.Add(8 * time.Hour),
		DurationMinutes: duration,
		AutoApply:       true,
	})

	if !errors.Is(err, mayadomain.ErrActionNotProposed) {
		t.Fatalf("CreateWorkPlan error = %v, want %v", err, mayadomain.ErrActionNotProposed)
	}
	if repo.completeRunCalls != 0 || plan.Run.Status != RunStatusRunning {
		t.Fatalf("terminalization conflict must leave the run recoverable: calls=%d run=%q", repo.completeRunCalls, plan.Run.Status)
	}
	if len(plan.Actions) == 0 {
		t.Fatal("expected committed actions to be returned with the persistence error")
	}
	for _, action := range plan.Actions {
		if action.Status != ActionStatusApplied {
			t.Fatalf("committed action %s reported as %q", action.ID, action.Status)
		}
	}
	if storiesService.updatedAssignee == nil || *storiesService.updatedAssignee != ownerID {
		t.Fatalf("committed assignment was lost: %v", storiesService.updatedAssignee)
	}
	if len(calendarService.reconciliations) == 0 || len(calendarService.reconciliations[0].Segments) == 0 {
		t.Fatalf("committed schedule was lost: %#v", calendarService.reconciliations)
	}
}

func TestApplyWorkPlanReturnsCommittedOutcomesWithPersistenceConflict(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	requestedBy := uuid.New()
	mayaActorID := uuid.New()
	ownerID := uuid.New()
	duration := 60
	storyUpdatedAt := time.Date(2026, 8, 28, 7, 0, 0, 0, time.UTC)
	startAt := storyUpdatedAt.Add(2 * time.Hour)
	repo := &fakeMayaRepository{}
	storiesService := &fakeMayaStories{story: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: uuid.New(), Title: "Apply stored outcome", UpdatedAt: storyUpdatedAt,
	}}
	calendarService := &fakeMayaCalendar{schedulingView: calendar.CoreSchedule{Timezone: "UTC"}}
	service := New(Dependencies{
		Repository: repo,
		Stories:    storiesService,
		Reports: &fakeMayaReports{analysis: reports.CoreWorkloadAnalysis{
			Members: []reports.CoreMemberWorkload{{UserID: ownerID, FullName: "Ada"}},
		}},
		Calendar:    calendarService,
		Users:       &fakeMayaUsers{members: []users.CoreUser{{ID: ownerID, FullName: "Ada"}}},
		Planner:     NewPlanner(),
		Clock:       fixedClock{now: storyUpdatedAt},
		MayaActorID: mayaActorID,
	})
	preview, err := service.CreateWorkPlan(context.Background(), CreateWorkPlanInput{
		WorkspaceID: workspaceID, StoryID: storyID, TriggeredBy: requestedBy,
		Trigger: RunTriggerManual, WindowStart: startAt, WindowEnd: startAt.Add(8 * time.Hour),
		DurationMinutes: duration,
	})
	if err != nil {
		t.Fatalf("CreateWorkPlan preview returned error: %v", err)
	}
	repo.markActionAppliedErr = mayadomain.ErrActionNotProposed

	applied, err := service.ApplyWorkPlan(context.Background(), ApplyWorkPlanInput{
		RunID: preview.Run.ID, WorkspaceID: workspaceID, TriggeredBy: requestedBy,
	})

	if !errors.Is(err, mayadomain.ErrActionNotProposed) {
		t.Fatalf("ApplyWorkPlan error = %v, want %v", err, mayadomain.ErrActionNotProposed)
	}
	if len(applied.Actions) == 0 {
		t.Fatal("expected committed stored-plan actions with the persistence error")
	}
	for _, action := range applied.Actions {
		if action.Status != ActionStatusApplied {
			t.Fatalf("committed stored-plan action %s reported as %q", action.ID, action.Status)
		}
	}
	if storiesService.updatedAssignee == nil || *storiesService.updatedAssignee != ownerID || len(calendarService.reconciliations) == 0 {
		t.Fatalf("stored-plan mutation was not preserved: assignee=%v reconciliations=%#v", storiesService.updatedAssignee, calendarService.reconciliations)
	}
}

func TestApplyActionsPropagatesFailedOutcomePersistenceConflict(t *testing.T) {
	repo := &fakeMayaRepository{markActionFailedErr: mayadomain.ErrActionNotProposed}
	service := New(Dependencies{Repository: repo, Clock: fixedClock{now: time.Now().UTC()}})
	actionID := uuid.New()

	actions, err := service.applyActions(context.Background(), []CoreAction{{
		ID:     actionID,
		Type:   ActionType("unsupported"),
		Status: ActionStatusProposed,
	}})

	if !errors.Is(err, mayadomain.ErrActionNotProposed) {
		t.Fatalf("applyActions error = %v, want %v", err, mayadomain.ErrActionNotProposed)
	}
	if len(actions) != 1 || actions[0].Status != ActionStatusFailed || actions[0].Error == nil ||
		!strings.Contains(*actions[0].Error, "unsupported maya action type") {
		t.Fatalf("failed mutation outcome was not preserved: %#v", actions)
	}
	if len(repo.failedActionIDs) != 1 || repo.failedActionIDs[0] != actionID {
		t.Fatalf("failed action persistence attempts = %v, want [%s]", repo.failedActionIDs, actionID)
	}
}

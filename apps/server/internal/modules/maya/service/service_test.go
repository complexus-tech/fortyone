package maya

import (
	"context"
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

	repo := &fakeMayaRepository{}
	storiesSvc := &fakeMayaStories{
		story: stories.CoreSingleStory{
			ID:        storyID,
			Workspace: workspaceID,
			Team:      teamID,
			Title:     "Schedule me",
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
		DurationMinutes: 60,
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
	expectedScheduledEndAt := startAt.Add(60 * time.Minute)
	if storiesSvc.updatedStartDate == nil || !storiesSvc.updatedStartDate.Equal(startAt) {
		t.Fatalf("expected story start date update to %s, got %v", startAt, storiesSvc.updatedStartDate)
	}
	if storiesSvc.updatedEndDate == nil || !storiesSvc.updatedEndDate.Equal(expectedScheduledEndAt) {
		t.Fatalf("expected story end date update to %s, got %v", expectedScheduledEndAt, storiesSvc.updatedEndDate)
	}
	if len(repo.appliedActionIDs) != 2 {
		t.Fatalf("expected two applied action marks, got %d", len(repo.appliedActionIDs))
	}
	if len(storiesSvc.updateReasons) != 2 {
		t.Fatalf("expected two reason-aware story updates, got %d", len(storiesSvc.updateReasons))
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

func TestCreateWorkPlanOverwritesStoryDatesWithPlannedWorkWindow(t *testing.T) {
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
	expectedStartDate := windowStart
	expectedEndDate := windowStart.Add(60 * time.Minute)

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

	if value, ok := storiesSvc.lastUpdates["start_date"].(time.Time); !ok || !value.Equal(expectedStartDate) {
		t.Fatalf("expected start date to be overwritten with %s, got %#v", expectedStartDate, storiesSvc.lastUpdates["start_date"])
	}
	if value, ok := storiesSvc.lastUpdates["end_date"].(time.Time); !ok || !value.Equal(expectedEndDate) {
		t.Fatalf("expected end date to be overwritten with %s, got %#v", expectedEndDate, storiesSvc.lastUpdates["end_date"])
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
	actions          []CoreAction
	appliedActionIDs []uuid.UUID
}

func (f *fakeMayaRepository) CreateRun(_ context.Context, input CreateRunInput) (CoreRun, error) {
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

type fakeMayaStories struct {
	story            stories.CoreSingleStory
	actorID          uuid.UUID
	updatedAssignee  *uuid.UUID
	updatedStartDate *time.Time
	updatedEndDate   *time.Time
	lastUpdates      map[string]any
	updateReasons    []string
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
	if value, ok := updates["assignee_id"].(uuid.UUID); ok {
		f.updatedAssignee = &value
	}
	if value, ok := updates["start_date"].(time.Time); ok {
		f.updatedStartDate = &value
	}
	if value, ok := updates["end_date"].(time.Time); ok {
		f.updatedEndDate = &value
	}
	return nil
}

type fakeMayaReports struct {
	analysis reports.CoreWorkloadAnalysis
}

func (f *fakeMayaReports) GetWorkloadAnalysis(_ context.Context, _ uuid.UUID, _ reports.ReportFilters) (reports.CoreWorkloadAnalysis, error) {
	return f.analysis, nil
}

type fakeMayaCalendar struct {
	createdBlock      calendar.CoreScheduleBlockInput
	listScheduleCalls int
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

//go:build integration

package reportsrepository

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

func TestTaskReportSectionsApplyEverySelectedItemFilter(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	workspaceID := insertReportTestWorkspace(t, ctx, postgres.Pool, "task-filter-scope")
	actorID := insertReportTestUser(t, ctx, postgres.Pool, "selected-assignee", true)
	otherActorID := insertReportTestUser(t, ctx, postgres.Pool, "other-assignee", true)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceID, actorID)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceID, otherActorID)
	teamID := insertReportTestTeam(t, ctx, postgres.Pool, workspaceID, "task-filter-scope")
	insertReportTestTeamMember(t, ctx, postgres.Pool, teamID, actorID)
	insertReportTestTeamMember(t, ctx, postgres.Pool, teamID, otherActorID)
	startedID := insertReportTestStatus(t, ctx, postgres.Pool, workspaceID, teamID, "Started", "started")
	doneID := insertReportTestStatus(t, ctx, postgres.Pool, workspaceID, teamID, "Done", "completed")
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 7)
	createdAt := startDate.AddDate(0, 0, 1)
	completedAt := startDate.AddDate(0, 0, 3)
	oldCreatedAt := startDate.AddDate(0, 0, -60)
	oldCompletedAt := startDate.AddDate(0, 0, -30)
	objectiveIDs := []uuid.UUID{uuid.New(), uuid.New()}
	sprintIDs := []uuid.UUID{uuid.New(), uuid.New()}
	for index := range objectiveIDs {
		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO objectives (objective_id, sequence_id, name, team_id, workspace_id)
			VALUES ($1, $2, $3, $4, $5)
		`, objectiveIDs[index], index+1, uuid.NewString(), teamID, workspaceID)
		if err != nil {
			t.Fatalf("insert filter objective: %v", err)
		}
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO sprints (sprint_id, name, team_id, workspace_id, start_date, end_date)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, sprintIDs[index], uuid.NewString(), teamID, workspaceID, startDate, endDate)
		if err != nil {
			t.Fatalf("insert filter sprint: %v", err)
		}
	}
	sequence := 0
	insertStory := func(assignee, objective, sprint, status uuid.UUID, created time.Time, completed *time.Time) {
		t.Helper()
		sequence++
		storyID := uuid.New()
		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO stories (id, sequence_id, title, workspace_id, team_id, status_id,
				assignee_id, reporter_id, objective_id, sprint_id, priority,
				created_at, updated_at, completed_at)
			VALUES ($1, $2, 'Filtered report task', $3, $4, $5, $6, $7, $8, $9, 'High', $10, $11, $12)
		`, storyID, sequence, workspaceID, teamID, status, assignee, actorID, objective, sprint,
			created, completedAt, completed)
		if err != nil {
			t.Fatalf("insert filtered task: %v", err)
		}
		if completed == nil {
			return
		}
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO story_activities (story_id, activity_type, field_changed, current_value,
				user_id, workspace_id, new_value, created_at)
			VALUES ($1, 'update', 'status_id', $2, $3, $4, to_jsonb(CAST($2 AS text)), $5)
		`, storyID, startedID.String(), actorID, workspaceID, created.AddDate(0, 0, 1))
		if err != nil {
			t.Fatalf("insert filtered task start history: %v", err)
		}
	}
	insertStory(actorID, objectiveIDs[0], sprintIDs[0], doneID, createdAt, &completedAt)
	insertStory(actorID, objectiveIDs[0], sprintIDs[0], startedID, createdAt, nil)
	insertStory(otherActorID, objectiveIDs[0], sprintIDs[0], doneID, createdAt, &completedAt)
	insertStory(actorID, objectiveIDs[1], sprintIDs[0], doneID, createdAt, &completedAt)
	insertStory(actorID, objectiveIDs[0], sprintIDs[1], doneID, createdAt, &completedAt)
	insertStory(actorID, objectiveIDs[0], sprintIDs[0], doneID, oldCreatedAt, &completedAt)
	insertStory(actorID, objectiveIDs[0], sprintIDs[0], doneID, oldCreatedAt, &oldCompletedAt)
	repository := New(logger.NewWithText(io.Discard, slog.LevelError, "task-filter-scope"), postgres.Pool)
	for _, test := range []struct {
		name                     string
		assignees, goals, cycles []uuid.UUID
		created, cohortDone      int
		completed                int
	}{
		{name: "all scoped tasks", created: 5, cohortDone: 4, completed: 5},
		{name: "assignee", assignees: []uuid.UUID{actorID}, created: 4, cohortDone: 3, completed: 4},
		{name: "objective", goals: []uuid.UUID{objectiveIDs[0]}, created: 4, cohortDone: 3, completed: 4},
		{name: "sprint", cycles: []uuid.UUID{sprintIDs[0]}, created: 4, cohortDone: 3, completed: 4},
		{name: "combined", assignees: []uuid.UUID{actorID}, goals: []uuid.UUID{objectiveIDs[0]}, cycles: []uuid.UUID{sprintIDs[0]}, created: 2, cohortDone: 1, completed: 2},
		{name: "no matching assignee", assignees: []uuid.UUID{uuid.New()}},
	} {
		t.Run(test.name, func(t *testing.T) {
			filters := reportdomain.ReportFilters{
				ActorID: actorID, TeamIDs: []uuid.UUID{teamID}, AssigneeIDs: test.assignees,
				ObjectiveIDs: test.goals, SprintIDs: test.cycles, StartDate: &startDate, EndDate: &endDate,
			}
			overview, err := repository.GetWorkspaceOverview(ctx, workspaceID, filters)
			if err != nil {
				t.Fatalf("get filtered overview: %v", err)
			}
			if overview.Metrics.TotalStories != test.created || overview.Metrics.CompletedStories != test.cohortDone || overview.Metrics.CompletedInPeriod != test.completed {
				t.Errorf("overview created/cohort done/period done = %d/%d/%d, want %d/%d/%d", overview.Metrics.TotalStories, overview.Metrics.CompletedStories, overview.Metrics.CompletedInPeriod, test.created, test.cohortDone, test.completed)
			}
			var created, completed, velocity int
			for _, point := range overview.CompletionTrend {
				created += point.Total
				completed += point.Completed
			}
			for _, point := range overview.VelocityTrend {
				velocity += point.Velocity
			}
			if created != test.created || completed != test.completed || velocity != test.completed {
				t.Errorf("overview trend created/completed/velocity = %d/%d/%d, want %d/%d/%d", created, completed, velocity, test.created, test.completed, test.completed)
			}
			stories, err := repository.GetStoryAnalytics(ctx, workspaceID, filters)
			if err != nil {
				t.Fatalf("get filtered task distributions: %v", err)
			}
			var statuses, priorities, teamCreated, teamDone int
			for _, point := range stories.StatusBreakdown {
				statuses += point.Count
			}
			for _, point := range stories.PriorityDistribution {
				priorities += point.Count
			}
			for _, point := range stories.CompletionByTeam {
				teamCreated += point.Total
				teamDone += point.Completed
			}
			if statuses != test.created || priorities != test.created || teamCreated != test.created || teamDone != test.cohortDone {
				t.Errorf("task distribution status/priority/team created/team done = %d/%d/%d/%d, want %d/%d/%d/%d", statuses, priorities, teamCreated, teamDone, test.created, test.created, test.created, test.cohortDone)
			}
			trends, err := repository.GetTimelineTrends(ctx, workspaceID, filters)
			if err != nil {
				t.Fatalf("get filtered delivery trends: %v", err)
			}
			created, completed, velocity = 0, 0, 0
			var cycleSamples int
			var intake float64
			for _, point := range trends.StoryCompletion {
				created += point.Created
				completed += point.Completed
			}
			for _, point := range trends.TeamVelocity {
				velocity += point.Velocity
			}
			for _, point := range trends.KeyMetricsTrend {
				cycleSamples += point.CycleTimeSamples
				intake += point.StoriesPerDay
			}
			if created != test.created || completed != test.completed || velocity != test.completed || cycleSamples != test.completed || intake != float64(test.created) {
				t.Errorf("delivery created/completed/velocity/cycle samples/intake = %d/%d/%d/%d/%v, want %d/%d/%d/%d/%d", created, completed, velocity, cycleSamples, intake, test.created, test.completed, test.completed, test.completed, test.created)
			}
			teams, err := repository.GetTeamPerformance(ctx, workspaceID, filters)
			if err != nil {
				t.Fatalf("get filtered team performance: %v", err)
			}
			var teamAssigned, memberAssigned, trendAssigned int
			for _, point := range teams.TeamWorkload {
				teamAssigned += point.Assigned
			}
			for _, point := range teams.MemberContributions {
				memberAssigned += point.Assigned
			}
			for _, point := range teams.WorkloadTrend {
				trendAssigned += point.Assigned
			}
			if teamAssigned != test.created || memberAssigned != test.created || trendAssigned != test.created {
				t.Errorf("team/member/trend assigned = %d/%d/%d, want %d each", teamAssigned, memberAssigned, trendAssigned, test.created)
			}
		})
	}
}

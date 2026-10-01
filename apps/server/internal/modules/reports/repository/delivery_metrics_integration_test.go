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

func TestDeliveryMetricsUseCompletionHistoryAndCanonicalBlockers(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	workspaceID := insertReportTestWorkspace(t, ctx, postgres.Pool, "delivery")
	actorID := insertReportTestUser(t, ctx, postgres.Pool, "delivery", true)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceID, actorID)
	teamID := insertReportTestTeam(t, ctx, postgres.Pool, workspaceID, "delivery")
	insertReportTestTeamMember(t, ctx, postgres.Pool, teamID, actorID)
	startedID := insertReportTestStatus(t, ctx, postgres.Pool, workspaceID, teamID, "Active", "started")
	doneID := insertReportTestStatus(t, ctx, postgres.Pool, workspaceID, teamID, "Done", "completed")
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 7)
	completedAt := startDate.AddDate(0, 0, 3)
	createdAt := startDate.AddDate(0, 0, 2)

	sequenceID := 0
	insertStory := func(statusID uuid.UUID, created time.Time, completed *time.Time) uuid.UUID {
		t.Helper()
		id := uuid.New()
		sequenceID++
		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO stories (id, sequence_id, team_id, title, status_id, assignee_id, reporter_id,
				priority, workspace_id, created_at, updated_at, completed_at)
			VALUES ($1, $9, $2, 'Delivery metric test', $3, $4, $4, 'High', $5, $6, $7, $8)
		`, id, teamID, statusID, actorID, workspaceID, created, endDate.AddDate(0, 0, 30), completed, sequenceID)
		if err != nil {
			t.Fatalf("insert metric story: %v", err)
		}
		return id
	}
	completedID := insertStory(doneID, startDate.AddDate(0, -1, 0), &completedAt)
	insertStory(doneID, startDate.AddDate(0, -1, 1), &completedAt) // No start history: excluded from cycle sample.
	insertStory(startedID, createdAt, nil)
	blockerID := insertStory(startedID, createdAt, nil)
	blockedID := insertStory(startedID, createdAt, nil)
	_, err := postgres.Pool.Exec(ctx, `
		INSERT INTO story_activities (story_id, activity_type, field_changed, current_value,
			user_id, workspace_id, new_value, created_at)
		VALUES ($1, 'update', 'status_id', $2, $3, $4, to_jsonb(CAST($2 AS text)), $5)
	`, completedID, startedID.String(), actorID, workspaceID, startDate.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("insert start transition: %v", err)
	}
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO story_associations (from_story_id, to_story_id, association_type, workspace_id)
		VALUES ($1, $2, 'blocking', $3)
	`, blockerID, blockedID, workspaceID)
	if err != nil {
		t.Fatalf("insert canonical blocking relationship: %v", err)
	}
	repository := New(logger.NewWithText(io.Discard, slog.LevelError, "delivery-metrics"), postgres.Pool)
	filters := reportdomain.ReportFilters{ActorID: actorID, TeamIDs: []uuid.UUID{teamID}, StartDate: &startDate, EndDate: &endDate}
	trends, err := repository.GetTimelineTrends(ctx, workspaceID, filters)
	if err != nil {
		t.Fatalf("get timeline: %v", err)
	}
	var created, completed, velocity, samples int
	for _, point := range trends.StoryCompletion {
		created += point.Created
		completed += point.Completed
	}
	for _, point := range trends.TeamVelocity {
		velocity += point.Velocity
	}
	for _, point := range trends.KeyMetricsTrend {
		samples += point.CycleTimeSamples
		if point.CycleTimeSamples > 0 && point.AvgCycleTime != 2 {
			t.Fatalf("cycle time = %v, want 2 days from start to completion", point.AvgCycleTime)
		}
	}
	if created != 3 || completed != 2 || velocity != 2 || samples != 1 {
		t.Fatalf("created/completed/velocity/samples = %d/%d/%d/%d, want 3/2/2/1", created, completed, velocity, samples)
	}
	overview, err := repository.GetWorkspaceOverview(ctx, workspaceID, filters)
	if err != nil {
		t.Fatalf("get overview: %v", err)
	}
	if overview.Metrics.CompletedInPeriod != 2 || overview.Metrics.TotalStories != 3 || overview.Metrics.CompletedStories != 0 {
		t.Fatalf("period completions must be separate from creation cohort: %#v", overview.Metrics)
	}
	var weeklyCreated, weeklyCompleted int
	for _, point := range overview.CompletionTrend {
		weeklyCreated += point.Total
		weeklyCompleted += point.Completed
	}
	if weeklyCreated != 3 || weeklyCompleted != 2 {
		t.Fatalf("weekly created/completed = %d/%d, want 3/2", weeklyCreated, weeklyCompleted)
	}
	health, err := repository.GetPulseStoryHealth(ctx, workspaceID, filters)
	if err != nil || health.BlockedStories != 1 {
		t.Fatalf("modern relationship blocking = %#v, error %v", health, err)
	}
	_, err = postgres.Pool.Exec(ctx, `UPDATE stories SET status_id = $1, completed_at = $2 WHERE id = $3`, doneID, completedAt, blockerID)
	if err != nil {
		t.Fatalf("complete blocker: %v", err)
	}
	health, err = repository.GetPulseStoryHealth(ctx, workspaceID, filters)
	if err != nil || health.BlockedStories != 0 {
		t.Fatalf("completed blockers must not count: %#v, error %v", health, err)
	}
}

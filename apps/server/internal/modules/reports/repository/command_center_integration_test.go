//go:build integration

package reportsrepository

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportservice "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

func TestCommandCenterIncludesOldOpenWorkWithoutChangingDeliveryPeriod(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	workspaceID := insertReportTestWorkspace(t, ctx, postgres.Pool, "current-risk")
	actorID := insertReportTestUser(t, ctx, postgres.Pool, "current-risk", true)
	otherActorID := insertReportTestUser(t, ctx, postgres.Pool, "other-assignee", true)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceID, actorID)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceID, otherActorID)
	teamID := insertReportTestTeam(t, ctx, postgres.Pool, workspaceID, "current-risk")
	otherTeamID := insertReportTestTeam(t, ctx, postgres.Pool, workspaceID, "other-risk")
	insertReportTestTeamMember(t, ctx, postgres.Pool, teamID, actorID)
	insertReportTestTeamMember(t, ctx, postgres.Pool, teamID, otherActorID)
	insertReportTestTeamMember(t, ctx, postgres.Pool, otherTeamID, actorID)
	startedID := insertReportTestStatus(t, ctx, postgres.Pool, workspaceID, teamID, "Started", "started")
	doneID := insertReportTestStatus(t, ctx, postgres.Pool, workspaceID, teamID, "Done", "completed")
	cancelledID := insertReportTestStatus(t, ctx, postgres.Pool, workspaceID, teamID, "Cancelled", "cancelled")
	otherStartedID := insertReportTestStatus(t, ctx, postgres.Pool, workspaceID, otherTeamID, "Started", "started")

	today := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := today.AddDate(0, 0, -7)
	endDate := today.AddDate(0, 0, 1)
	oldCreated := startDate.AddDate(0, 0, -60)
	recentCreated := startDate.AddDate(0, 0, 1)
	oldCompletion := startDate.AddDate(0, 0, -30)
	recentCompletion := startDate.AddDate(0, 0, 2)
	deadline := today.AddDate(0, 0, -1)

	sequence := 0
	insertStory := func(team, status, assignee uuid.UUID, created time.Time, completed *time.Time) {
		t.Helper()
		sequence++
		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO stories (id, sequence_id, team_id, title, status_id, assignee_id,
				reporter_id, priority, workspace_id, estimate_unit, created_at, updated_at,
				completed_at, end_date)
			VALUES ($1, $2, $3, 'Command center scope', $4, $5, $6, 'High', $7, 3, $8, $8, $9, $10)
		`, uuid.New(), sequence, team, status, assignee, actorID, workspaceID, created, completed, deadline)
		if err != nil {
			t.Fatalf("insert command center story: %v", err)
		}
	}
	insertStory(teamID, startedID, actorID, oldCreated, nil)
	insertStory(teamID, startedID, actorID, recentCreated, nil)
	insertStory(teamID, doneID, actorID, oldCreated, &oldCompletion)
	insertStory(teamID, cancelledID, actorID, oldCreated, nil)
	insertStory(teamID, doneID, actorID, oldCreated, &recentCompletion)
	insertStory(teamID, startedID, otherActorID, oldCreated, nil)
	insertStory(otherTeamID, otherStartedID, actorID, oldCreated, nil)

	log := logger.NewWithText(io.Discard, slog.LevelError, "command-center-integration")
	service := reportservice.New(log, New(log, postgres.Pool))
	report, err := service.GetWorkspaceCommandCenterReport(platformauth.SetUserID(ctx, actorID), workspaceID, reportdomain.ReportFilters{
		TeamIDs: []uuid.UUID{teamID}, AssigneeIDs: []uuid.UUID{actorID}, StartDate: &startDate, EndDate: &endDate,
	})
	if err != nil {
		t.Fatalf("get command center: %v", err)
	}
	if len(report.SectionErrors) != 0 {
		t.Fatalf("command center sections failed: %#v", report.SectionErrors)
	}
	if report.Workload.Summary.TotalOpenStories != 2 || report.Workload.Summary.OverdueStories != 2 || report.Workload.Summary.TotalEstimate != 6 {
		t.Fatalf("current workload must include old backlog and exclude closed/other-scope work: %#v", report.Workload.Summary)
	}
	if report.Pulse.Summary.OpenStories != 2 || report.Pulse.Summary.OverdueStories != 2 || report.Pulse.Stories.OpenStories != 2 {
		t.Fatalf("current pulse must agree with operational workload: %#v", report.Pulse)
	}
	if report.Overview.Metrics.TotalStories != 1 || report.Overview.Metrics.CompletedInPeriod != 1 || report.Overview.Metrics.CompletedStories != 0 {
		t.Fatalf("delivery must keep creation cohort separate from period completions: %#v", report.Overview.Metrics)
	}
	var created, completed int
	for _, point := range report.Trends.StoryCompletion {
		created += point.Created
		completed += point.Completed
	}
	if created != 1 || completed != 1 {
		t.Fatalf("period delivery trend created/completed = %d/%d, want 1/1", created, completed)
	}
	if report.Pulse.Filters.StartDate != nil || report.Pulse.Filters.EndDate != nil || report.Filters.StartDate == nil || report.Filters.EndDate == nil {
		t.Fatalf("filter metadata must distinguish current risk from selected period: pulse=%#v period=%#v", report.Pulse.Filters, report.Filters)
	}
}

func TestPulseObjectiveHealthExcludesClosedObjectivesWithStaleRisk(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	workspaceID := insertReportTestWorkspace(t, ctx, postgres.Pool, "objective-risk")
	actorID := insertReportTestUser(t, ctx, postgres.Pool, "objective-risk", true)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceID, actorID)
	teamID := insertReportTestTeam(t, ctx, postgres.Pool, workspaceID, "objective-risk")
	insertReportTestTeamMember(t, ctx, postgres.Pool, teamID, actorID)
	for sequence, objective := range []struct{ category, health string }{
		{"started", "At Risk"}, {"started", "Off Track"},
		{"completed", "At Risk"}, {"cancelled", "Off Track"},
	} {
		statusID := uuid.New()
		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO objective_statuses (status_id, name, category, workspace_id)
			VALUES ($1, $2, $2, $3)
		`, statusID, objective.category, workspaceID)
		if err != nil {
			t.Fatalf("insert objective status: %v", err)
		}
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO objectives (objective_id, sequence_id, name, team_id, workspace_id,
				status_id, lead_user_id, health, created_at, end_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, CAST($8 AS objective_health_status),
				NOW() - INTERVAL '60 days', CURRENT_DATE - 1)
		`, uuid.New(), sequence+1, uuid.NewString(), teamID, workspaceID, statusID, actorID, objective.health)
		if err != nil {
			t.Fatalf("insert objective: %v", err)
		}
	}
	repository := New(logger.NewWithText(io.Discard, slog.LevelError, "pulse-objective-integration"), postgres.Pool)
	health, err := repository.GetPulseObjectiveHealth(ctx, workspaceID, reportdomain.ReportFilters{
		ActorID: actorID, TeamIDs: []uuid.UUID{teamID}, AssigneeIDs: []uuid.UUID{actorID},
	})
	if err != nil {
		t.Fatalf("get current objective health: %v", err)
	}
	if health.ActiveObjectives != 2 || health.AtRiskObjectives != 1 || health.OffTrackObjectives != 1 || health.OverdueObjectives != 2 {
		t.Fatalf("only unfinished objectives should create operational risk: %#v", health)
	}
}

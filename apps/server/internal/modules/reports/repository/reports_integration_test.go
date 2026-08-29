//go:build integration

package reportsrepository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	reportdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReportsRepositoryEnforcesActorTenantAndTeamBoundaries(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	workspaceA := insertReportTestWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertReportTestWorkspace(t, ctx, postgres.Pool, "b")
	activeA := insertReportTestUser(t, ctx, postgres.Pool, "active-a", true)
	activeB := insertReportTestUser(t, ctx, postgres.Pool, "active-b", true)
	inactiveA := insertReportTestUser(t, ctx, postgres.Pool, "inactive-a", false)
	guestA := insertReportTestUser(t, ctx, postgres.Pool, "guest-a", true)
	adminA := insertReportTestUser(t, ctx, postgres.Pool, "admin-a", true)
	privateMemberA := insertReportTestUser(t, ctx, postgres.Pool, "private-member-a", true)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, activeA)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, inactiveA)
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceB, activeB)
	insertReportTestWorkspaceMemberWithRole(t, ctx, postgres.Pool, workspaceA, guestA, "guest")
	insertReportTestWorkspaceMemberWithRole(t, ctx, postgres.Pool, workspaceA, adminA, "admin")
	insertReportTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, privateMemberA)
	teamA := insertReportTestTeam(t, ctx, postgres.Pool, workspaceA, "a")
	teamB := insertReportTestTeam(t, ctx, postgres.Pool, workspaceB, "b")
	privateTeamA := insertReportTestTeamWithPrivacy(t, ctx, postgres.Pool, workspaceA, "private-a", true)
	insertReportTestTeamMember(t, ctx, postgres.Pool, teamA, activeA)
	insertReportTestTeamMember(t, ctx, postgres.Pool, teamB, activeB)
	insertReportTestTeamMember(t, ctx, postgres.Pool, privateTeamA, privateMemberA)
	statusA := insertReportTestStatus(t, ctx, postgres.Pool, workspaceA, teamA, "Started", "started")
	statusB := insertReportTestStatus(t, ctx, postgres.Pool, workspaceB, teamB, "Started", "started")
	privateStatusA := insertReportTestStatus(t, ctx, postgres.Pool, workspaceA, privateTeamA, "Private Started", "started")
	insertReportTestStory(t, ctx, postgres.Pool, workspaceA, teamA, statusA, activeA, "workspace-a")
	insertReportTestStory(t, ctx, postgres.Pool, workspaceB, teamB, statusB, activeB, "workspace-b")
	insertReportTestStory(t, ctx, postgres.Pool, workspaceA, privateTeamA, privateStatusA, privateMemberA, "private-workspace-a")

	repository := New(logger.NewWithText(io.Discard, slog.LevelError, "reports-integration"), postgres.Pool)
	startDate := time.Now().UTC().Add(-24 * time.Hour)
	endDate := time.Now().UTC().Add(24 * time.Hour)

	storyStats, err := repository.GetStoryStats(ctx, workspaceA, reportdomain.StoryStatsFilters{
		ActorID: activeA, StartDate: startDate, EndDate: endDate,
	})
	if err != nil {
		t.Fatalf("get authorized story stats: %v", err)
	}
	if storyStats.Assigned != 1 || storyStats.Created != 1 || storyStats.InProgress != 1 {
		t.Fatalf("authorized story stats = %#v", storyStats)
	}

	analysis, err := repository.GetWorkloadAnalysis(ctx, workspaceA, reportdomain.ReportFilters{ActorID: activeA})
	if err != nil {
		t.Fatalf("get authorized workload: %v", err)
	}
	if analysis.Summary.TotalOpenStories != 1 || len(analysis.Members) != 1 || analysis.Members[0].UserID != activeA {
		t.Fatalf("authorized workload = %#v", analysis)
	}

	storyAnalytics, err := repository.GetStoryAnalytics(ctx, workspaceA, reportdomain.ReportFilters{
		ActorID:   activeA,
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err != nil {
		t.Fatalf("get member story analytics with empty team filter: %v", err)
	}
	if len(storyAnalytics.CompletionByTeam) != 1 || storyAnalytics.CompletionByTeam[0].TeamID != teamA {
		t.Fatalf("member completion teams = %#v, want only visible public team", storyAnalytics.CompletionByTeam)
	}

	_, err = repository.GetWorkloadAnalysis(ctx, workspaceA, reportdomain.ReportFilters{
		ActorID: activeA,
		TeamIDs: []uuid.UUID{teamB},
	})
	if !errors.Is(err, reportdomain.ErrReportsAccessDenied) {
		t.Fatalf("cross-tenant team filter error = %v, want access denied", err)
	}

	_, err = repository.GetWorkloadAnalysis(ctx, workspaceA, reportdomain.ReportFilters{
		ActorID: activeA,
		TeamIDs: []uuid.UUID{privateTeamA},
	})
	if !errors.Is(err, reportdomain.ErrReportsAccessDenied) {
		t.Fatalf("unauthorized private-team filter error = %v, want access denied", err)
	}

	privateAnalysis, err := repository.GetWorkloadAnalysis(ctx, workspaceA, reportdomain.ReportFilters{ActorID: privateMemberA})
	if err != nil {
		t.Fatalf("get private-team member workload: %v", err)
	}
	if privateAnalysis.Summary.TotalOpenStories != 2 {
		t.Fatalf("private-team member open stories = %d, want public plus joined private stories", privateAnalysis.Summary.TotalOpenStories)
	}

	adminAnalysis, err := repository.GetWorkloadAnalysis(ctx, workspaceA, reportdomain.ReportFilters{ActorID: adminA})
	if err != nil {
		t.Fatalf("get admin workload: %v", err)
	}
	if adminAnalysis.Summary.TotalOpenStories != 2 {
		t.Fatalf("admin open stories = %d, want all workspace stories", adminAnalysis.Summary.TotalOpenStories)
	}

	for name, actorID := range map[string]uuid.UUID{
		"cross-tenant actor": activeB,
		"inactive actor":     inactiveA,
		"guest actor":        guestA,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := repository.GetWorkloadAnalysis(ctx, workspaceA, reportdomain.ReportFilters{ActorID: actorID})
			if !errors.Is(err, reportdomain.ErrReportsAccessDenied) {
				t.Fatalf("workload error = %v, want access denied", err)
			}
		})
	}

	if err := repository.CreateWorkspaceAnalyticsEvent(ctx, reportdomain.CoreWorkspaceAnalyticsEventInput{
		WorkspaceID: workspaceA,
		UserID:      activeA,
		TeamID:      &teamB,
		EventName:   "report_opened",
		Surface:     "analytics",
		Properties:  map[string]any{"report": "workload"},
		OccurredAt:  time.Now().UTC(),
	}); !errors.Is(err, reportdomain.ErrReportsAccessDenied) {
		t.Fatalf("cross-tenant analytics event error = %v, want access denied", err)
	}
	if err := repository.CreateWorkspaceAnalyticsEvent(ctx, reportdomain.CoreWorkspaceAnalyticsEventInput{
		WorkspaceID: workspaceA,
		UserID:      activeA,
		TeamID:      &privateTeamA,
		EventName:   "report_opened",
		Surface:     "analytics",
		Properties:  map[string]any{"report": "workload"},
		OccurredAt:  time.Now().UTC(),
	}); !errors.Is(err, reportdomain.ErrReportsAccessDenied) {
		t.Fatalf("private-team analytics event error = %v, want access denied", err)
	}

	if err := repository.CreateWorkspaceAnalyticsEvent(ctx, reportdomain.CoreWorkspaceAnalyticsEventInput{
		WorkspaceID: workspaceA,
		UserID:      activeA,
		TeamID:      &teamA,
		EventName:   "report_opened",
		Surface:     "analytics",
		Properties:  map[string]any{"report": "workload"},
		OccurredAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create authorized analytics event: %v", err)
	}
	var eventCount int
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM workspace_analytics_events
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceA, activeA).Scan(&eventCount); err != nil {
		t.Fatalf("count analytics events: %v", err)
	}
	if eventCount != 1 {
		t.Fatalf("analytics event count = %d, want 1", eventCount)
	}
}

func insertReportTestWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`, id, "Reports "+label, "reports-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertReportTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, $5, FALSE)
	`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", "Reports "+label, active); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertReportTestWorkspaceMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, userID uuid.UUID) {
	t.Helper()
	insertReportTestWorkspaceMemberWithRole(t, ctx, pool, workspaceID, userID, "member")
}

func insertReportTestWorkspaceMemberWithRole(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, userID uuid.UUID, role string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)`, workspaceID, userID, role); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertReportTestTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, label string) uuid.UUID {
	t.Helper()
	return insertReportTestTeamWithPrivacy(t, ctx, pool, workspaceID, label, false)
}

func insertReportTestTeamWithPrivacy(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, label string, private bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, code, color, workspace_id, is_private)
		VALUES ($1, $2, $3, '#000000', $4, $5)
	`, id, "Reports Team "+label, "R"+uuid.NewString()[:7], workspaceID, private); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func insertReportTestTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID uuid.UUID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID); err != nil {
		t.Fatalf("insert team member: %v", err)
	}
}

func insertReportTestStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, teamID uuid.UUID, name string, category string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO statuses (status_id, name, category, order_index, workspace_id, team_id)
		VALUES ($1, $2, $3, 1, $4, $5)
	`, id, name, category, workspaceID, teamID); err != nil {
		t.Fatalf("insert status: %v", err)
	}
	return id
}

func insertReportTestStory(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, teamID uuid.UUID, statusID uuid.UUID, userID uuid.UUID, label string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO stories (
			id, sequence_id, team_id, title, status_id, assignee_id, reporter_id,
			priority, workspace_id, estimate_unit, created_at, updated_at
		)
		VALUES ($1, 1, $2, $3, $4, $5, $5, 'High', $6, 3, NOW(), NOW())
	`, uuid.New(), teamID, "Reports story "+label, statusID, userID, workspaceID); err != nil {
		t.Fatalf("insert story: %v", err)
	}
}

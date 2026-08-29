//go:build integration

package teamsettingsrepository

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

type teamSettingsFixture struct {
	workspaceA    uuid.UUID
	workspaceB    uuid.UUID
	teamA         uuid.UUID
	teamB         uuid.UUID
	activeUserA   uuid.UUID
	inactiveUserA uuid.UUID
}

func newTeamSettingsFixture(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
) teamSettingsFixture {
	t.Helper()
	workspaceA := insertTeamSettingsWorkspace(t, ctx, postgres, "a")
	workspaceB := insertTeamSettingsWorkspace(t, ctx, postgres, "b")
	teamA := insertTeamSettingsTeam(t, ctx, postgres, workspaceA, "A")
	teamB := insertTeamSettingsTeam(t, ctx, postgres, workspaceB, "B")
	activeUserA := insertTeamSettingsUser(t, ctx, postgres, "active-a", true)
	inactiveUserA := insertTeamSettingsUser(t, ctx, postgres, "inactive-a", false)
	insertTeamSettingsWorkspaceMember(t, ctx, postgres, workspaceA, activeUserA)
	insertTeamSettingsWorkspaceMember(t, ctx, postgres, workspaceA, inactiveUserA)
	insertTeamSettingsTeamMember(t, ctx, postgres, teamA, activeUserA)
	insertTeamSettingsTeamMember(t, ctx, postgres, teamA, inactiveUserA)
	return teamSettingsFixture{
		workspaceA:    workspaceA,
		workspaceB:    workspaceB,
		teamA:         teamA,
		teamB:         teamB,
		activeUserA:   activeUserA,
		inactiveUserA: inactiveUserA,
	}
}

func insertTeamSettingsWorkspace(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	label string,
) uuid.UUID {
	t.Helper()
	workspaceID := uuid.New()
	suffix := uuid.NewString()
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, $2, $3)
	`, workspaceID, "Team settings "+label, "team-settings-"+label+"-"+suffix); err != nil {
		t.Fatalf("insert workspace %s: %v", label, err)
	}
	return workspaceID
}

func insertTeamSettingsTeam(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	workspaceID uuid.UUID,
	label string,
) uuid.UUID {
	t.Helper()
	teamID := uuid.New()
	code := fmt.Sprintf("TS%s", uuid.NewString()[:6])
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, teamID, "Team settings "+label, workspaceID, code); err != nil {
		t.Fatalf("insert team %s: %v", label, err)
	}
	return teamID
}

func insertTeamSettingsUser(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	label string,
	active bool,
) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	suffix := uuid.NewString()
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, label+"-"+suffix, label+"-"+suffix+"@example.com", "Team settings "+label, active); err != nil {
		t.Fatalf("insert user %s: %v", label, err)
	}
	return userID
}

func insertTeamSettingsWorkspaceMember(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	workspaceID, userID uuid.UUID,
) {
	t.Helper()
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertTeamSettingsTeamMember(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	teamID, userID uuid.UUID,
) {
	t.Helper()
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
	`, teamID, userID); err != nil {
		t.Fatalf("insert team member: %v", err)
	}
}

func insertTeamSettingsSprint(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	teamID, workspaceID uuid.UUID,
	name string,
	startDate, endDate time.Time,
	managed bool,
) uuid.UUID {
	t.Helper()
	sprintID := uuid.New()
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO sprints (
			sprint_id, team_id, workspace_id, name, start_date, end_date,
			schedule_managed_by_automation
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, sprintID, teamID, workspaceID, name, startDate, endDate, managed); err != nil {
		t.Fatalf("insert sprint %s: %v", name, err)
	}
	return sprintID
}

func readSprintDates(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	sprintID uuid.UUID,
) (time.Time, time.Time) {
	t.Helper()
	var startDate, endDate time.Time
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT start_date, end_date
		FROM sprints
		WHERE sprint_id = $1
	`, sprintID).Scan(&startDate, &endDate); err != nil {
		t.Fatalf("read sprint dates: %v", err)
	}
	return startDate, endDate
}

func assertSettingsRowCount(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	table string,
	teamID, workspaceID uuid.UUID,
	want int,
) {
	t.Helper()
	allowedTables := map[string]bool{
		"team_sprint_settings":           true,
		"team_story_automation_settings": true,
		"team_estimation_settings":       true,
	}
	if !allowedTables[table] {
		t.Fatalf("unsupported settings table %q", table)
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE team_id = $1 AND workspace_id = $2", table)
	var count int
	if err := postgres.Pool.QueryRow(ctx, query, teamID, workspaceID).Scan(&count); err != nil {
		t.Fatalf("count %s rows: %v", table, err)
	}
	if count != want {
		t.Fatalf("%s row count = %d, want %d", table, count, want)
	}
}

func readAuditEventCount(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	workspaceID, teamID uuid.UUID,
) int {
	t.Helper()
	var count int
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM audit_events
		WHERE workspace_id = $1 AND team_id = $2
	`, workspaceID, teamID).Scan(&count); err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	return count
}

func assertAuditEventCount(
	t *testing.T,
	ctx context.Context,
	postgres *testkit.Postgres,
	workspaceID, teamID uuid.UUID,
	want int,
) {
	t.Helper()
	if got := readAuditEventCount(t, ctx, postgres, workspaceID, teamID); got != want {
		t.Fatalf("audit event count = %d, want %d", got, want)
	}
}

func assertPostgres18(t *testing.T, ctx context.Context, postgres *testkit.Postgres) {
	t.Helper()
	var rawVersionNumber string
	if err := postgres.Pool.QueryRow(ctx, "SHOW server_version_num").Scan(&rawVersionNumber); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	versionNumber, err := strconv.Atoi(rawVersionNumber)
	if err != nil {
		t.Fatalf("parse PostgreSQL version %q: %v", rawVersionNumber, err)
	}
	if versionNumber < 180000 || versionNumber >= 190000 {
		t.Fatalf("PostgreSQL version = %d, want 18.x", versionNumber)
	}
}

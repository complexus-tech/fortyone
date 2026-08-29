//go:build integration

package objectivesrepository

import (
	"context"
	"fmt"
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type objectiveIntegrationFixture struct {
	postgres *testkit.Postgres
	repo     *Repository

	workspaceA uuid.UUID
	workspaceB uuid.UUID
	teamA      uuid.UUID
	teamB      uuid.UUID
	statusA    uuid.UUID
	statusB    uuid.UUID
	actorA     uuid.UUID
	actorB     uuid.UUID
	assigneeA  uuid.UUID
	outsiderA  uuid.UUID
	inactiveA  uuid.UUID
	guestA     uuid.UUID
}

func newObjectiveIntegrationFixture(t *testing.T, ctx context.Context) objectiveIntegrationFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	fixture := objectiveIntegrationFixture{
		postgres: postgres, repo: New(postgres.Pool),
		workspaceA: uuid.New(), workspaceB: uuid.New(),
		teamA: uuid.New(), teamB: uuid.New(), statusA: uuid.New(), statusB: uuid.New(),
		actorA: uuid.New(), actorB: uuid.New(), assigneeA: uuid.New(),
		outsiderA: uuid.New(), inactiveA: uuid.New(), guestA: uuid.New(),
	}

	insertObjectiveUser(t, ctx, postgres.Pool, fixture.actorA, "actor-a", true)
	insertObjectiveUser(t, ctx, postgres.Pool, fixture.actorB, "actor-b", true)
	insertObjectiveUser(t, ctx, postgres.Pool, fixture.assigneeA, "assignee-a", true)
	insertObjectiveUser(t, ctx, postgres.Pool, fixture.outsiderA, "outsider-a", true)
	insertObjectiveUser(t, ctx, postgres.Pool, fixture.inactiveA, "inactive-a", false)
	insertObjectiveUser(t, ctx, postgres.Pool, fixture.guestA, "guest-a", true)
	insertObjectiveWorkspace(t, ctx, postgres.Pool, fixture.workspaceA, fixture.actorA, "a")
	insertObjectiveWorkspace(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB, "b")
	insertObjectiveTeam(t, ctx, postgres.Pool, fixture.teamA, fixture.workspaceA, "OBJA")
	insertObjectiveTeam(t, ctx, postgres.Pool, fixture.teamB, fixture.workspaceB, "OBJB")
	insertObjectiveStatus(t, ctx, postgres.Pool, fixture.statusA, fixture.workspaceA, "Started A")
	insertObjectiveStatus(t, ctx, postgres.Pool, fixture.statusB, fixture.workspaceB, "Started B")

	for _, userID := range []uuid.UUID{fixture.actorA, fixture.assigneeA, fixture.inactiveA} {
		insertObjectiveWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, userID)
		insertObjectiveTeamMember(t, ctx, postgres.Pool, fixture.teamA, userID)
	}
	insertObjectiveWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.outsiderA)
	mustObjectiveExec(t, ctx, postgres.Pool, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'guest')
	`, fixture.workspaceA, fixture.guestA)
	insertObjectiveTeamMember(t, ctx, postgres.Pool, fixture.teamA, fixture.guestA)
	insertObjectiveWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB)
	insertObjectiveTeamMember(t, ctx, postgres.Pool, fixture.teamB, fixture.actorB)
	return fixture
}

func (fixture objectiveIntegrationFixture) createCommand(name string) objectivesdomain.CreateCommand {
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	return objectivesdomain.CreateCommand{
		WorkspaceID: fixture.workspaceA,
		Objective: objectivesdomain.NewObjective{
			Name: name, Team: fixture.teamA, Status: fixture.statusA,
			LeadUser: &fixture.assigneeA, StartDate: &start, EndDate: &end,
			Color: "#686DE0", CreatedBy: fixture.actorA,
		},
		KeyResults: []objectivesdomain.NewKeyResult{
			{
				Name: "Adoption", MeasurementType: "percentage", TargetValue: 80,
				Lead: &fixture.assigneeA, Contributors: []uuid.UUID{fixture.assigneeA, fixture.assigneeA},
				StartDate: &start, EndDate: &end,
			},
			{
				Name: "Reliability", MeasurementType: "number", TargetValue: 99.9,
				StartDate: &start, EndDate: &end,
			},
		},
	}
}

func (fixture objectiveIntegrationFixture) foreignCreateCommand(name string) objectivesdomain.CreateCommand {
	command := fixture.createCommand(name)
	command.WorkspaceID = fixture.workspaceB
	command.Objective.Team = fixture.teamB
	command.Objective.Status = fixture.statusB
	command.Objective.CreatedBy = fixture.actorB
	command.Objective.LeadUser = nil
	command.KeyResults = nil
	return command
}

func insertObjectiveUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, label string, active bool) {
	t.Helper()
	suffix := uuid.NewString()
	mustObjectiveExec(t, ctx, pool, `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", "Objective "+label, active)
}

func insertObjectiveWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, creatorID uuid.UUID, label string) {
	t.Helper()
	mustObjectiveExec(t, ctx, pool, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, id, "Objectives "+label, "objectives-"+label+"-"+uuid.NewString(), creatorID)
}

func insertObjectiveTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID, code string) {
	t.Helper()
	mustObjectiveExec(t, ctx, pool, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, "Objectives "+code, workspaceID, code)
}

func insertObjectiveStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID, name string) {
	t.Helper()
	mustObjectiveExec(t, ctx, pool, `
		INSERT INTO objective_statuses (status_id, name, category, workspace_id, color)
		VALUES ($1, $2, 'started', $3, '#000000')
	`, id, name, workspaceID)
}

func insertObjectiveWorkspaceMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID) {
	t.Helper()
	mustObjectiveExec(t, ctx, pool, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, workspaceID, userID)
}

func insertObjectiveTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	mustObjectiveExec(t, ctx, pool, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID)
}

func mustObjectiveExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, arguments ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, query, arguments...); err != nil {
		t.Fatalf("execute objective fixture SQL: %v", err)
	}
}

func objectiveRowCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, arguments ...any) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, query, arguments...).Scan(&count); err != nil {
		t.Fatalf("count objective rows: %v", err)
	}
	return count
}

func uniqueObjectiveName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.NewString())
}

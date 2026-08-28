//go:build integration

package keyresultsrepository

import (
	"context"
	"fmt"
	"testing"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type keyResultIntegrationFixture struct {
	postgres *testkit.Postgres
	repo     *Repository

	workspaceA uuid.UUID
	workspaceB uuid.UUID
	teamA      uuid.UUID
	teamB      uuid.UUID
	objectiveA uuid.UUID
	objectiveB uuid.UUID
	actorA     uuid.UUID
	actorB     uuid.UUID
	assigneeA  uuid.UUID
	outsiderA  uuid.UUID
	inactiveA  uuid.UUID
}

func newKeyResultIntegrationFixture(t *testing.T, ctx context.Context) keyResultIntegrationFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	fixture := keyResultIntegrationFixture{
		postgres:   postgres,
		repo:       New(postgres.Pool),
		workspaceA: uuid.New(), workspaceB: uuid.New(),
		teamA: uuid.New(), teamB: uuid.New(), objectiveA: uuid.New(), objectiveB: uuid.New(),
		actorA: uuid.New(), actorB: uuid.New(), assigneeA: uuid.New(), outsiderA: uuid.New(), inactiveA: uuid.New(),
	}

	insertKeyResultUser(t, ctx, postgres.Pool, fixture.actorA, "actor-a", true)
	insertKeyResultUser(t, ctx, postgres.Pool, fixture.actorB, "actor-b", true)
	insertKeyResultUser(t, ctx, postgres.Pool, fixture.assigneeA, "assignee-a", true)
	insertKeyResultUser(t, ctx, postgres.Pool, fixture.outsiderA, "outsider-a", true)
	insertKeyResultUser(t, ctx, postgres.Pool, fixture.inactiveA, "inactive-a", false)
	insertKeyResultWorkspace(t, ctx, postgres.Pool, fixture.workspaceA, fixture.actorA, "a")
	insertKeyResultWorkspace(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB, "b")
	insertKeyResultTeam(t, ctx, postgres.Pool, fixture.teamA, fixture.workspaceA, "KRA")
	insertKeyResultTeam(t, ctx, postgres.Pool, fixture.teamB, fixture.workspaceB, "KRB")

	for _, userID := range []uuid.UUID{fixture.actorA, fixture.assigneeA, fixture.inactiveA} {
		insertKeyResultWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, userID, "member")
		insertKeyResultTeamMember(t, ctx, postgres.Pool, fixture.teamA, userID)
	}
	insertKeyResultWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.outsiderA, "member")
	insertKeyResultWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB, "admin")
	insertKeyResultTeamMember(t, ctx, postgres.Pool, fixture.teamB, fixture.actorB)
	insertKeyResultObjective(t, ctx, postgres.Pool, fixture.objectiveA, fixture.teamA, fixture.workspaceA, fixture.actorA, 1, "Objective A")
	insertKeyResultObjective(t, ctx, postgres.Pool, fixture.objectiveB, fixture.teamB, fixture.workspaceB, fixture.actorB, 1, "Objective B")
	return fixture
}

func (fixture keyResultIntegrationFixture) accessA() keyresultsdomain.AccessScope {
	return keyresultsdomain.AccessScope{WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA, AllTeams: true}
}

func (fixture keyResultIntegrationFixture) createCommand(name string) keyresultsdomain.CreateCommand {
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	return keyresultsdomain.CreateCommand{
		Access: fixture.accessA(),
		KeyResults: []keyresultsdomain.NewKeyResult{{
			ObjectiveID: fixture.objectiveA, Name: name, MeasurementType: "percentage",
			CurrentValue: 25, TargetValue: 100, Lead: &fixture.assigneeA,
			Contributors: []uuid.UUID{fixture.assigneeA, fixture.assigneeA},
			StartDate:    &start, EndDate: &end, CreatedBy: fixture.actorA,
		}},
	}
}

func (fixture keyResultIntegrationFixture) foreignCreateCommand(name string) keyresultsdomain.CreateCommand {
	command := fixture.createCommand(name)
	command.Access = keyresultsdomain.AccessScope{WorkspaceID: fixture.workspaceB, ActorID: fixture.actorB, AllTeams: true}
	command.KeyResults[0].ObjectiveID = fixture.objectiveB
	command.KeyResults[0].CreatedBy = fixture.actorB
	command.KeyResults[0].Lead = nil
	command.KeyResults[0].Contributors = nil
	return command
}

func insertKeyResultUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, label string, active bool) {
	t.Helper()
	suffix := uuid.NewString()
	mustKeyResultExec(t, ctx, pool, `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", "Key Result "+label, active)
}

func insertKeyResultWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, creatorID uuid.UUID, label string) {
	t.Helper()
	mustKeyResultExec(t, ctx, pool, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, id, "Key Results "+label, "key-results-"+label+"-"+uuid.NewString(), creatorID)
}

func insertKeyResultTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID, code string) {
	t.Helper()
	mustKeyResultExec(t, ctx, pool, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, "Key Results "+code, workspaceID, code)
}

func insertKeyResultWorkspaceMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, role string) {
	t.Helper()
	mustKeyResultExec(t, ctx, pool, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
	`, workspaceID, userID, role)
}

func insertKeyResultTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	mustKeyResultExec(t, ctx, pool, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID)
}

func insertKeyResultObjective(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, teamID, workspaceID, actorID uuid.UUID,
	sequenceID int,
	name string,
) {
	t.Helper()
	mustKeyResultExec(t, ctx, pool, `
		INSERT INTO objectives (
			objective_id, name, team_id, workspace_id, created_by, sequence_id, color
		) VALUES ($1, $2, $3, $4, $5, $6, '#686DE0')
	`, id, name+"-"+uuid.NewString(), teamID, workspaceID, actorID, sequenceID)
}

func mustKeyResultExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, arguments ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, query, arguments...); err != nil {
		t.Fatalf("execute key-result fixture SQL: %v", err)
	}
}

func keyResultRowCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, arguments ...any) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, query, arguments...).Scan(&count); err != nil {
		t.Fatalf("count key-result rows: %v", err)
	}
	return count
}

func uniqueKeyResultName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.NewString())
}

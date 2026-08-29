//go:build integration

package okractivitiesrepository

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	okractivitiesdomain "github.com/complexus-tech/projects-api/internal/modules/okractivities/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryTenantTransactionsAndStablePagination(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	fixture := newActivityIntegrationFixture(t, ctx)
	assertActivityPostgres18(t, ctx, fixture.postgres.Pool)

	keyResultID := fixture.keyResultA
	activities := []okractivitiesdomain.NewActivity{
		fixture.activity(okractivitiesdomain.ActivityTypeCreate, okractivitiesdomain.UpdateTypeObjective, nil, "name"),
		fixture.activity(okractivitiesdomain.ActivityTypeCreate, okractivitiesdomain.UpdateTypeKeyResult, &keyResultID, "name"),
		fixture.activity(okractivitiesdomain.ActivityTypeUpdate, okractivitiesdomain.UpdateTypeKeyResult, &keyResultID, "current_value"),
	}
	if err := fixture.repo.CreateBatch(ctx, activities); err != nil {
		t.Fatalf("create activity batch: %v", err)
	}

	testActivityStablePagination(t, ctx, fixture)
	testActivityTenantAndMembershipScope(t, ctx, fixture)
	testActivityBatchRollback(t, ctx, fixture)
	testActivityResourceRelationshipScope(t, ctx, fixture)
}

type activityIntegrationFixture struct {
	postgres *testkit.Postgres
	repo     *Repository

	workspaceA uuid.UUID
	workspaceB uuid.UUID
	teamA      uuid.UUID
	teamB      uuid.UUID
	objectiveA uuid.UUID
	objectiveB uuid.UUID
	keyResultA uuid.UUID
	keyResultB uuid.UUID
	actorA     uuid.UUID
	actorB     uuid.UUID
	outsiderA  uuid.UUID
	inactiveA  uuid.UUID
}

func newActivityIntegrationFixture(t *testing.T, ctx context.Context) activityIntegrationFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	fixture := activityIntegrationFixture{
		postgres:   postgres,
		repo:       New(postgres.Pool),
		workspaceA: uuid.New(), workspaceB: uuid.New(),
		teamA: uuid.New(), teamB: uuid.New(), objectiveA: uuid.New(), objectiveB: uuid.New(),
		keyResultA: uuid.New(), keyResultB: uuid.New(), actorA: uuid.New(), actorB: uuid.New(),
		outsiderA: uuid.New(), inactiveA: uuid.New(),
	}

	insertActivityUser(t, ctx, postgres.Pool, fixture.actorA, "activity-actor-a", true)
	insertActivityUser(t, ctx, postgres.Pool, fixture.actorB, "activity-actor-b", true)
	insertActivityUser(t, ctx, postgres.Pool, fixture.outsiderA, "activity-outsider-a", true)
	insertActivityUser(t, ctx, postgres.Pool, fixture.inactiveA, "activity-inactive-a", false)
	insertActivityWorkspace(t, ctx, postgres.Pool, fixture.workspaceA, fixture.actorA, "activity-a")
	insertActivityWorkspace(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB, "activity-b")
	insertActivityTeam(t, ctx, postgres.Pool, fixture.teamA, fixture.workspaceA, "OAA")
	insertActivityTeam(t, ctx, postgres.Pool, fixture.teamB, fixture.workspaceB, "OBB")

	for _, actorID := range []uuid.UUID{fixture.actorA, fixture.inactiveA} {
		insertActivityWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, actorID, "member")
		insertActivityTeamMember(t, ctx, postgres.Pool, fixture.teamA, actorID)
	}
	insertActivityWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.outsiderA, "member")
	insertActivityWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB, "admin")
	insertActivityTeamMember(t, ctx, postgres.Pool, fixture.teamB, fixture.actorB)

	insertActivityObjective(t, ctx, postgres.Pool, fixture.objectiveA, fixture.teamA, fixture.workspaceA, fixture.actorA, "Objective A")
	insertActivityObjective(t, ctx, postgres.Pool, fixture.objectiveB, fixture.teamB, fixture.workspaceB, fixture.actorB, "Objective B")
	insertActivityKeyResult(t, ctx, postgres.Pool, fixture.keyResultA, fixture.objectiveA, fixture.teamA, fixture.actorA, "Key Result A")
	insertActivityKeyResult(t, ctx, postgres.Pool, fixture.keyResultB, fixture.objectiveB, fixture.teamB, fixture.actorB, "Key Result B")
	return fixture
}

func (fixture activityIntegrationFixture) activity(
	activityType okractivitiesdomain.ActivityType,
	updateType okractivitiesdomain.UpdateType,
	keyResultID *uuid.UUID,
	field string,
) okractivitiesdomain.NewActivity {
	return okractivitiesdomain.NewActivity{
		ObjectiveID: fixture.objectiveA, KeyResultID: keyResultID,
		UserID: fixture.actorA, Type: activityType, UpdateType: updateType,
		Field: field, CurrentValue: "after", Comment: "integration test",
		WorkspaceID: fixture.workspaceA,
	}
}

func testActivityStablePagination(t *testing.T, ctx context.Context, fixture activityIntegrationFixture) {
	t.Helper()
	query := okractivitiesdomain.ListQuery{
		ObjectiveID: fixture.objectiveA, WorkspaceID: fixture.workspaceA,
		ActorID: fixture.actorA, Page: 1, PageSize: 2,
	}
	first, hasMore, err := fixture.repo.List(ctx, query)
	if err != nil || len(first) != 2 || !hasMore {
		t.Fatalf("first objective activity page = %#v, hasMore=%v, err=%v", first, hasMore, err)
	}
	repeated, repeatedHasMore, err := fixture.repo.List(ctx, query)
	if err != nil || !repeatedHasMore || len(repeated) != 2 || repeated[0].ID != first[0].ID || repeated[1].ID != first[1].ID {
		t.Fatalf("repeated objective page = %#v, hasMore=%v, err=%v", repeated, repeatedHasMore, err)
	}
	query.Page = 2
	second, hasMore, err := fixture.repo.List(ctx, query)
	if err != nil || len(second) != 1 || hasMore {
		t.Fatalf("second objective activity page = %#v, hasMore=%v, err=%v", second, hasMore, err)
	}

	keyResultID := fixture.keyResultA
	keyResultActivities, hasMore, err := fixture.repo.List(ctx, okractivitiesdomain.ListQuery{
		KeyResultID: &keyResultID, WorkspaceID: fixture.workspaceA,
		ActorID: fixture.actorA, Page: 1, PageSize: 1,
	})
	if err != nil || len(keyResultActivities) != 1 || !hasMore || keyResultActivities[0].KeyResultID == nil || *keyResultActivities[0].KeyResultID != keyResultID {
		t.Fatalf("key-result activity page = %#v, hasMore=%v, err=%v", keyResultActivities, hasMore, err)
	}
}

func testActivityTenantAndMembershipScope(t *testing.T, ctx context.Context, fixture activityIntegrationFixture) {
	t.Helper()
	denied := []okractivitiesdomain.ListQuery{
		{ObjectiveID: fixture.objectiveA, WorkspaceID: fixture.workspaceB, ActorID: fixture.actorB, Page: 1, PageSize: 10},
		{ObjectiveID: fixture.objectiveA, WorkspaceID: fixture.workspaceA, ActorID: fixture.actorB, Page: 1, PageSize: 10},
		{ObjectiveID: fixture.objectiveA, WorkspaceID: fixture.workspaceA, ActorID: fixture.outsiderA, Page: 1, PageSize: 10},
		{ObjectiveID: fixture.objectiveA, WorkspaceID: fixture.workspaceA, ActorID: fixture.inactiveA, Page: 1, PageSize: 10},
	}
	for _, query := range denied {
		activities, hasMore, err := fixture.repo.List(ctx, query)
		if err != nil || len(activities) != 0 || hasMore {
			t.Fatalf("denied activity query %#v = %#v, hasMore=%v, err=%v", query, activities, hasMore, err)
		}
	}

	unauthorizedCreates := []okractivitiesdomain.NewActivity{
		{
			ObjectiveID: fixture.objectiveA, UserID: fixture.outsiderA,
			Type: okractivitiesdomain.ActivityTypeUpdate, UpdateType: okractivitiesdomain.UpdateTypeObjective,
			WorkspaceID: fixture.workspaceA,
		},
		{
			ObjectiveID: fixture.objectiveA, UserID: fixture.inactiveA,
			Type: okractivitiesdomain.ActivityTypeUpdate, UpdateType: okractivitiesdomain.UpdateTypeObjective,
			WorkspaceID: fixture.workspaceA,
		},
		{
			ObjectiveID: fixture.objectiveA, UserID: fixture.actorB,
			Type: okractivitiesdomain.ActivityTypeUpdate, UpdateType: okractivitiesdomain.UpdateTypeObjective,
			WorkspaceID: fixture.workspaceB,
		},
	}
	for _, activity := range unauthorizedCreates {
		if err := fixture.repo.Create(ctx, activity); !errors.Is(err, okractivitiesdomain.ErrScopeMismatch) {
			t.Fatalf("unauthorized create %#v error = %v, want ErrScopeMismatch", activity, err)
		}
	}
}

func testActivityBatchRollback(t *testing.T, ctx context.Context, fixture activityIntegrationFixture) {
	t.Helper()
	before := activityRowCount(t, ctx, fixture.postgres.Pool, fixture.objectiveA)
	valid := fixture.activity(okractivitiesdomain.ActivityTypeUpdate, okractivitiesdomain.UpdateTypeObjective, nil, "description")
	invalid := valid
	invalid.ObjectiveID = fixture.objectiveB
	if err := fixture.repo.CreateBatch(ctx, []okractivitiesdomain.NewActivity{valid, invalid}); !errors.Is(err, okractivitiesdomain.ErrScopeMismatch) {
		t.Fatalf("mixed-scope batch error = %v, want ErrScopeMismatch", err)
	}
	if after := activityRowCount(t, ctx, fixture.postgres.Pool, fixture.objectiveA); after != before {
		t.Fatalf("failed batch persisted activity: before=%d after=%d", before, after)
	}
}

func testActivityResourceRelationshipScope(t *testing.T, ctx context.Context, fixture activityIntegrationFixture) {
	t.Helper()
	foreignKeyResultID := fixture.keyResultB
	activity := fixture.activity(okractivitiesdomain.ActivityTypeUpdate, okractivitiesdomain.UpdateTypeKeyResult, &foreignKeyResultID, "name")
	if err := fixture.repo.Create(ctx, activity); !errors.Is(err, okractivitiesdomain.ErrScopeMismatch) {
		t.Fatalf("mismatched key-result/objective error = %v, want ErrScopeMismatch", err)
	}
}

func insertActivityUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, label string, active bool) {
	t.Helper()
	suffix := uuid.NewString()
	mustActivityExec(t, ctx, pool, `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", label, active)
}

func insertActivityWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, creatorID uuid.UUID, label string) {
	t.Helper()
	mustActivityExec(t, ctx, pool, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, id, label, label+"-"+uuid.NewString(), creatorID)
}

func insertActivityTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID, code string) {
	t.Helper()
	mustActivityExec(t, ctx, pool, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, code, workspaceID, code)
}

func insertActivityWorkspaceMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, role string) {
	t.Helper()
	mustActivityExec(t, ctx, pool, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)`, workspaceID, userID, role)
}

func insertActivityTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	mustActivityExec(t, ctx, pool, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID)
}

func insertActivityObjective(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, teamID, workspaceID, actorID uuid.UUID,
	name string,
) {
	t.Helper()
	mustActivityExec(t, ctx, pool, `
		INSERT INTO objectives (objective_id, name, team_id, workspace_id, created_by, sequence_id, color)
		VALUES ($1, $2, $3, $4, $5, 1, '#686DE0')
	`, id, name+"-"+uuid.NewString(), teamID, workspaceID, actorID)
}

func insertActivityKeyResult(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, objectiveID, teamID, actorID uuid.UUID,
	name string,
) {
	t.Helper()
	mustActivityExec(t, ctx, pool, `
		INSERT INTO key_results (
			id, objective_id, team_id, sequence_id, name, measurement_type,
			start_value, current_value, target_value, start_date, end_date, created_by
		) VALUES ($1, $2, $3, 1, $4, 'percentage', 0, 0, 100, DATE '2026-08-01', DATE '2026-09-30', $5)
	`, id, objectiveID, teamID, name+"-"+uuid.NewString(), actorID)
}

func mustActivityExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, arguments ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, query, arguments...); err != nil {
		t.Fatalf("execute OKR activity fixture SQL: %v", err)
	}
}

func activityRowCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, objectiveID uuid.UUID) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM okr_activities WHERE objective_id = $1`, objectiveID).Scan(&count); err != nil {
		t.Fatalf("count OKR activities: %v", err)
	}
	return count
}

func assertActivityPostgres18(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var raw string
	if err := pool.QueryRow(ctx, "SHOW server_version_num").Scan(&raw); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	version, err := strconv.Atoi(raw)
	if err != nil || version < 180000 || version >= 190000 {
		t.Fatalf("PostgreSQL version = %q, want 18.x", raw)
	}
}

//go:build integration

package sprintsrepository

import (
	"context"
	"fmt"
	"testing"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sprintIntegrationFixture struct {
	postgres *testkit.Postgres
	repo     *Repository

	workspaceA      uuid.UUID
	workspaceB      uuid.UUID
	teamA           uuid.UUID
	teamOtherA      uuid.UUID
	teamB           uuid.UUID
	objectiveA      uuid.UUID
	objectiveA2     uuid.UUID
	objectiveOtherA uuid.UUID
	objectiveB      uuid.UUID
	actorA          uuid.UUID
	actorB          uuid.UUID
	assigneeA       uuid.UUID
	outsiderA       uuid.UUID
	inactiveA       uuid.UUID
	guestA          uuid.UUID
	revocableA      uuid.UUID
	statuses        map[string]uuid.UUID
}

func newSprintIntegrationFixture(t *testing.T, ctx context.Context) sprintIntegrationFixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	fixture := sprintIntegrationFixture{
		postgres:        postgres,
		repo:            New(postgres.Pool),
		workspaceA:      uuid.New(),
		workspaceB:      uuid.New(),
		teamA:           uuid.New(),
		teamOtherA:      uuid.New(),
		teamB:           uuid.New(),
		objectiveA:      uuid.New(),
		objectiveA2:     uuid.New(),
		objectiveOtherA: uuid.New(),
		objectiveB:      uuid.New(),
		actorA:          uuid.New(),
		actorB:          uuid.New(),
		assigneeA:       uuid.New(),
		outsiderA:       uuid.New(),
		inactiveA:       uuid.New(),
		guestA:          uuid.New(),
		revocableA:      uuid.New(),
		statuses:        make(map[string]uuid.UUID),
	}

	insertSprintUser(t, ctx, postgres.Pool, fixture.actorA, "actor-a", true)
	insertSprintUser(t, ctx, postgres.Pool, fixture.actorB, "actor-b", true)
	insertSprintUser(t, ctx, postgres.Pool, fixture.assigneeA, "assignee-a", true)
	insertSprintUser(t, ctx, postgres.Pool, fixture.outsiderA, "outsider-a", true)
	insertSprintUser(t, ctx, postgres.Pool, fixture.inactiveA, "inactive-a", false)
	insertSprintUser(t, ctx, postgres.Pool, fixture.guestA, "guest-a", true)
	insertSprintUser(t, ctx, postgres.Pool, fixture.revocableA, "revocable-a", true)

	insertSprintWorkspace(t, ctx, postgres.Pool, fixture.workspaceA, fixture.actorA, "a")
	insertSprintWorkspace(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB, "b")
	insertSprintTeam(t, ctx, postgres.Pool, fixture.teamA, fixture.workspaceA, "SPA")
	insertSprintTeam(t, ctx, postgres.Pool, fixture.teamOtherA, fixture.workspaceA, "SPO")
	insertSprintTeam(t, ctx, postgres.Pool, fixture.teamB, fixture.workspaceB, "SPB")

	insertSprintWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.actorA, "member")
	insertSprintWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.assigneeA, "member")
	insertSprintWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.outsiderA, "member")
	insertSprintWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.inactiveA, "member")
	insertSprintWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.guestA, "guest")
	insertSprintWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceA, fixture.revocableA, "member")
	insertSprintWorkspaceMember(t, ctx, postgres.Pool, fixture.workspaceB, fixture.actorB, "admin")
	for _, userID := range []uuid.UUID{
		fixture.actorA, fixture.assigneeA, fixture.inactiveA, fixture.guestA, fixture.revocableA,
	} {
		insertSprintTeamMember(t, ctx, postgres.Pool, fixture.teamA, userID)
	}
	insertSprintTeamMember(t, ctx, postgres.Pool, fixture.teamB, fixture.actorB)

	insertSprintObjective(t, ctx, postgres.Pool, fixture.objectiveA, fixture.workspaceA, fixture.teamA, fixture.actorA, "Primary")
	insertSprintObjective(t, ctx, postgres.Pool, fixture.objectiveA2, fixture.workspaceA, fixture.teamA, fixture.actorA, "Secondary")
	insertSprintObjective(t, ctx, postgres.Pool, fixture.objectiveOtherA, fixture.workspaceA, fixture.teamOtherA, fixture.actorA, "Other team")
	insertSprintObjective(t, ctx, postgres.Pool, fixture.objectiveB, fixture.workspaceB, fixture.teamB, fixture.actorB, "Foreign")

	for index, category := range []string{"completed", "started", "unstarted", "paused", "cancelled"} {
		statusID := uuid.New()
		fixture.statuses[category] = statusID
		mustSprintExec(t, ctx, postgres.Pool, `
			INSERT INTO statuses (status_id, name, category, order_index, workspace_id, team_id, color)
			VALUES ($1, $2, $3, $4, $5, $6, '#000000')
		`, statusID, "Sprint "+category, category, index, fixture.workspaceA, fixture.teamA)
	}
	mustSprintExec(t, ctx, postgres.Pool, `
		INSERT INTO workspace_settings (workspace_id, working_days)
		VALUES ($1, ARRAY[1, 2, 3, 4, 5]::smallint[])
	`, fixture.workspaceA)
	mustSprintExec(t, ctx, postgres.Pool, `
		INSERT INTO team_sprint_settings (team_id, workspace_id, auto_create_sprints)
		VALUES ($1, $2, TRUE)
	`, fixture.teamA, fixture.workspaceA)
	return fixture
}

func (fixture sprintIntegrationFixture) createCommand(name string) sprintdomain.CreateCommand {
	goal := "Ship the smallest coherent planning increment"
	return sprintdomain.CreateCommand{
		ActorID: fixture.actorA,
		Sprint: sprintdomain.NewSprint{
			Name: name, Goal: &goal, ObjectiveID: &fixture.objectiveA,
			TeamID: fixture.teamA, WorkspaceID: fixture.workspaceA,
			StartDate: time.Date(2026, time.August, 24, 13, 0, 0, 0, time.FixedZone("fixture", 2*60*60)),
			EndDate:   time.Date(2026, time.September, 4, 17, 0, 0, 0, time.FixedZone("fixture", 2*60*60)),
		},
	}
}

func (fixture sprintIntegrationFixture) foreignCreateCommand(name string) sprintdomain.CreateCommand {
	command := fixture.createCommand(name)
	command.ActorID = fixture.actorB
	command.Sprint.WorkspaceID = fixture.workspaceB
	command.Sprint.TeamID = fixture.teamB
	command.Sprint.ObjectiveID = &fixture.objectiveB
	return command
}

func (fixture sprintIntegrationFixture) insertStory(
	t *testing.T,
	ctx context.Context,
	sprintID, assigneeID uuid.UUID,
	category string,
	sequence int,
	deleted, archived bool,
) uuid.UUID {
	t.Helper()
	storyID := uuid.New()
	var deletedAt, archivedAt *time.Time
	now := time.Now().UTC()
	if deleted {
		deletedAt = &now
	}
	if archived {
		archivedAt = &now
	}
	mustSprintExec(t, ctx, fixture.postgres.Pool, `
		INSERT INTO stories (
			id, sequence_id, team_id, title, status_id, assignee_id, reporter_id,
			sprint_id, workspace_id, deleted_at, archived_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, storyID, sequence, fixture.teamA, fmt.Sprintf("Sprint story %d", sequence),
		fixture.statuses[category], assigneeID, fixture.actorA, sprintID, fixture.workspaceA,
		deletedAt, archivedAt)
	return storyID
}

func insertSprintUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, label string, active bool) {
	t.Helper()
	suffix := uuid.NewString()
	mustSprintExec(t, ctx, pool, `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", "Sprint "+label, active)
}

func insertSprintWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, creatorID uuid.UUID, label string) {
	t.Helper()
	mustSprintExec(t, ctx, pool, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, id, "Sprints "+label, "sprints-"+label+"-"+uuid.NewString(), creatorID)
}

func insertSprintTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, workspaceID uuid.UUID, code string) {
	t.Helper()
	mustSprintExec(t, ctx, pool, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, "Sprints "+code, workspaceID, code)
}

func insertSprintWorkspaceMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID, userID uuid.UUID,
	role string,
) {
	t.Helper()
	mustSprintExec(t, ctx, pool, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST($3 AS user_role))
	`, workspaceID, userID, role)
}

func insertSprintTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	mustSprintExec(t, ctx, pool, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID)
}

func insertSprintObjective(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	objectiveID, workspaceID, teamID, creatorID uuid.UUID,
	name string,
) {
	t.Helper()
	mustSprintExec(t, ctx, pool, `
		INSERT INTO objectives (objective_id, name, workspace_id, team_id, created_by, sequence_id)
		VALUES (
			$1, $2, $3, $4, $5,
			COALESCE((SELECT MAX(sequence_id) + 1 FROM objectives WHERE team_id = $4), 1)
		)
	`, objectiveID, name+" "+uuid.NewString(), workspaceID, teamID, creatorID)
}

func mustSprintExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, arguments ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, query, arguments...); err != nil {
		t.Fatalf("execute sprint fixture SQL: %v", err)
	}
}

func sprintRowCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, arguments ...any) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, query, arguments...).Scan(&count); err != nil {
		t.Fatalf("count sprint fixture rows: %v", err)
	}
	return count
}

func uniqueSprintName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.NewString())
}

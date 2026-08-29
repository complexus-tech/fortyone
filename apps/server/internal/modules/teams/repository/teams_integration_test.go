//go:build integration

package teamsrepository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTeamsEnforceWorkspaceMembershipAndOrderingTransactions(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workspaceA := insertTeamTestWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertTeamTestWorkspace(t, ctx, postgres.Pool, "b")
	adminA := insertTeamTestUser(t, ctx, postgres.Pool, "admin-a", true)
	memberA := insertTeamTestUser(t, ctx, postgres.Pool, "member-a", true)
	memberB := insertTeamTestUser(t, ctx, postgres.Pool, "member-b", true)
	inactiveA := insertTeamTestUser(t, ctx, postgres.Pool, "inactive-a", false)
	insertTeamTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, adminA, "admin")
	insertTeamTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, memberA, "member")
	insertTeamTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, inactiveA, "member")
	insertTeamTestWorkspaceMember(t, ctx, postgres.Pool, workspaceB, memberB, "member")

	publicA := insertTeamTestTeam(t, ctx, postgres.Pool, workspaceA, "Public A", false)
	privateA := insertTeamTestTeam(t, ctx, postgres.Pool, workspaceA, "Private A", true)
	publicB := insertTeamTestTeam(t, ctx, postgres.Pool, workspaceB, "Public B", false)
	insertTeamTestTeamMember(t, ctx, postgres.Pool, privateA, memberA)
	insertTeamTestTeamMember(t, ctx, postgres.Pool, publicA, inactiveA)

	repository := newTeamIntegrationRepository(postgres.Pool)

	memberTeams, err := repository.List(ctx, workspaceA, memberA, teams.CoreListTeamsFilter{})
	if err != nil {
		t.Fatalf("list member teams: %v", err)
	}
	assertTeamIDs(t, memberTeams, privateA)

	adminTeams, err := repository.List(ctx, workspaceA, adminA, teams.CoreListTeamsFilter{})
	if err != nil {
		t.Fatalf("list admin teams: %v", err)
	}
	assertTeamIDs(t, adminTeams, publicA, privateA)

	crossWorkspaceTeams, err := repository.List(ctx, workspaceB, memberA, teams.CoreListTeamsFilter{})
	if err != nil {
		t.Fatalf("list cross-workspace teams: %v", err)
	}
	if len(crossWorkspaceTeams) != 0 {
		t.Fatalf("cross-workspace teams = %#v, want none", crossWorkspaceTeams)
	}

	inactiveTeams, err := repository.List(ctx, workspaceA, inactiveA, teams.CoreListTeamsFilter{})
	if err != nil {
		t.Fatalf("list inactive actor teams: %v", err)
	}
	if len(inactiveTeams) != 0 {
		t.Fatalf("inactive actor teams = %#v, want none", inactiveTeams)
	}

	if _, err := repository.GetByID(ctx, publicB, workspaceB, memberA); !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("cross-workspace get error = %v, want ErrTeamNotFound", err)
	}

	publicTeams, err := repository.ListPublicTeams(ctx, workspaceA, memberA, teams.CoreListTeamsFilter{})
	if err != nil {
		t.Fatalf("list public teams: %v", err)
	}
	assertTeamIDs(t, publicTeams, publicA)

	if err := repository.AddMember(ctx, publicA, memberB, workspaceA); !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("cross-workspace member add error = %v, want ErrTeamNotFound", err)
	}
	assertTeamMembership(t, ctx, postgres.Pool, publicA, memberB, false)

	if err := repository.AddMember(ctx, publicB, memberA, workspaceA); !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("cross-workspace team add error = %v, want ErrTeamNotFound", err)
	}
	assertTeamMembership(t, ctx, postgres.Pool, publicB, memberA, false)

	if err := repository.JoinPublicTeam(ctx, teams.CorePublicTeamJoin{
		TeamID: publicB, ActorID: memberA, WorkspaceID: workspaceA,
	}); !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("cross-workspace join error = %v, want ErrTeamNotFound", err)
	}
	if err := repository.JoinPublicTeam(ctx, teams.CorePublicTeamJoin{
		TeamID: privateA, ActorID: memberA, WorkspaceID: workspaceA,
	}); !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("private join error = %v, want ErrTeamNotFound", err)
	}
	if err := repository.JoinPublicTeam(ctx, teams.CorePublicTeamJoin{
		TeamID: publicA, ActorID: memberA, WorkspaceID: workspaceA,
	}); err != nil {
		t.Fatalf("same-workspace public join: %v", err)
	}
	if err := repository.JoinPublicTeam(ctx, teams.CorePublicTeamJoin{
		TeamID: publicA, ActorID: memberA, WorkspaceID: workspaceA,
	}); !errors.Is(err, teams.ErrTeamMemberExists) {
		t.Fatalf("duplicate join error = %v, want ErrTeamMemberExists", err)
	}

	if err := repository.UpdateMemberAIContext(ctx, privateA, memberA, workspaceB, teams.CoreTeamMemberAIContext{
		RoleTitle: "Cross tenant",
	}); !errors.Is(err, teams.ErrTeamMemberNotFound) {
		t.Fatalf("cross-workspace AI context error = %v, want ErrTeamMemberNotFound", err)
	}
	if err := repository.UpdateMemberAIContext(ctx, publicA, inactiveA, workspaceA, teams.CoreTeamMemberAIContext{
		RoleTitle: "Inactive member",
	}); !errors.Is(err, teams.ErrTeamMemberNotFound) {
		t.Fatalf("inactive-member AI context error = %v, want ErrTeamMemberNotFound", err)
	}
	if err := repository.RemoveMember(ctx, publicA, inactiveA, workspaceA); !errors.Is(err, teams.ErrTeamMemberNotFound) {
		t.Fatalf("inactive-member removal error = %v, want ErrTeamMemberNotFound", err)
	}
	if err := repository.RemoveMember(ctx, privateA, memberA, workspaceB); !errors.Is(err, teams.ErrTeamMemberNotFound) {
		t.Fatalf("cross-workspace remove error = %v, want ErrTeamMemberNotFound", err)
	}
	assertTeamMembership(t, ctx, postgres.Pool, publicA, inactiveA, true)
	assertTeamMembership(t, ctx, postgres.Pool, privateA, memberA, true)

	insertTeamTestOrder(t, ctx, postgres.Pool, memberA, privateA, workspaceA, 7)
	err = repository.UpdateUserTeamOrdering(ctx, memberA, workspaceA, []uuid.UUID{publicA, publicB})
	if !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("cross-workspace ordering error = %v, want ErrTeamNotFound", err)
	}
	assertTeamOrders(t, ctx, postgres.Pool, memberA, workspaceA, []teamOrder{{teamID: privateA, index: 7}})

	if err := repository.UpdateUserTeamOrdering(ctx, memberA, workspaceA, []uuid.UUID{publicA, privateA}); err != nil {
		t.Fatalf("same-workspace ordering: %v", err)
	}
	assertTeamOrders(t, ctx, postgres.Pool, memberA, workspaceA, []teamOrder{
		{teamID: publicA, index: 0},
		{teamID: privateA, index: 1},
	})
}

func TestCreateTeamInitializesDefaultsAndHonorsCallerRollback(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	workspaceID := insertTeamTestWorkspace(t, ctx, postgres.Pool, "defaults")
	repository := newTeamIntegrationRepository(postgres.Pool)

	created, err := repository.Create(ctx, teams.CoreTeam{
		Name: "Platform", Code: "PLT", Color: "#000000", Workspace: workspaceID,
	})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	assertTeamDefaults(t, ctx, postgres.Pool, created.ID, workspaceID, true)

	tx, err := postgres.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin caller transaction: %v", err)
	}
	workspaceTransaction, err := repository.BindWorkspaceTransaction(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("bind workspace transaction: %v", err)
	}
	rolledBack, err := workspaceTransaction.CreateTeam(ctx, WorkspaceTeamInput{
		Name: "Rollback", Code: "RBK", Color: "#111111", Workspace: workspaceID,
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("create team in caller transaction: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("roll back caller transaction: %v", err)
	}
	assertTeamDefaults(t, ctx, postgres.Pool, rolledBack.ID, workspaceID, false)
}

func TestCreateTeamRollsBackWhenRequiredDefaultFails(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	workspaceID := insertTeamTestWorkspace(t, ctx, postgres.Pool, "rollback")
	repository := newTeamIntegrationRepository(postgres.Pool)

	if _, err := postgres.Pool.Exec(ctx, `
		CREATE FUNCTION reject_blocked_team_status() RETURNS trigger AS $$
		BEGIN
			IF NEW.name = 'Blocked' THEN
				RAISE EXCEPTION 'forced default status failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`); err != nil {
		t.Fatalf("install failure function: %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `
		CREATE TRIGGER reject_blocked_team_status
		BEFORE INSERT ON statuses
		FOR EACH ROW EXECUTE FUNCTION reject_blocked_team_status();
	`); err != nil {
		t.Fatalf("install failure trigger: %v", err)
	}

	_, err := repository.Create(ctx, teams.CoreTeam{
		Name: "Must Roll Back", Code: "RBK", Color: "#000000", Workspace: workspaceID,
	})
	if err == nil {
		t.Fatal("create team error = nil, want default-status failure")
	}

	var teamCount, settingsCount, statusCount int
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM teams WHERE workspace_id = $1 AND code = 'RBK'),
			(SELECT COUNT(*) FROM team_story_automation_settings WHERE workspace_id = $1),
			(SELECT COUNT(*) FROM statuses WHERE workspace_id = $1)
	`, workspaceID).Scan(&teamCount, &settingsCount, &statusCount); err != nil {
		t.Fatalf("read rollback state: %v", err)
	}
	if teamCount != 0 || settingsCount != 0 || statusCount != 0 {
		t.Fatalf("rollback state teams/settings/statuses = %d/%d/%d, want zeros", teamCount, settingsCount, statusCount)
	}
}

func newTeamIntegrationRepository(pool *pgxpool.Pool) *repo {
	return New(pool)
}

func insertTeamTestWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, $2, $3)
	`, id, "Teams "+label, "teams-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace %s: %v", label, err)
	}
	return id
}

func insertTeamTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", "Teams "+label, active); err != nil {
		t.Fatalf("insert user %s: %v", label, err)
	}
	return id
}

func insertTeamTestWorkspaceMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID, userID uuid.UUID,
	role string,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
	`, workspaceID, userID, role); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertTeamTestTeam(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	name string,
	private bool,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	code := fmt.Sprintf("T%s", uuid.NewString()[:7])
	if _, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color, is_private)
		VALUES ($1, $2, $3, $4, '#000000', $5)
	`, id, name, workspaceID, code, private); err != nil {
		t.Fatalf("insert team %s: %v", name, err)
	}
	return id
}

func insertTeamTestTeamMember(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
	`, teamID, userID); err != nil {
		t.Fatalf("insert team member: %v", err)
	}
}

func insertTeamTestOrder(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID, teamID, workspaceID uuid.UUID,
	index int,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_team_orders (user_id, team_id, workspace_id, order_index)
		VALUES ($1, $2, $3, $4)
	`, userID, teamID, workspaceID, index); err != nil {
		t.Fatalf("insert team order: %v", err)
	}
}

func assertTeamIDs(t *testing.T, got []teams.CoreTeam, want ...uuid.UUID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("team count = %d, want %d: %#v", len(got), len(want), got)
	}
	wantSet := make(map[uuid.UUID]struct{}, len(want))
	for _, id := range want {
		wantSet[id] = struct{}{}
	}
	for _, team := range got {
		if _, ok := wantSet[team.ID]; !ok {
			t.Fatalf("unexpected team %s in %#v", team.ID, got)
		}
	}
}

func assertTeamMembership(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	teamID, userID uuid.UUID,
	want bool,
) {
	t.Helper()
	var got bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2
		)
	`, teamID, userID).Scan(&got); err != nil {
		t.Fatalf("read team membership: %v", err)
	}
	if got != want {
		t.Fatalf("team membership = %v, want %v", got, want)
	}
}

type teamOrder struct {
	teamID uuid.UUID
	index  int32
}

func assertTeamOrders(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID, workspaceID uuid.UUID,
	want []teamOrder,
) {
	t.Helper()
	rows, err := pool.Query(ctx, `
		SELECT team_id, order_index
		FROM user_team_orders
		WHERE user_id = $1 AND workspace_id = $2
		ORDER BY order_index, team_id
	`, userID, workspaceID)
	if err != nil {
		t.Fatalf("read team orders: %v", err)
	}
	defer rows.Close()
	got := make([]teamOrder, 0, len(want))
	for rows.Next() {
		var order teamOrder
		if err := rows.Scan(&order.teamID, &order.index); err != nil {
			t.Fatalf("scan team order: %v", err)
		}
		got = append(got, order)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate team orders: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("team orders = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("team order %d = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func assertTeamDefaults(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	teamID, workspaceID uuid.UUID,
	want bool,
) {
	t.Helper()
	var teamCount, settingsCount, statusCount int
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM teams WHERE team_id = $1 AND workspace_id = $2),
			(SELECT COUNT(*) FROM team_story_automation_settings WHERE team_id = $1 AND workspace_id = $2),
			(SELECT COUNT(*) FROM statuses WHERE team_id = $1 AND workspace_id = $2)
	`, teamID, workspaceID).Scan(&teamCount, &settingsCount, &statusCount); err != nil {
		t.Fatalf("read team defaults: %v", err)
	}
	if want {
		if teamCount != 1 || settingsCount != 1 || statusCount != len(teams.DefaultStoryStatuses) {
			t.Fatalf("team/default state = %d/%d/%d", teamCount, settingsCount, statusCount)
		}
		return
	}
	if teamCount != 0 || settingsCount != 0 || statusCount != 0 {
		t.Fatalf("rolled-back team/default state = %d/%d/%d, want zeros", teamCount, settingsCount, statusCount)
	}
}

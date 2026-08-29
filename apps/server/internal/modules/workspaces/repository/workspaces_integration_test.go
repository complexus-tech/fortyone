//go:build integration

package workspacesrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestWorkspaceRepositoryEnforcesLiveMembershipAndTenantBoundaries(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	workspaceA := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "b")
	adminA := insertWorkspaceTestUser(t, ctx, postgres.Pool, "admin-a", true)
	memberA := insertWorkspaceTestUser(t, ctx, postgres.Pool, "member-a", true)
	inactiveA := insertWorkspaceTestUser(t, ctx, postgres.Pool, "inactive-a", false)
	memberB := insertWorkspaceTestUser(t, ctx, postgres.Pool, "member-b", true)
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspaceA, adminA, "admin")
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspaceA, memberA, "member")
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspaceA, inactiveA, "member")
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspaceB, memberB, "member")

	repository := New(postgres.Pool)
	membership, err := repository.ResolveCurrentMembership(ctx, workspaceTestSlug(t, ctx, postgres.Pool, workspaceA), adminA)
	if err != nil {
		t.Fatalf("resolve current administrator membership: %v", err)
	}
	if membership.WorkspaceID != workspaceA || membership.Role != "admin" {
		t.Fatalf("administrator membership = %#v", membership)
	}

	tests := []struct {
		name   string
		slug   string
		userID uuid.UUID
	}{
		{name: "wrong tenant", slug: workspaceTestSlug(t, ctx, postgres.Pool, workspaceA), userID: memberB},
		{name: "inactive account", slug: workspaceTestSlug(t, ctx, postgres.Pool, workspaceA), userID: inactiveA},
		{name: "unknown workspace", slug: "unknown-" + uuid.NewString(), userID: adminA},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := repository.ResolveCurrentMembership(ctx, test.slug, test.userID); !errors.Is(err, workspacedomain.ErrNotFound) {
				t.Fatalf("resolve error = %v, want ErrNotFound", err)
			}
		})
	}

	if err := repository.UpdateMemberRole(ctx, workspaceA, memberA, "guest"); err != nil {
		t.Fatalf("demote member: %v", err)
	}
	membership, err = repository.ResolveCurrentMembership(ctx, workspaceTestSlug(t, ctx, postgres.Pool, workspaceA), memberA)
	if err != nil {
		t.Fatalf("resolve demoted membership: %v", err)
	}
	if membership.Role != "guest" {
		t.Fatalf("demoted role = %q, want guest", membership.Role)
	}
	if _, err := postgres.Pool.Exec(ctx, `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, workspaceA, memberA); err != nil {
		t.Fatalf("revoke membership: %v", err)
	}
	if _, err := repository.ResolveCurrentMembership(ctx, workspaceTestSlug(t, ctx, postgres.Pool, workspaceA), memberA); !errors.Is(err, workspacedomain.ErrNotFound) {
		t.Fatalf("resolve revoked membership error = %v, want ErrNotFound", err)
	}
}

func TestWorkspaceRepositoryAccessAndMemberRemovalAreMembershipSafe(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	workspaceA := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "access-a")
	workspaceB := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "access-b")
	member := insertWorkspaceTestUser(t, ctx, postgres.Pool, "access-member", true)
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspaceA, member, "member")
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspaceB, member, "member")
	teamA := insertWorkspaceTestTeam(t, ctx, postgres.Pool, workspaceA, "team-a")
	teamB := insertWorkspaceTestTeam(t, ctx, postgres.Pool, workspaceB, "team-b")
	insertWorkspaceTestTeamMembership(t, ctx, postgres.Pool, teamA, member)
	insertWorkspaceTestTeamMembership(t, ctx, postgres.Pool, teamB, member)

	repository := New(postgres.Pool)
	old := time.Now().Add(-48 * time.Hour).UTC().Truncate(time.Microsecond)
	setWorkspaceAccessTimes(t, ctx, postgres.Pool, workspaceA, member, old)
	if err := repository.RecordAccess(ctx, workspaceA, member); err != nil {
		t.Fatalf("record current member access: %v", err)
	}
	assertWorkspaceAccessTouched(t, ctx, postgres.Pool, workspaceA, member, old, true)

	if err := repository.RemoveMember(ctx, workspaceA, member); err != nil {
		t.Fatalf("remove member: %v", err)
	}
	assertWorkspaceMembership(t, ctx, postgres.Pool, workspaceA, member, false)
	assertWorkspaceTeamMembership(t, ctx, postgres.Pool, teamA, member, false)
	assertWorkspaceTeamMembership(t, ctx, postgres.Pool, teamB, member, true)

	setWorkspaceAccessTimes(t, ctx, postgres.Pool, workspaceA, member, old)
	if err := repository.RecordAccess(ctx, workspaceA, member); err != nil {
		t.Fatalf("record revoked member access: %v", err)
	}
	assertWorkspaceAccessTouched(t, ctx, postgres.Pool, workspaceA, member, old, false)
}

func TestWorkspaceMemberRemovalRollsBackWhenTeamCleanupFails(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	workspaceID := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "remove-rollback")
	memberID := insertWorkspaceTestUser(t, ctx, postgres.Pool, "remove-rollback", true)
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspaceID, memberID, "member")
	teamID := insertWorkspaceTestTeam(t, ctx, postgres.Pool, workspaceID, "remove-rollback")
	insertWorkspaceTestTeamMembership(t, ctx, postgres.Pool, teamID, memberID)
	if _, err := postgres.Pool.Exec(ctx, `
		CREATE FUNCTION reject_test_team_member_delete() RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'forced team membership cleanup failure';
			RETURN OLD;
		END;
		$$ LANGUAGE plpgsql
	`); err != nil {
		t.Fatalf("install team cleanup failure function: %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `
		CREATE TRIGGER reject_test_team_member_delete
		BEFORE DELETE ON team_members
		FOR EACH ROW EXECUTE FUNCTION reject_test_team_member_delete()
	`); err != nil {
		t.Fatalf("install team cleanup failure trigger: %v", err)
	}

	if err := New(postgres.Pool).RemoveMember(ctx, workspaceID, memberID); err == nil {
		t.Fatal("remove member error = nil, want team cleanup failure")
	}
	assertWorkspaceMembership(t, ctx, postgres.Pool, workspaceID, memberID, true)
	assertWorkspaceTeamMembership(t, ctx, postgres.Pool, teamID, memberID, true)
}

func insertWorkspaceTestWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`, id, "Workspace "+label, "workspace-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace %q: %v", label, err)
	}
	return id
}

func insertWorkspaceTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO users (user_id, username, email, full_name, is_active) VALUES ($1, $2, $3, $4, $5)`, id, label+"-"+suffix, label+"-"+suffix+"@example.com", "Workspace "+label, active); err != nil {
		t.Fatalf("insert user %q: %v", label, err)
	}
	return id
}

func insertWorkspaceTestMembership(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, role string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)`, workspaceID, userID, role); err != nil {
		t.Fatalf("insert workspace membership: %v", err)
	}
}

func insertWorkspaceTestTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO teams (team_id, name, code, color, workspace_id) VALUES ($1, $2, $3, '#000000', $4)`, id, "Team "+label, "T"+uuid.NewString()[:7], workspaceID); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func insertWorkspaceTestTeamMembership(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(ctx, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID); err != nil {
		t.Fatalf("insert team membership: %v", err)
	}
}

func workspaceTestSlug(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID) string {
	t.Helper()
	var slug string
	if err := pool.QueryRow(ctx, `SELECT slug FROM workspaces WHERE workspace_id = $1`, workspaceID).Scan(&slug); err != nil {
		t.Fatalf("read workspace slug: %v", err)
	}
	return slug
}

func setWorkspaceAccessTimes(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, timestamp time.Time) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		UPDATE users
		SET last_login_at = CAST($2 AS timestamptz),
		    inactivity_warning_sent_at = CAST($2 AS timestamp)
		WHERE user_id = $1
	`, userID, timestamp); err != nil {
		t.Fatalf("set user access timestamps: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE workspaces SET last_accessed_at = $2, inactivity_warning_sent_at = $2 WHERE workspace_id = $1`, workspaceID, timestamp); err != nil {
		t.Fatalf("set workspace access timestamps: %v", err)
	}
}

func assertWorkspaceMembership(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, want bool) {
	t.Helper()
	var got bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2)`, workspaceID, userID).Scan(&got); err != nil {
		t.Fatalf("read workspace membership: %v", err)
	}
	if got != want {
		t.Fatalf("workspace membership = %t, want %t", got, want)
	}
}

func assertWorkspaceTeamMembership(t *testing.T, ctx context.Context, pool *pgxpool.Pool, teamID, userID uuid.UUID, want bool) {
	t.Helper()
	var got bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)`, teamID, userID).Scan(&got); err != nil {
		t.Fatalf("read team membership: %v", err)
	}
	if got != want {
		t.Fatalf("team membership = %t, want %t", got, want)
	}
}

func assertWorkspaceAccessTouched(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, old time.Time, wantTouched bool) {
	t.Helper()
	var userAccess, workspaceAccess time.Time
	var userWarning, workspaceWarning *time.Time
	if err := pool.QueryRow(ctx, `
		SELECT account.last_login_at, account.inactivity_warning_sent_at,
		       workspace.last_accessed_at, workspace.inactivity_warning_sent_at
		FROM users AS account CROSS JOIN workspaces AS workspace
		WHERE account.user_id = $1 AND workspace.workspace_id = $2
	`, userID, workspaceID).Scan(&userAccess, &userWarning, &workspaceAccess, &workspaceWarning); err != nil {
		t.Fatalf("read access timestamps: %v", err)
	}
	if wantTouched {
		if !userAccess.After(old) || !workspaceAccess.After(old) || userWarning != nil || workspaceWarning != nil {
			t.Fatalf("touched access = user %s/%v workspace %s/%v", userAccess, userWarning, workspaceAccess, workspaceWarning)
		}
		return
	}
	if !userAccess.Equal(old) || !workspaceAccess.Equal(old) || userWarning == nil || workspaceWarning == nil {
		t.Fatalf("revoked access changed = user %s/%v workspace %s/%v, want unchanged", userAccess, userWarning, workspaceAccess, workspaceWarning)
	}
}

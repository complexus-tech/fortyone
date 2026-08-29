//go:build integration

package githubrepository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestGetWorkspaceRoleUsesCurrentWorkspaceMembership(t *testing.T) {
	testDatabase := testkit.NewPostgres(t)
	ctx := t.Context()
	if err := testDatabase.Pool.Ping(ctx); err != nil {
		t.Fatalf("ping isolated GitHub authorization database: %v", err)
	}

	workspaceA, workspaceB := uuid.New(), uuid.New()
	guestID, memberID, adminID := uuid.New(), uuid.New(), uuid.New()
	suffix := uuid.NewString()
	members := []struct {
		id   uuid.UUID
		name string
		role authorization.WorkspaceRole
	}{
		{id: guestID, name: "guest", role: authorization.WorkspaceRoleGuest},
		{id: memberID, name: "member", role: authorization.WorkspaceRoleMember},
		{id: adminID, name: "admin", role: authorization.WorkspaceRoleAdmin},
	}
	for _, member := range members {
		if _, err := testDatabase.Pool.Exec(ctx, `
			INSERT INTO users (user_id, username, email, full_name)
			VALUES ($1, $2, $3, $4)
		`, member.id, "github-role-"+member.name+"-"+suffix, "github-role-"+member.name+"-"+suffix+"@example.com", "GitHub role test"); err != nil {
			t.Fatalf("insert GitHub authorization user: %v", err)
		}
	}
	if _, err := testDatabase.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'GitHub authorization A', $2, $3),
		       ($4, 'GitHub authorization B', $5, $3)
	`, workspaceA, "github-role-a-"+suffix, adminID, workspaceB, "github-role-b-"+suffix); err != nil {
		t.Fatalf("insert GitHub authorization workspaces: %v", err)
	}
	if _, err := testDatabase.Pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'guest'), ($1, $3, 'member'), ($1, $4, 'admin')
	`, workspaceA, guestID, memberID, adminID); err != nil {
		t.Fatalf("insert GitHub authorization memberships: %v", err)
	}

	repository := New(testDatabase.Pool)
	for _, member := range members {
		t.Run(member.name, func(t *testing.T) {
			role, err := repository.GetWorkspaceRole(ctx, workspaceA, member.id)
			if err != nil {
				t.Fatalf("GetWorkspaceRole() error = %v", err)
			}
			if role != member.role {
				t.Fatalf("role = %q, want %q", role, member.role)
			}
		})
	}

	if _, err := repository.GetWorkspaceRole(ctx, workspaceB, adminID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-workspace role lookup error = %v, want sql.ErrNoRows", err)
	}
}

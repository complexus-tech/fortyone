//go:build integration

package usersrepository

import (
	"context"
	"fmt"
	"testing"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertUserTestWorkspace(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	label string,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, $2, $3)
	`, id, "Users "+label, "users-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace %s: %v", label, err)
	}
	return id
}

func insertUserTestAccount(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	label string,
	active bool,
	system bool,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, label+"-"+id.String(), userTestEmail(label, id), "Users "+label, active, system); err != nil {
		t.Fatalf("insert user %s: %v", label, err)
	}
	return id
}

func userTestEmail(label string, userID uuid.UUID) string {
	return fmt.Sprintf("%s-%s@example.com", label, userID)
}

func insertUserTestWorkspaceMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	role string,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST($3 AS user_role))
	`, workspaceID, userID, role); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertUserTestTeam(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	name string,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, name, workspaceID, "U"+uuid.NewString()[:7]); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func insertUserTestTeamMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	teamID uuid.UUID,
	userID uuid.UUID,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
	`, teamID, userID); err != nil {
		t.Fatalf("insert team member: %v", err)
	}
}

func assertUserIDs(t *testing.T, actual []users.CoreUser, expected ...uuid.UUID) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("user count = %d, want %d: %#v", len(actual), len(expected), actual)
	}
	for _, expectedID := range expected {
		found := false
		for _, user := range actual {
			if user.ID == expectedID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("users %#v do not contain %s", actual, expectedID)
		}
	}
}

func containsSystemUser(usersList []users.CoreUser) bool {
	for _, user := range usersList {
		if user.IsSystem {
			return true
		}
	}
	return false
}

func containsInactiveUser(usersList []users.CoreUser) bool {
	for _, user := range usersList {
		if !user.IsActive {
			return true
		}
	}
	return false
}

func assertLastUsedWorkspace(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	want *uuid.UUID,
) {
	t.Helper()
	var got *uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT last_used_workspace_id FROM users WHERE user_id = $1`, userID).Scan(&got); err != nil {
		t.Fatalf("read last used workspace: %v", err)
	}
	if want == nil && got == nil {
		return
	}
	if want == nil || got == nil || *want != *got {
		t.Fatalf("last used workspace = %v, want %v", got, want)
	}
}

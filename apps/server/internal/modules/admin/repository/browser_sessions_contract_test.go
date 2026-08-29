package adminrepository

import (
	"os"
	"strings"
	"testing"
)

func TestAdminUserStateTransitionsFenceSessionsAndOwnReactivationPolicy(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/mutations.sql")
	if err != nil {
		t.Fatalf("read admin mutations: %v", err)
	}
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")
	for _, clause := range []string{
		"login_reactivation_policy = case",
		"then 'verified_sign_in'",
		"then 'admin_only'",
		"and is_active <> cast(sqlc.arg(new_is_active) as boolean) then auth_session_version + 1",
		"and is_active = false",
		"and cast(sqlc.arg(new_is_active) as boolean) = true then null",
		"-- name: revokeadminuserbrowsersessions :one",
		"auth_session_version = auth_session_version + 1",
		"returning auth_session_version",
	} {
		if !strings.Contains(query, clause) {
			t.Fatalf("admin browser-session mutation contract is missing %q", clause)
		}
	}
}

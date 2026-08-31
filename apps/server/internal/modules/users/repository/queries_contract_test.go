package usersrepository

import (
	"os"
	"strings"
	"testing"
)

func TestUserQueriesRetainIdentityAndTenantSecurityBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		file     string
		required []string
	}{
		{
			name: "workspace user list",
			file: "queries/accounts.sql",
			required: []string{
				"workspace_membership.workspace_id = sqlc.arg(workspace_id)",
				"selected_team.workspace_id = workspace_membership.workspace_id",
				"account.is_active = TRUE",
				"account.is_system = FALSE",
			},
		},
		{
			name: "private memories",
			file: "queries/memories.sql",
			required: []string{
				"memory.user_id = sqlc.arg(user_id)",
				"memory.workspace_id = sqlc.arg(workspace_id)",
				"membership.workspace_id = memory.workspace_id",
				"account.is_active = TRUE",
			},
		},
		{
			name: "automation preferences",
			file: "queries/automation_preferences.sql",
			required: []string{
				"membership.user_id = sqlc.arg(user_id)",
				"membership.workspace_id = sqlc.arg(workspace_id)",
				"account.is_active = TRUE",
				"ON CONFLICT (user_id, workspace_id)",
			},
		},
		{
			name: "onboarding tour progress",
			file: "queries/onboarding_tour_progress.sql",
			required: []string{
				"membership.user_id = sqlc.arg(user_id)",
				"membership.workspace_id = sqlc.arg(workspace_id)",
				"account.is_active = TRUE",
				"ON CONFLICT (user_id, workspace_id, tour_key, tour_version)",
			},
		},
		{
			name: "verification token one-time use",
			file: "queries/verification_tokens.sql",
			required: []string{
				"verification_token.used_at IS NULL",
				"verification_token.expires_at > CAST(sqlc.arg(consumed_at) AS timestamptz)",
				"FOR UPDATE",
				"digest_candidate.token_digest = verification_token.token_digest",
				"digest_candidate.token_key_id = verification_token.token_key_id",
				"digest_candidate.token_version = verification_token.token_version",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(test.file)
			if err != nil {
				t.Fatalf("read %s: %v", test.file, err)
			}
			for _, clause := range test.required {
				if !strings.Contains(string(source), clause) {
					t.Fatalf("%s is missing security clause %q", test.file, clause)
				}
			}
		})
	}
}

func TestUserQueryFilesContainNoWildcardProjection(t *testing.T) {
	t.Parallel()

	for _, file := range []string{
		"queries/accounts.sql",
		"queries/automation_preferences.sql",
		"queries/external_identities.sql",
		"queries/memories.sql",
		"queries/onboarding_tour_progress.sql",
		"queries/verification_tokens.sql",
	} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(strings.ToUpper(string(source)), "SELECT *") {
			t.Fatalf("%s contains a wildcard projection", file)
		}
	}
}

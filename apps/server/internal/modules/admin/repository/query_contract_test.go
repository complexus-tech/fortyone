package adminrepository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminApplicationSQLUsesStaticTypedQueries(t *testing.T) {
	entries, err := filepath.Glob("queries/*.sql")
	if err != nil || len(entries) == 0 {
		t.Fatalf("discover admin SQLC queries: %v", err)
	}
	for _, path := range entries {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		query := string(raw)
		if strings.Contains(query, "::") {
			t.Errorf("%s uses PostgreSQL shorthand cast; use CAST(value AS type)", path)
		}
		if strings.Contains(strings.ToUpper(query), "SELECT *") {
			t.Errorf("%s contains SELECT *", path)
		}
	}
}

func TestAdminListsHaveUniqueTieBreakers(t *testing.T) {
	raw, err := os.ReadFile("queries/reads.sql")
	if err != nil {
		t.Fatal(err)
	}
	queries := string(raw)
	for _, ordering := range []string{
		"ORDER BY workspace.created_at DESC, workspace.workspace_id DESC",
		"ORDER BY target.created_at DESC, target.user_id DESC",
	} {
		if !strings.Contains(queries, ordering) {
			t.Errorf("reads.sql missing stable ordering %q", ordering)
		}
	}

	auditRaw, err := os.ReadFile("queries/audit.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, ordering := range []string{
		"ORDER BY audit.created_at DESC, audit.id DESC",
		"ORDER BY note.created_at DESC, note.id DESC",
	} {
		if !strings.Contains(string(auditRaw), ordering) {
			t.Errorf("audit.sql missing stable ordering %q", ordering)
		}
	}
}

func TestAdminAIUsageResetPreservesHistoryAndAdvancesTheBaseline(t *testing.T) {
	raw, err := os.ReadFile("queries/mutations.sql")
	if err != nil {
		t.Fatal(err)
	}
	query := strings.Join(strings.Fields(string(raw)), " ")
	for _, contract := range []string{
		"INSERT INTO user_ai_usage_resets",
		"usage.raw_message_count - COALESCE(reset.baseline_message_count, 0)",
		"ON CONFLICT (user_id, workspace_id, period_start) DO UPDATE",
		"baseline_message_count = EXCLUDED.baseline_message_count",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("AI usage reset query is missing contract %q", contract)
		}
	}
	if strings.Contains(query, "DELETE FROM chat_") {
		t.Fatal("AI usage reset must not delete chat history")
	}
}

func TestAdminIntegrationReadsExcludeProviderCredentials(t *testing.T) {
	raw, err := os.ReadFile("queries/integrations.sql")
	if err != nil {
		t.Fatal(err)
	}
	query := strings.ToLower(string(raw))
	for _, secretColumn := range []string{
		"bot_access_token",
		"token_payload",
		"passcode_hash",
		"permissions",
		"scopes",
	} {
		if strings.Contains(query, secretColumn) {
			t.Errorf("admin integration query exposes credential field %q", secretColumn)
		}
	}
	for _, requiredSource := range []string{
		"slack_workspaces",
		"slack_user_links",
		"github_installations",
		"calendar_connections",
		"figma_connections",
	} {
		if !strings.Contains(query, requiredSource) {
			t.Errorf("admin integration query is missing provider source %q", requiredSource)
		}
	}
}

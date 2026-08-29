package objectivesrepository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectiveApplicationSQLIsStaticTenantScopedAndUsesCast(t *testing.T) {
	t.Parallel()

	entries, err := filepath.Glob(filepath.Join("queries", "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	for _, path := range entries {
		contents, err := os.ReadFile(path)
		require.NoError(t, err)
		sql := string(contents)
		require.NotContains(t, sql, "::", path)
		require.Contains(t, sql, "workspace_id", path)
		require.NotContains(t, strings.ToLower(sql), "select *", path)
	}
}

func TestObjectiveProgressChartGuardsUUIDCastsAndScopesActivities(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile(filepath.Join("queries", "analytics.sql"))
	require.NoError(t, err)
	query := string(contents)
	for _, contract := range []string{
		"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
		"activity.workspace_id = story.workspace_id",
	} {
		require.Contains(t, query, contract)
	}
}

func TestObjectiveMutationSQLRechecksCurrentAuthorizationAndReferences(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile(filepath.Join("queries", "mutations.sql"))
	require.NoError(t, err)
	query := string(contents)
	for _, contract := range []string{
		"membership.role IN ('member', 'admin')",
		"team.workspace_id = sqlc.arg(workspace_id)",
		"status.workspace_id = objective.workspace_id",
		"membership.workspace_id = objective.workspace_id",
		"contributor_workspace.workspace_id = objective.workspace_id",
		"actor_workspace.user_id = sqlc.arg(actor_id)",
	} {
		require.Contains(t, query, contract)
	}
}

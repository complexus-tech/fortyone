package objectivesrepository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuildObjectiveUpdateStatementScopesAndOrdersUpdates(t *testing.T) {
	objectiveID := uuid.New()
	workspaceID := uuid.New()

	query, params := buildObjectiveUpdateStatement(objectiveID, workspaceID, map[string]any{
		"priority": "high",
		"name":     "Grow enterprise adoption",
	})

	require.Equal(t,
		"UPDATE objectives SET name = :name, priority = :priority, updated_at = NOW() WHERE objective_id = :id AND workspace_id = :workspace_id",
		query,
	)
	require.Equal(t, objectiveID, params["id"])
	require.Equal(t, workspaceID, params["workspace_id"])
	require.Equal(t, "Grow enterprise adoption", params["name"])
	require.Equal(t, "high", params["priority"])
}

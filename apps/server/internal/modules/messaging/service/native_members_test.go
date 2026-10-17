package messaging

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNativeMemberDiscoveryBindsCurrentTeamWithoutModelRoutingID(t *testing.T) {
	t.Parallel()
	for _, teamValue := range []string{"", `,"team_id":null`} {
		executor, scope, _, members, _ := nativeFixture(t)
		output, err := executor.Execute(t.Context(), scope, ToolCall{Name: toolListTeamMembers,
			Arguments: json.RawMessage(`{"query":null,"limit":10` + teamValue + `}`)})
		require.NoError(t, err)
		require.Equal(t, scope.SourceScope.TeamID, *members.lastListCall().filter.TeamID)
		require.Contains(t, string(output), members.items[0].ID.String())
	}
}

func TestNativeMemberDiscoveryRejectsExplicitForeignOrInvalidTeams(t *testing.T) {
	t.Parallel()
	for _, teamID := range []string{uuid.NewString(), "PRO", ""} {
		executor, scope, _, _, _ := nativeFixture(t)
		arguments, err := json.Marshal(map[string]any{"team_id": teamID, "query": nil, "limit": 10})
		require.NoError(t, err)
		_, err = executor.Execute(t.Context(), scope, ToolCall{Name: toolListTeamMembers, Arguments: arguments})
		require.ErrorIs(t, err, ErrTeamNotAccessible)
	}
	executor, scope, _, members, _ := nativeFixture(t)
	arguments, err := json.Marshal(map[string]any{"team_id": scope.SourceScope.TeamID, "query": nil, "limit": 10})
	require.NoError(t, err)
	_, err = executor.Execute(t.Context(), scope, ToolCall{Name: toolListTeamMembers, Arguments: arguments})
	require.NoError(t, err)
	require.Equal(t, scope.SourceScope.TeamID, *members.lastListCall().filter.TeamID)
	scope.SourceScope = nil
	_, err = executor.Execute(t.Context(), scope, ToolCall{Name: toolListTeamMembers, Arguments: json.RawMessage(`{"query":null,"limit":10}`)})
	require.ErrorIs(t, err, ErrInvalidToolArguments, "provider calls still require an explicit team ID")
}

func TestNativeMemberToolSchemaLeavesProviderContractUnchanged(t *testing.T) {
	t.Parallel()
	executor, scope, _, _, _ := nativeFixture(t)
	definitions := executor.Definitions()
	native := sourceToolDefinitions(definitions, scope.SourceScope)
	for _, definition := range native {
		if definition.Name != toolListTeamMembers {
			continue
		}
		require.NotContains(t, definition.Parameters["properties"], "team_id")
		require.NoError(t, validateStrictObjectSchema(definition.Parameters, definition.Name))
	}
	for _, definition := range definitions {
		if definition.Name == toolListTeamMembers {
			require.Contains(t, definition.Parameters["properties"], "team_id")
		}
	}
}

package slackrepository

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListWorkspaceTeamsForUserQueryUsesPersonalOrdering(t *testing.T) {
	t.Parallel()

	query := normalizeQuery(listWorkspaceTeamsForUserQuery)
	wantOrdering := normalizeQuery(`
		ORDER BY
			CASE WHEN uto.order_index IS NOT NULL THEN 0 ELSE 1 END,
			uto.order_index ASC NULLS LAST,
			t.created_at DESC,
			t.team_id ASC
	`)
	if !strings.Contains(query, wantOrdering) {
		t.Fatalf("Slack team query ordering = %q, want %q", query, wantOrdering)
	}
	if !strings.Contains(query, normalizeQuery(`
		LEFT JOIN user_team_orders uto ON uto.team_id = t.team_id
			AND uto.user_id = $2
			AND uto.workspace_id = $1
	`)) {
		t.Fatalf("Slack team query does not scope personal ordering to the actor and workspace: %q", query)
	}
}

func TestListWorkspaceTeamsForUserQueryRetainsMembershipAuthorization(t *testing.T) {
	t.Parallel()

	query := normalizeQuery(listWorkspaceTeamsForUserQuery)
	for _, clause := range []string{
		"JOIN team_members tm ON tm.team_id = t.team_id",
		"JOIN workspace_members wm ON wm.workspace_id = t.workspace_id AND wm.user_id = tm.user_id",
		"JOIN users u ON u.user_id = tm.user_id",
		"WHERE t.workspace_id = $1",
		"AND tm.user_id = $2",
		"AND u.is_active = true",
	} {
		if !strings.Contains(query, clause) {
			t.Fatalf("Slack team query is missing authorization clause %q: %q", clause, query)
		}
	}
}

func normalizeQuery(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

func TestUpsertAgentSettingsRejectsOversizedGuidanceBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()

	repo := &Repo{}
	_, err := repo.UpsertAgentSettings(context.Background(), uuid.New(), AgentSettingsInput{
		AssistantEnabled:       true,
		WorkflowActionsEnabled: true,
		Guidance:               strings.Repeat("a", MaxSlackAgentGuidanceRunes+1),
	})

	require.ErrorContains(t, err, "4000 characters or fewer")
}

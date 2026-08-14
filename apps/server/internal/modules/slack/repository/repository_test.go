package slackrepository

import (
	"context"
	"database/sql"
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

func TestAuthorizedAssistantChannelTeamScopeQueryEnforcesSafeV1Boundary(t *testing.T) {
	t.Parallel()

	query := normalizeQuery(authorizedAssistantChannelTeamScopeQuery)
	for _, clause := range []string{
		"FROM slack_channel_team_access access",
		"JOIN slack_channels configured_channel ON configured_channel.workspace_id = access.workspace_id AND configured_channel.slack_workspace_id = access.slack_workspace_id AND configured_channel.slack_channel_id = access.slack_channel_id AND configured_channel.is_active = true AND configured_channel.is_assistant_configured = true",
		"JOIN teams mapped_team ON mapped_team.team_id = access.team_id AND mapped_team.workspace_id = access.workspace_id AND mapped_team.is_private = false",
		"SELECT EXISTS (SELECT 1 FROM configured_public_teams) AS is_configured",
		"configuration.is_configured AS explicitly_mapped",
		"JOIN team_members membership ON membership.team_id = team.team_id AND membership.user_id = $4",
		"JOIN workspace_members workspace_membership ON workspace_membership.workspace_id = team.workspace_id AND workspace_membership.user_id = membership.user_id",
		"JOIN users actor ON actor.user_id = membership.user_id AND actor.is_active = true",
		"AND team.is_private = false",
		"configuration.is_configured AND team.team_id IN (SELECT team_id FROM configured_public_teams)",
		"NOT configuration.is_configured",
	} {
		if !strings.Contains(query, clause) {
			t.Fatalf("Slack assistant audience query is missing authorization clause %q: %q", clause, query)
		}
	}
}

func TestAssistantChannelAudienceQueriesUseUnifiedChannelAccess(t *testing.T) {
	t.Parallel()

	queries := map[string]string{
		"authorize": authorizedAssistantChannelTeamScopeQuery,
		"delete":    deleteAssistantPublicChannelTeamAccessQuery,
		"insert":    insertAssistantChannelTeamAccessQuery,
		"list":      listAssistantChannelTeamAccessQuery,
	}
	for name, rawQuery := range queries {
		query := normalizeQuery(rawQuery)
		require.Contains(t, query, "slack_channel_team_access", name)
		require.NotContains(t, query, "slack_channel_assistant_team_access", name)
	}

	require.Contains(t, normalizeQuery(listAssistantChannelTeamAccessQuery), "AND team.is_private = false")
	require.Contains(t, normalizeQuery(insertAssistantChannelTeamAccessQuery), "AND team.is_private = false")
	require.Equal(t, normalizeQuery(`
		DELETE FROM slack_channel_team_access access
		USING teams team
		WHERE access.workspace_id = $1
		  AND access.slack_workspace_id = $2
		  AND access.slack_channel_id = $3
		  AND team.team_id = access.team_id
		  AND team.workspace_id = access.workspace_id
		  AND team.is_private = false
	`), normalizeQuery(deleteAssistantPublicChannelTeamAccessQuery))
}

type assistantChannelTeamAccessTxStub struct {
	execCalls int
	getCalls  int
}

func (s *assistantChannelTeamAccessTxStub) ExecContext(
	context.Context,
	string,
	...any,
) (sql.Result, error) {
	s.execCalls++
	return assistantChannelTeamAccessResult(1), nil
}

func (s *assistantChannelTeamAccessTxStub) GetContext(
	context.Context,
	any,
	string,
	...any,
) error {
	s.getCalls++
	return nil
}

type assistantChannelTeamAccessResult int64

func (r assistantChannelTeamAccessResult) LastInsertId() (int64, error) { return 0, nil }
func (r assistantChannelTeamAccessResult) RowsAffected() (int64, error) { return int64(r), nil }

func TestUnconfigureAssistantChannelPreservesUnifiedTeamMappings(t *testing.T) {
	t.Parallel()

	tx := &assistantChannelTeamAccessTxStub{}
	err := replaceAssistantPublicChannelTeamAccessTx(
		context.Background(),
		tx,
		uuid.New(),
		uuid.New(),
		"C1",
		false,
		[]uuid.UUID{uuid.New()},
		uuid.New(),
	)

	require.NoError(t, err)
	require.Zero(t, tx.execCalls, "unconfiguring must return before deleting or inserting mappings")
	require.Zero(t, tx.getCalls, "unconfiguring must not inspect or rewrite team mappings")
}

func TestAssistantChannelTeamScopeSharesOnlyExplicitMappings(t *testing.T) {
	t.Parallel()

	defaultPublicTeamID := uuid.New()
	explicitPublicTeamID := uuid.New()
	scope := assistantChannelTeamScope([]assistantChannelTeamScopeRow{
		{TeamID: defaultPublicTeamID},
		{TeamID: explicitPublicTeamID, ExplicitlyMapped: true},
		{TeamID: uuid.Nil, ExplicitlyMapped: true},
	})

	require.Equal(t, []uuid.UUID{defaultPublicTeamID, explicitPublicTeamID}, scope.AllowedTeamIDs)
	require.Equal(t, []uuid.UUID{explicitPublicTeamID}, scope.SharedTeamIDs)
}

func normalizeQuery(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

func TestUpsertAgentSettingsRejectsOversizedGuidanceBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()

	repo := &Repo{}
	_, err := repo.UpsertAgentSettings(context.Background(), uuid.New(), AgentSettingsInput{
		Guidance: strings.Repeat("a", MaxSlackAgentGuidanceRunes+1),
	})

	require.ErrorContains(t, err, "4000 characters or fewer")
}

package storiesrepository

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSecondaryMutationQueriesLockAndAuthorizeEveryTenantTarget(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("queries/secondary_mutations.sql")
	require.NoError(t, err)
	queries := string(data)

	for _, contract := range []string{
		"story.id = ANY(CAST(sqlc.arg(story_ids) AS uuid[]))",
		"story.workspace_id = sqlc.arg(workspace_id)",
		"ORDER BY story.id",
		"FOR UPDATE",
		"workspace_member.workspace_id = target.workspace_id",
		"team_member.team_id = target.team_id",
		"credential.revoked_at IS NULL",
		"credential.expires_at > sqlc.arg(now)",
		"credential_scope.scope = 'stories:write'",
		"restriction.team_id = target.team_id",
	} {
		require.Contains(t, queries, contract)
	}
}

func TestSecondaryDeleteQueriesPreserveDesiredStateAndHardDeleteScope(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("queries/secondary_mutations.sql")
	require.NoError(t, err)
	queries := string(data)
	interactiveData, err := os.ReadFile("queries/interactive_hard_delete.sql")
	require.NoError(t, err)
	interactiveQueries := string(interactiveData)
	retentionData, err := os.ReadFile("queries/retention.sql")
	require.NoError(t, err)
	retentionQueries := string(retentionData)

	require.Contains(t, queries, "story.deleted_at IS NULL")
	require.Contains(t, queries, "-- name: HardDeleteSecondaryStories :many")
	require.Contains(t, interactiveQueries, "story.workspace_id = sqlc.arg(workspace_id)")
	require.Contains(t, interactiveQueries, "attachment.workspace_id = story.workspace_id")
	require.Contains(t, retentionQueries, "document_relation.attachment_id = attachment.attachment_id")
}

func TestOrderSecondarySubsetPreservesRequestOrder(t *testing.T) {
	t.Parallel()
	first := uuid.New()
	second := uuid.New()
	third := uuid.New()
	got := orderSecondarySubset([]uuid.UUID{first, second, third}, []uuid.UUID{third, first})
	require.Equal(t, []uuid.UUID{first, third}, got)
}

func TestSecondaryMutationSourceInsertsEventsAfterStateChange(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("secondary_mutations.go")
	require.NoError(t, err)
	source := string(data)
	mutation := strings.Index(source, "applySecondaryLifecycleState(")
	event := strings.Index(source[mutation:], "insertMutationEvent(")
	require.Greater(t, mutation, -1)
	require.Greater(t, event, 0)
}

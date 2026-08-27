package jobs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPurgeDeletedChatSessionsPreservesUnresolvedMutationApprovals(t *testing.T) {
	query := strings.ToLower(strings.Join(strings.Fields(purgeDeletedChatSessionsQuery), " "))

	require.Contains(t, query, "delete from chat_sessions as session")
	require.Contains(t, query, "not exists")
	require.Contains(t, query, "execution.session_id = session.id")
	require.Contains(
		t,
		query,
		"execution.status in ('ready', 'retry_ready', 'executing', 'failed_uncertain')",
	)
}

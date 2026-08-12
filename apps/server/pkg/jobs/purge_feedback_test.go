package jobs

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFeedbackContributorsToPurgeQueryLocksItemRowsWithoutDistinct(t *testing.T) {
	query := feedbackContributorsToPurgeQuery()

	require.Contains(t, query, "SELECT contributor_id")
	require.Contains(t, query, "FOR UPDATE")
	require.NotContains(t, strings.ToUpper(query), "DISTINCT")
}

func TestUniqueUUIDsPreservesFirstOccurrence(t *testing.T) {
	first := uuid.New()
	second := uuid.New()

	require.Equal(t, []uuid.UUID{first, second}, uniqueUUIDs([]uuid.UUID{
		first,
		first,
		second,
		first,
	}))
}

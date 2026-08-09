package slack

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAssistantAudienceFingerprintIsCanonicalAndSetScoped(t *testing.T) {
	t.Parallel()

	teamA := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	teamB := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")

	canonical := assistantAudienceFingerprint([]uuid.UUID{teamA, teamB})
	reordered := assistantAudienceFingerprint([]uuid.UUID{teamB, uuid.Nil, teamA, teamB})
	different := assistantAudienceFingerprint([]uuid.UUID{teamA})

	require.Equal(t, canonical, reordered)
	require.NotEqual(t, canonical, different)
	require.Regexp(t, `^v1:[0-9a-f]{64}$`, canonical)
}

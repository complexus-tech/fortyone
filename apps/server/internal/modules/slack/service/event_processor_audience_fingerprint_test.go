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

	canonical := assistantAudienceFingerprint([]uuid.UUID{teamA, teamB}, []uuid.UUID{teamA})
	reordered := assistantAudienceFingerprint(
		[]uuid.UUID{teamB, uuid.Nil, teamA, teamB},
		[]uuid.UUID{teamA, uuid.Nil, teamA},
	)
	differentAllowed := assistantAudienceFingerprint([]uuid.UUID{teamA}, []uuid.UUID{teamA})
	differentShared := assistantAudienceFingerprint([]uuid.UUID{teamA, teamB}, []uuid.UUID{teamB})

	require.Equal(t, canonical, reordered)
	require.NotEqual(t, canonical, differentAllowed)
	require.NotEqual(t, canonical, differentShared)
	require.Regexp(t, `^v2:[0-9a-f]{64}$`, canonical)
}

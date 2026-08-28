package emailreply

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestValidateAuthorizedTeamIDsRejectsMalformedOrUnboundedScopes(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	require.NoError(t, validateAuthorizedTeamIDs([]uuid.UUID{teamID}))
	require.ErrorIs(t, validateAuthorizedTeamIDs([]uuid.UUID{uuid.Nil}), ErrActionUnauthorized)
	require.ErrorIs(t, validateAuthorizedTeamIDs([]uuid.UUID{teamID, teamID}), ErrActionUnauthorized)

	tooMany := make([]uuid.UUID, maximumAuthorizedTeams+1)
	for index := range tooMany {
		tooMany[index] = uuid.New()
	}
	require.ErrorIs(t, validateAuthorizedTeamIDs(tooMany), ErrActionUnauthorized)
}

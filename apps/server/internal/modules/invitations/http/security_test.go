package invitationshttp

import (
	"errors"
	"net/http"
	"testing"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/stretchr/testify/require"
)

func TestPublicInvitationBearerErrorsAreEnumerationSafe(t *testing.T) {
	t.Parallel()

	for _, internalError := range []error{
		invitations.ErrInvitationNotFound,
		invitations.ErrInvitationExpired,
		invitations.ErrInvitationUsed,
		invitations.ErrInvitationRevoked,
		fmtWrappedInvitationError{invitations.ErrInvitationExpired},
	} {
		status, handled, publicError := publicInvitationBearerError(internalError)
		require.True(t, handled)
		require.ErrorIs(t, publicError, invitations.ErrInvitationNotFound)
		require.Equal(t, http.StatusNotFound, status)
	}

	status, handled, publicError := publicInvitationBearerError(errors.New("database unavailable"))
	require.False(t, handled)
	require.NoError(t, publicError)
	require.Zero(t, status)
}

type fmtWrappedInvitationError struct{ error }

func (e fmtWrappedInvitationError) Error() string { return "wrapped: " + e.error.Error() }
func (e fmtWrappedInvitationError) Unwrap() error { return e.error }

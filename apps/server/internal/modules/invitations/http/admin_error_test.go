package invitationshttp

import (
	"errors"
	"net/http"
	"testing"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
)

func TestInvitationAdminErrorStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		fallback int
		want     int
	}{
		{name: "forbidden", err: authorization.ErrWorkspaceAdminRequired, fallback: http.StatusInternalServerError, want: http.StatusForbidden},
		{name: "invalid role", err: invitations.ErrInvalidInvitationRole, fallback: http.StatusInternalServerError, want: http.StatusBadRequest},
		{name: "invalid email", err: invitations.ErrInvalidInvitationEmail, fallback: http.StatusInternalServerError, want: http.StatusBadRequest},
		{name: "invalid team", err: invitations.ErrInvalidInvitationTeam, fallback: http.StatusInternalServerError, want: http.StatusBadRequest},
		{name: "bulk limit", err: invitations.ErrTooManyInvitations, fallback: http.StatusInternalServerError, want: http.StatusBadRequest},
		{name: "duplicate", err: invitations.ErrDuplicateInvitation, fallback: http.StatusInternalServerError, want: http.StatusConflict},
		{name: "missing invitation", err: invitations.ErrInvitationNotFound, fallback: http.StatusInternalServerError, want: http.StatusNotFound},
		{name: "fallback", err: errors.New("dependency failed"), fallback: http.StatusInternalServerError, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := invitationAdminErrorStatus(test.err, test.fallback); got != test.want {
				t.Fatalf("status = %d, want %d", got, test.want)
			}
		})
	}
}

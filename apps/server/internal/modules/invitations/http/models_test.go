package invitationshttp

import (
	"errors"
	"testing"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
)

func TestAppNewInvitationBulkValidateRoles(t *testing.T) {
	t.Parallel()

	for _, role := range []string{
		invitations.InvitationRoleGuest,
		invitations.InvitationRoleMember,
		invitations.InvitationRoleAdmin,
	} {
		role := role
		t.Run(role, func(t *testing.T) {
			t.Parallel()
			request := AppNewInvitationBulk{Invitations: []AppNewInvitation{{Email: "person@example.com", Role: role}}}
			if err := request.Validate(); err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}

	request := AppNewInvitationBulk{Invitations: []AppNewInvitation{{Email: "person@example.com", Role: "system"}}}
	if err := request.Validate(); !errors.Is(err, invitations.ErrInvalidInvitationRole) {
		t.Fatalf("Validate() error = %v, want ErrInvalidInvitationRole", err)
	}
}

package authorization

import (
	"errors"
	"testing"
)

func TestRequireWorkspaceAdminRoleMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		role WorkspaceRole
		err  error
	}{
		{name: "guest", role: WorkspaceRoleGuest, err: ErrWorkspaceAdminRequired},
		{name: "member", role: WorkspaceRoleMember, err: ErrWorkspaceAdminRequired},
		{name: "admin", role: WorkspaceRoleAdmin},
		{name: "unknown", role: WorkspaceRole("owner"), err: ErrInvalidWorkspaceRole},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := RequireWorkspaceAdmin(test.role)
			if test.err == nil && err != nil {
				t.Fatalf("RequireWorkspaceAdmin(%q) error = %v, want nil", test.role, err)
			}
			if test.err != nil && !errors.Is(err, test.err) {
				t.Fatalf("RequireWorkspaceAdmin(%q) error = %v, want %v", test.role, err, test.err)
			}
		})
	}
}

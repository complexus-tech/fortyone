package mayahttp

import (
	"errors"
	"testing"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
)

func TestRequireRealtimeInternalUserFlagRejectsExternalUser(t *testing.T) {
	err := requireRealtimeInternalUserFlag(users.CoreUser{
		IsInternal: false,
	})

	if !errors.Is(err, ErrMayaRealtimeInternalAccessRequired) {
		t.Fatalf("expected internal access error, got %v", err)
	}
}

func TestRequireRealtimeInternalUserFlagAllowsInternalUser(t *testing.T) {
	err := requireRealtimeInternalUserFlag(users.CoreUser{
		IsInternal: true,
	})

	if err != nil {
		t.Fatalf("expected internal user to be allowed, got %v", err)
	}
}

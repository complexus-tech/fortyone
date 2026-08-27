package usersrepository

import (
	"errors"
	"strings"
	"testing"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
)

func TestUserMemoryMutationsRequireOwnerAndWorkspaceScope(t *testing.T) {
	t.Parallel()

	for name, query := range map[string]string{
		"update": updateUserMemoryQuery,
		"delete": deleteUserMemoryQuery,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			normalized := strings.ToLower(query)
			for _, required := range []string{
				"where id = :id",
				"user_id = :user_id",
				"workspace_id = :workspace_id",
			} {
				if !strings.Contains(normalized, required) {
					t.Fatalf("memory mutation must enforce %q: %s", required, normalized)
				}
			}
		})
	}
}

func TestUserMemoryMutationRejectsMissingOrOutOfScopeRecord(t *testing.T) {
	t.Parallel()

	if err := validateUserMemoryMutation(1); err != nil {
		t.Fatalf("owned memory mutation returned an error: %v", err)
	}
	if err := validateUserMemoryMutation(0); !errors.Is(err, users.ErrMemoryNotFound) {
		t.Fatalf("zero-row mutation error = %v, want ErrMemoryNotFound", err)
	}
}

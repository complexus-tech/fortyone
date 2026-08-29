package githubhttp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/authorization"
)

func TestGitHubAdminErrorStatus(t *testing.T) {
	t.Parallel()

	if status := githubAdminErrorStatus(authorization.ErrWorkspaceAdminRequired, http.StatusInternalServerError); status != http.StatusForbidden {
		t.Fatalf("authorization status = %d, want %d", status, http.StatusForbidden)
	}
	if status := githubAdminErrorStatus(errors.New("provider failed"), http.StatusBadRequest); status != http.StatusBadRequest {
		t.Fatalf("fallback status = %d, want %d", status, http.StatusBadRequest)
	}
}

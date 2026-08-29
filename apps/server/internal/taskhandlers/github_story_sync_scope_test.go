package taskhandlers

import (
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestGitHubStorySyncScopeUsesExplicitWorkspaceBoundSystemActor(t *testing.T) {
	t.Parallel()

	systemUserID, workspaceID := uuid.New(), uuid.New()
	handler := &handlers{systemUserID: systemUserID}
	scope, err := handler.githubStorySyncScope(workspaceID)
	if err != nil {
		t.Fatalf("githubStorySyncScope() error = %v", err)
	}
	if scope.WorkspaceID != workspaceID || scope.Actor.WorkspaceID != workspaceID {
		t.Fatalf("scope workspace = %#v", scope)
	}
	if scope.Actor.Kind != platformauth.PrincipalSystem || scope.Actor.PrincipalID != systemUserID ||
		!scope.Actor.Scopes.Has(platformauth.ScopeStoriesWrite) || scope.Actor.CredentialID != uuid.Nil {
		t.Fatalf("scope actor = %#v", scope.Actor)
	}
	if scope.ActivityUser == nil || *scope.ActivityUser != systemUserID {
		t.Fatalf("scope activity user = %v", scope.ActivityUser)
	}
}

func TestGitHubStorySyncScopeRejectsMissingIdentity(t *testing.T) {
	t.Parallel()

	if _, err := (&handlers{}).githubStorySyncScope(uuid.New()); err == nil {
		t.Fatal("githubStorySyncScope() accepted a missing system actor")
	}
	if _, err := (&handlers{systemUserID: uuid.New()}).githubStorySyncScope(uuid.Nil); err == nil {
		t.Fatal("githubStorySyncScope() accepted a missing workspace")
	}
}

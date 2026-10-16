package stories

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type attachmentReadRepository struct {
	scope   StoryReadScope
	storyID uuid.UUID
}

func (repository *attachmentReadRepository) AuthorizeVisibleStoryAttachmentRead(_ context.Context, scope StoryReadScope, storyID uuid.UUID) error {
	repository.scope, repository.storyID = scope, storyID
	return nil
}

func TestAuthorizeAttachmentReadPreservesCredentialTeamRestrictions(t *testing.T) {
	actorID, workspaceID, storyID, teamID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	access, err := auth.RestrictedTeamAccess(teamID)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := auth.NewActor(actorID, auth.PrincipalPersonalToken, uuid.New(), auth.MustScopeSet(auth.ScopeStoriesRead), access)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := auth.SetActor(t.Context(), actor)
	if err != nil {
		t.Fatal(err)
	}
	repository := &attachmentReadRepository{}
	if err := New(nil, repository, nil, nil).AuthorizeAttachmentRead(ctx, storyID, workspaceID); err != nil {
		t.Fatal(err)
	}
	if repository.storyID != storyID || repository.scope.ActorID != actorID || repository.scope.WorkspaceID != workspaceID || repository.scope.UnrestrictedTeamAccess {
		t.Fatalf("incorrect attachment read scope: %#v", repository)
	}
	if len(repository.scope.AllowedTeamIDs) != 1 || repository.scope.AllowedTeamIDs[0] != teamID {
		t.Fatalf("credential team restriction lost: %v", repository.scope.AllowedTeamIDs)
	}
}

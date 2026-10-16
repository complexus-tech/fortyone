//go:build integration

package storiesrepository

import (
	"errors"
	"testing"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestStoryAttachmentAccessObservesLiveMembershipAndDeletion(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	fixture := seedStoryReadFixture(t, ctx, postgres.Pool)
	repository := New(nil, postgres.Pool)
	scope := unrestrictedStoryReadScope(fixture.actor, fixture.workspaceA)
	if err := repository.AuthorizeVisibleStoryAttachmentRead(ctx, scope, fixture.visible); err != nil {
		t.Fatalf("visible story access: %v", err)
	}
	assertDenied := func(scope storydomain.ReadScope, storyID uuid.UUID) {
		t.Helper()
		if err := repository.AuthorizeVisibleStoryAttachmentRead(ctx, scope, storyID); !errors.Is(err, storydomain.ErrNotFound) {
			t.Fatalf("attachment story access error = %v, want not found", err)
		}
	}
	assertDenied(scope, fixture.hidden)
	assertDenied(scope, fixture.crossTenant)
	assertDenied(unrestrictedStoryReadScope(fixture.inactive, fixture.workspaceA), fixture.visible)
	restricted := scope
	restricted.UnrestrictedTeamAccess = false
	restricted.AllowedTeamIDs = []uuid.UUID{fixture.teamHidden}
	assertDenied(restricted, fixture.visible)

	mustStoryReadExec(t, ctx, postgres.Pool, "DELETE FROM team_members WHERE team_id = $1 AND user_id = $2", fixture.teamA, fixture.actor)
	assertDenied(scope, fixture.visible)
	mustStoryReadExec(t, ctx, postgres.Pool, "INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)", fixture.teamA, fixture.actor)
	if err := repository.AuthorizeVisibleStoryAttachmentRead(ctx, scope, fixture.visible); err != nil {
		t.Fatalf("restored team membership access: %v", err)
	}
	mustStoryReadExec(t, ctx, postgres.Pool, "UPDATE stories SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", fixture.visible)
	assertDenied(scope, fixture.visible)
	mustStoryReadExec(t, ctx, postgres.Pool, "UPDATE stories SET deleted_at = NULL WHERE id = $1", fixture.visible)
	if err := repository.AuthorizeVisibleStoryAttachmentRead(ctx, scope, fixture.visible); err != nil {
		t.Fatalf("restored story access: %v", err)
	}
	mustStoryReadExec(t, ctx, postgres.Pool, "DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2", fixture.workspaceA, fixture.actor)
	assertDenied(scope, fixture.visible)
}

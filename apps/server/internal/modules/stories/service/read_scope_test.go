package stories

import (
	"context"
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestReadScopeFromContextPropagatesCredentialRestrictions(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	teamAccess, err := auth.RestrictedTeamAccess(teamID)
	if err != nil {
		t.Fatalf("restricted team access: %v", err)
	}
	actor, err := auth.NewActor(
		actorID,
		auth.PrincipalPersonalToken,
		uuid.New(),
		auth.MustScopeSet(auth.ScopeStoriesRead),
		teamAccess,
	)
	if err != nil {
		t.Fatalf("new actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind workspace: %v", err)
	}
	ctx, err := auth.SetActor(context.Background(), actor)
	if err != nil {
		t.Fatalf("set actor: %v", err)
	}

	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		t.Fatalf("read scope: %v", err)
	}
	if scope.ActorID != actorID || scope.WorkspaceID != workspaceID || scope.UnrestrictedTeamAccess {
		t.Fatalf("scope = %#v", scope)
	}
	if len(scope.AllowedTeamIDs) != 1 || scope.AllowedTeamIDs[0] != teamID {
		t.Fatalf("allowed teams = %v, want %s", scope.AllowedTeamIDs, teamID)
	}
}

func TestReadScopeFromContextRejectsMissingScopeAndWorkspaceMismatch(t *testing.T) {
	workspaceID := uuid.New()
	otherWorkspaceID := uuid.New()

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing actor", ctx: context.Background()},
		{name: "missing stories scope", ctx: actorContextForReadScopeTest(t, workspaceID, auth.MustScopeSet(auth.ScopeTeamsRead))},
		{name: "different workspace", ctx: actorContextForReadScopeTest(t, otherWorkspaceID, auth.MustScopeSet(auth.ScopeStoriesRead))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readScopeFromContext(tt.ctx, workspaceID)
			if !errors.Is(err, ErrStoryReadForbidden) {
				t.Fatalf("error = %v, want ErrStoryReadForbidden", err)
			}
		})
	}
}

func actorContextForReadScopeTest(t *testing.T, workspaceID uuid.UUID, scopes auth.ScopeSet) context.Context {
	t.Helper()
	actor, err := auth.NewActor(
		uuid.New(),
		auth.PrincipalPersonalToken,
		uuid.New(),
		scopes,
		auth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("new actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind workspace: %v", err)
	}
	ctx, err := auth.SetActor(context.Background(), actor)
	if err != nil {
		t.Fatalf("set actor: %v", err)
	}
	return ctx
}

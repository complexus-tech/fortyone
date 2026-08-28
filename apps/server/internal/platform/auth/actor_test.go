package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestHumanActorContextAndWorkspaceBinding(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	workspaceID := uuid.New()
	ctx := SetUserID(context.Background(), userID)

	actor, err := GetActor(ctx)
	if err != nil {
		t.Fatalf("get actor: %v", err)
	}
	if actor.Kind != PrincipalHumanUser || actor.PrincipalID != userID || !actor.Scopes.Has(ScopeStoriesWrite) {
		t.Fatalf("unexpected first-party actor: %+v", actor)
	}

	ctx, err = BindWorkspace(ctx, workspaceID)
	if err != nil {
		t.Fatalf("bind workspace: %v", err)
	}
	bound, err := GetActor(ctx)
	if err != nil {
		t.Fatalf("get bound actor: %v", err)
	}
	if bound.WorkspaceID != workspaceID {
		t.Fatalf("workspace id = %s, want %s", bound.WorkspaceID, workspaceID)
	}
	if actor.WorkspaceID != uuid.Nil {
		t.Fatal("binding a workspace mutated the previously returned actor")
	}
}

func TestActorScopeAndTeamAccessAreDefensivelyCopied(t *testing.T) {
	t.Parallel()

	scopes := MustScopeSet(ScopeStoriesRead, ScopeCommentsRead)
	teamID := uuid.New()
	teamAccess, err := RestrictedTeamAccess(teamID)
	if err != nil {
		t.Fatalf("restricted team access: %v", err)
	}
	actor, err := NewActor(uuid.New(), PrincipalServiceAccount, uuid.New(), scopes, teamAccess)
	if err != nil {
		t.Fatalf("new actor: %v", err)
	}

	values := actor.Scopes.Values()
	values[0] = ScopeWebhooksManage
	teamIDs := actor.TeamAccess.RestrictedTeamIDs()
	teamIDs[0] = uuid.New()

	if !actor.Scopes.Has(ScopeStoriesRead) || actor.Scopes.Has(ScopeWebhooksManage) {
		t.Fatal("scope accessor allowed actor mutation")
	}
	if !actor.TeamAccess.Allows(teamID) {
		t.Fatal("team accessor allowed actor mutation")
	}
}

func TestGetUserIDRejectsNonUserPrincipal(t *testing.T) {
	t.Parallel()

	actor, err := NewActor(
		uuid.New(),
		PrincipalServiceAccount,
		uuid.New(),
		MustScopeSet(ScopeStoriesRead),
		UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("new service actor: %v", err)
	}
	ctx, err := SetActor(context.Background(), actor)
	if err != nil {
		t.Fatalf("set actor: %v", err)
	}
	if _, err := GetUserID(ctx); !errors.Is(err, ErrInvalidActor) {
		t.Fatalf("GetUserID error = %v, want ErrInvalidActor", err)
	}
}

func TestActorConstructorsFailClosed(t *testing.T) {
	t.Parallel()

	if _, err := NewScopeSet(Scope("stories:delete")); !errors.Is(err, ErrInvalidScope) {
		t.Fatalf("NewScopeSet error = %v, want ErrInvalidScope", err)
	}
	if _, err := RestrictedTeamAccess(uuid.Nil); !errors.Is(err, ErrInvalidActor) {
		t.Fatalf("RestrictedTeamAccess error = %v, want ErrInvalidActor", err)
	}
	if _, err := NewActor(uuid.Nil, PrincipalHumanUser, uuid.Nil, MustScopeSet(ScopeFirstParty), UnrestrictedTeamAccess()); !errors.Is(err, ErrInvalidActor) {
		t.Fatalf("NewActor error = %v, want ErrInvalidActor", err)
	}
}

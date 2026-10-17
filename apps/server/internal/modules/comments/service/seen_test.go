package comments

import (
	"context"
	"errors"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type seenStoreStub struct {
	serviceRepositoryFake
	seen func(context.Context, StoryScope, uuid.UUID) error
}

func (stub seenStoreStub) MarkSeen(ctx context.Context, scope StoryScope, id uuid.UUID) error {
	return stub.seen(ctx, scope, id)
}

func TestMarkSeenRejectsInvalidScopeBeforePersistence(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	valid := StoryScope{WorkspaceID: workspaceID, StoryID: uuid.New(), Actor: commentTestActor(t, uuid.New(), workspaceID)}
	commentID := uuid.New()
	for _, test := range []struct {
		name   string
		change func(*StoryScope, *uuid.UUID)
		want   error
	}{
		{name: "missing workspace", change: func(scope *StoryScope, _ *uuid.UUID) { scope.WorkspaceID = uuid.Nil }, want: ErrInvalidComment},
		{name: "missing story", change: func(scope *StoryScope, _ *uuid.UUID) { scope.StoryID = uuid.Nil }, want: ErrInvalidComment},
		{name: "missing comment", change: func(_ *StoryScope, id *uuid.UUID) { *id = uuid.Nil }, want: ErrInvalidComment},
		{name: "different actor workspace", change: func(scope *StoryScope, _ *uuid.UUID) { scope.Actor.WorkspaceID = uuid.New() }, want: ErrForbidden},
		{name: "missing actor", change: func(scope *StoryScope, _ *uuid.UUID) { scope.Actor = platformauth.Actor{} }, want: ErrForbidden},
		{name: "system actor", change: func(scope *StoryScope, _ *uuid.UUID) { scope.Actor.Kind = platformauth.PrincipalSystem }, want: ErrForbidden},
		{name: "application actor", change: func(scope *StoryScope, _ *uuid.UUID) { scope.Actor.Kind = platformauth.PrincipalOAuthApplication }, want: ErrForbidden},
		{name: "write scope without read", change: func(scope *StoryScope, _ *uuid.UUID) {
			scope.Actor.Scopes = platformauth.MustScopeSet(platformauth.ScopeCommentsWrite)
		}, want: ErrForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			scope, id := valid, commentID
			test.change(&scope, &id)
			service := New(seenStoreStub{seen: func(context.Context, StoryScope, uuid.UUID) error {
				t.Fatal("invalid scope reached persistence")
				return nil
			}})
			require.ErrorIs(t, service.MarkSeen(t.Context(), scope, id), test.want)
		})
	}
}

func TestMarkSeenAllowsReadOnlyUserAndPreservesExactScope(t *testing.T) {
	t.Parallel()
	workspaceID, actorID, storyID, commentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	actor, err := platformauth.NewActor(actorID, platformauth.PrincipalOAuthUser, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeCommentsRead), platformauth.UnrestrictedTeamAccess())
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	scope := StoryScope{WorkspaceID: workspaceID, StoryID: storyID, Actor: actor}
	calls := 0
	service := New(seenStoreStub{seen: func(ctx context.Context, actual StoryScope, id uuid.UUID) error {
		calls++
		require.Equal(t, t.Context(), ctx)
		require.Equal(t, scope, actual)
		require.Equal(t, commentID, id)
		return nil
	}})
	for range 2 {
		require.NoError(t, service.MarkSeen(t.Context(), scope, commentID))
	}
	require.Equal(t, 2, calls)
}

func TestMarkSeenPreservesRepositoryFailures(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	scope := StoryScope{WorkspaceID: workspaceID, StoryID: uuid.New(), Actor: commentTestActor(t, uuid.New(), workspaceID)}
	for _, failure := range []error{ErrNotFound, ErrForbidden, errors.New("receipt persistence unavailable")} {
		t.Run(failure.Error(), func(t *testing.T) {
			t.Parallel()
			service := New(seenStoreStub{seen: func(context.Context, StoryScope, uuid.UUID) error { return failure }})
			require.ErrorIs(t, service.MarkSeen(t.Context(), scope, uuid.New()), failure)
		})
	}
}

package feedback

import (
	"context"
	"errors"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAccessScopeFromContextPreservesCredentialTeamRestriction(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	teamID := uuid.New()
	teamAccess, err := platformauth.RestrictedTeamAccess(teamID)
	require.NoError(t, err)
	actor, err := platformauth.NewActor(
		actorID,
		platformauth.PrincipalHumanUser,
		uuid.Nil,
		platformauth.MustScopeSet(platformauth.ScopeFirstParty),
		teamAccess,
	)
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	ctx, err := platformauth.SetActor(context.Background(), actor)
	require.NoError(t, err)

	scope, err := accessScopeFromContext(ctx, workspaceID, actorID)
	require.NoError(t, err)
	require.Equal(t, actorID, scope.ActorID)
	require.False(t, scope.AllTeams)
	require.Equal(t, []uuid.UUID{teamID}, scope.CredentialTeamIDs)
}

func TestAccessScopeFromContextRejectsMachineAndDelegatedActors(t *testing.T) {
	workspaceID := uuid.New()
	for _, kind := range []platformauth.PrincipalKind{
		platformauth.PrincipalPersonalToken,
		platformauth.PrincipalOAuthUser,
		platformauth.PrincipalOAuthApplication,
		platformauth.PrincipalServiceAccount,
		platformauth.PrincipalSystem,
	} {
		t.Run(string(kind), func(t *testing.T) {
			actor, err := platformauth.NewActor(
				uuid.New(),
				kind,
				uuid.New(),
				platformauth.MustScopeSet(platformauth.ScopeFirstParty),
				platformauth.UnrestrictedTeamAccess(),
			)
			require.NoError(t, err)
			actor, err = actor.WithWorkspace(workspaceID)
			require.NoError(t, err)
			ctx, err := platformauth.SetActor(context.Background(), actor)
			require.NoError(t, err)

			_, err = accessScopeFromContext(ctx, workspaceID, uuid.Nil)
			require.ErrorIs(t, err, ErrForbidden)
		})
	}
}

func TestAccessScopeFromContextSupportsExplicitLegacyHumanFallback(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	scope, err := accessScopeFromContext(context.Background(), workspaceID, actorID)
	require.NoError(t, err)
	require.Equal(t, actorID, scope.ActorID)
	require.True(t, scope.AllTeams)

	_, err = accessScopeFromContext(context.Background(), workspaceID, uuid.Nil)
	require.True(t, errors.Is(err, ErrForbidden))
}

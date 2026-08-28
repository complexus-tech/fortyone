package developeraccess

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type machineResolverStub struct {
	actor platformauth.Actor
	err   error
	calls int
}

func (stub *machineResolverStub) ResolveMachineCredential(context.Context, string) (platformauth.Actor, error) {
	stub.calls++
	return stub.actor, stub.err
}

type oauthVerifierStub struct {
	identity developeroauthdomain.AccessIdentity
	err      error
	calls    int
}

func (stub *oauthVerifierStub) VerifyAccessToken(context.Context, string) (developeroauthdomain.AccessIdentity, error) {
	stub.calls++
	return stub.identity, stub.err
}

func TestResolverPreservesMachineCredentialBoundary(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	want, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalServiceAccount, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeStoriesWrite),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	want, err = want.WithWorkspace(workspaceID)
	require.NoError(t, err)
	machine := &machineResolverStub{actor: want}
	oauth := &oauthVerifierStub{err: errors.New("must not be called")}
	resolver, err := NewResolver(machine, oauth)
	require.NoError(t, err)

	got, err := resolver.ResolveDeveloperCredential(context.Background(), "f41_sak_v1_example")
	require.NoError(t, err)
	require.Equal(t, want, got)
	require.Equal(t, 1, machine.calls)
	require.Zero(t, oauth.calls)
}

func TestResolverCreatesAnUnboundOAuthUserActorWithPublishedScopes(t *testing.T) {
	t.Parallel()

	grantID := uuid.New()
	identity := developeroauthdomain.AccessIdentity{
		UserID: uuid.New(), ApplicationID: uuid.New(), GrantID: grantID,
		ActorKind: platformauth.PrincipalOAuthUser, Resource: "https://api.fortyone.app/api/v1",
		Scopes:    []string{developeroauthdomain.ScopeOfflineAccess, string(platformauth.ScopeStoriesRead)},
		ExpiresAt: time.Now().Add(time.Minute), OAuthCredential: uuid.New(),
	}
	identity.PrincipalID = identity.UserID
	resolver, err := NewResolver(
		&machineResolverStub{err: errors.New("not a machine key")},
		&oauthVerifierStub{identity: identity},
	)
	require.NoError(t, err)

	actor, err := resolver.ResolveDeveloperCredential(context.Background(), "header.payload.signature")
	require.NoError(t, err)
	require.Equal(t, platformauth.PrincipalOAuthUser, actor.Kind)
	require.Equal(t, identity.UserID, actor.PrincipalID)
	require.Equal(t, grantID, actor.CredentialID)
	require.Equal(t, uuid.Nil, actor.WorkspaceID)
	require.True(t, actor.Scopes.Has(platformauth.ScopeStoriesRead))
	require.False(t, actor.Scopes.Has(platformauth.ScopeStoriesWrite))
}

func TestResolverCreatesABoundOAuthApplicationActorFromInstallationIdentity(t *testing.T) {
	t.Parallel()

	principalID, installationID, workspaceID := uuid.New(), uuid.New(), uuid.New()
	identity := developeroauthdomain.AccessIdentity{
		PrincipalID: principalID, ApplicationID: uuid.New(), InstallationID: installationID,
		WorkspaceID: workspaceID, ActorKind: platformauth.PrincipalOAuthApplication,
		Resource:  "https://api.fortyone.app/api/v1",
		Scopes:    []string{string(platformauth.ScopeStoriesWrite)},
		ExpiresAt: time.Now().Add(time.Minute), OAuthCredential: uuid.New(),
	}
	resolver, err := NewResolver(
		&machineResolverStub{err: errors.New("not a machine key")},
		&oauthVerifierStub{identity: identity},
	)
	require.NoError(t, err)

	actor, err := resolver.ResolveDeveloperCredential(context.Background(), "header.payload.signature")
	require.NoError(t, err)
	require.Equal(t, platformauth.PrincipalOAuthApplication, actor.Kind)
	require.Equal(t, principalID, actor.PrincipalID)
	require.Equal(t, installationID, actor.CredentialID)
	require.Equal(t, workspaceID, actor.WorkspaceID)
	require.True(t, actor.Scopes.Has(platformauth.ScopeStoriesWrite))
	_, err = actor.UserID()
	require.Error(t, err, "an installation principal must never become its installer")
}

func TestResolverFailsClosedForInvalidFamiliesAndScopes(t *testing.T) {
	t.Parallel()

	validIdentity := developeroauthdomain.AccessIdentity{
		UserID: uuid.New(), ApplicationID: uuid.New(), GrantID: uuid.New(),
		ActorKind: platformauth.PrincipalOAuthUser, Resource: "https://api.fortyone.app/api/v1",
		Scopes: []string{developeroauthdomain.ScopeOfflineAccess, string(platformauth.ScopeStoriesRead)},
	}
	validIdentity.PrincipalID = validIdentity.UserID
	tests := []struct {
		name     string
		raw      string
		identity developeroauthdomain.AccessIdentity
		oauthErr error
	}{
		{name: "empty bearer"},
		{name: "surrounding whitespace", raw: " token ", identity: validIdentity},
		{name: "oversized bearer", raw: strings.Repeat("x", maximumBearerBytes+1), identity: validIdentity},
		{name: "OAuth verification failure", raw: "jwt", oauthErr: errors.New("invalid")},
		{name: "offline only", raw: "jwt", identity: func() developeroauthdomain.AccessIdentity {
			copy := validIdentity
			copy.Scopes = []string{developeroauthdomain.ScopeOfflineAccess}
			return copy
		}()},
		{name: "unknown scope", raw: "jwt", identity: func() developeroauthdomain.AccessIdentity {
			copy := validIdentity
			copy.Scopes = []string{"admin:*"}
			return copy
		}()},
		{name: "application actor shaped like user grant", raw: "jwt", identity: func() developeroauthdomain.AccessIdentity {
			copy := validIdentity
			copy.ActorKind = platformauth.PrincipalOAuthApplication
			return copy
		}()},
		{name: "application actor with offline access", raw: "jwt", identity: developeroauthdomain.AccessIdentity{
			PrincipalID: uuid.New(), ApplicationID: uuid.New(), InstallationID: uuid.New(), WorkspaceID: uuid.New(),
			ActorKind: platformauth.PrincipalOAuthApplication,
			Resource:  "https://api.fortyone.app/api/v1",
			Scopes:    []string{developeroauthdomain.ScopeOfflineAccess, string(platformauth.ScopeStoriesWrite)},
		}},
		{name: "application actor with installer user", raw: "jwt", identity: developeroauthdomain.AccessIdentity{
			PrincipalID: uuid.New(), UserID: uuid.New(), ApplicationID: uuid.New(), InstallationID: uuid.New(), WorkspaceID: uuid.New(),
			ActorKind: platformauth.PrincipalOAuthApplication,
			Resource:  "https://api.fortyone.app/api/v1",
			Scopes:    []string{string(platformauth.ScopeStoriesWrite)},
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			resolver, err := NewResolver(
				&machineResolverStub{err: errors.New("not a machine key")},
				&oauthVerifierStub{identity: test.identity, err: test.oauthErr},
			)
			require.NoError(t, err)
			_, err = resolver.ResolveDeveloperCredential(context.Background(), test.raw)
			require.ErrorIs(t, err, ErrAuthenticationFailed)
		})
	}
}

func TestNewResolverRequiresBothRevocationAwareVerifiers(t *testing.T) {
	t.Parallel()

	_, err := NewResolver(nil, &oauthVerifierStub{})
	require.Error(t, err)
	_, err = NewResolver(&machineResolverStub{}, nil)
	require.Error(t, err)
}

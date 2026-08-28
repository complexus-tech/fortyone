package outboundwebhooksservice

import (
	"testing"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOAuthUserWebhookManagementRequiresCurrentAdminRoleAndScope(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	allowed := newOAuthWebhookActor(t, workspaceID, platformauth.ScopeWebhooksManage)
	missingScope := newOAuthWebhookActor(t, workspaceID, platformauth.ScopeStoriesRead)
	wrongWorkspace := newOAuthWebhookActor(t, uuid.New(), platformauth.ScopeWebhooksManage)
	unbound, err := platformauth.NewActor(
		uuid.New(),
		platformauth.PrincipalOAuthUser,
		uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeWebhooksManage),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	oauthApplication, err := platformauth.NewActor(
		uuid.New(),
		platformauth.PrincipalOAuthApplication,
		uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeWebhooksManage),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	oauthApplication, err = oauthApplication.WithWorkspace(workspaceID)
	require.NoError(t, err)

	for _, test := range []struct {
		name    string
		access  Access
		wantErr error
	}{
		{
			name: "current admin with management scope",
			access: Access{
				Actor: allowed, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
		},
		{
			name: "current member",
			access: Access{
				Actor: allowed, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleMember,
			},
			wantErr: authorization.ErrInsufficientWorkspaceRole,
		},
		{
			name: "current guest",
			access: Access{
				Actor: allowed, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleGuest,
			},
			wantErr: authorization.ErrInsufficientWorkspaceRole,
		},
		{
			name: "management scope was revoked",
			access: Access{
				Actor: missingScope, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
			wantErr: authorization.ErrCredentialScopeDenied,
		},
		{
			name: "workspace does not match current membership",
			access: Access{
				Actor: wrongWorkspace, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
			wantErr: authorization.ErrWorkspaceMismatch,
		},
		{
			name: "workspace was not membership-bound",
			access: Access{
				Actor: unbound, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
			wantErr: authorization.ErrWorkspaceMismatch,
		},
		{
			name: "application actors are not released",
			access: Access{
				Actor: oauthApplication, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
			wantErr: authorization.ErrPrincipalKindDenied,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := authorizeManagement(test.access)
			if test.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, test.wantErr)
			require.ErrorIs(t, err, outboundwebhooksdomain.ErrEndpointOwnerInactive)
		})
	}
}

func TestOAuthUserWebhookManagerChecksAuthorizationBeforeSideEffects(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	for _, test := range []struct {
		name    string
		access  Access
		wantErr error
	}{
		{
			name: "member role",
			access: Access{
				Actor:         newOAuthWebhookActor(t, workspaceID, platformauth.ScopeWebhooksManage),
				WorkspaceID:   workspaceID,
				WorkspaceRole: authorization.WorkspaceRoleMember,
			},
			wantErr: authorization.ErrInsufficientWorkspaceRole,
		},
		{
			name: "missing scope",
			access: Access{
				Actor:         newOAuthWebhookActor(t, workspaceID, platformauth.ScopeStoriesRead),
				WorkspaceID:   workspaceID,
				WorkspaceRole: authorization.WorkspaceRoleAdmin,
			},
			wantErr: authorization.ErrCredentialScopeDenied,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repository := &endpointRepositoryStub{}
			principals := &principalResolverStub{id: uuid.New()}
			validator := &endpointValidatorStub{canonical: "https://hooks.example.com/receive"}
			manager, err := newManager(
				repository,
				principals,
				newTestSecretManager(t),
				validator,
				&testClock{values: []time.Time{time.Unix(1_700_000_000, 0)}},
				&testIDs{values: []uuid.UUID{uuid.New(), uuid.New()}},
			)
			require.NoError(t, err)

			_, err = manager.CreateEndpoint(t.Context(), test.access, validOAuthWebhookInput())

			require.ErrorIs(t, err, test.wantErr)
			require.Zero(t, validator.calls, "endpoint validation must not run before authorization")
			require.Zero(t, principals.calls, "owner resolution must not run before authorization")
			require.Zero(t, repository.createCalls, "persistence must not run before authorization")
		})
	}
}

func TestOAuthUserWebhookManagerPreservesUserAttribution(t *testing.T) {
	t.Parallel()

	workspaceID, userID, grantID := uuid.New(), uuid.New(), uuid.New()
	actor := newOAuthWebhookActorWithIdentity(
		t,
		workspaceID,
		userID,
		grantID,
		platformauth.ScopeWebhooksManage,
	)
	repository := &endpointRepositoryStub{}
	principals := &principalResolverStub{id: userID}
	validator := &endpointValidatorStub{canonical: "https://hooks.example.com/receive"}
	endpointID, auditID := uuid.New(), uuid.New()
	manager, err := newManager(
		repository,
		principals,
		newTestSecretManager(t),
		validator,
		&testClock{values: []time.Time{time.Unix(1_700_000_000, 0)}},
		&testIDs{values: []uuid.UUID{endpointID, auditID}},
	)
	require.NoError(t, err)

	created, err := manager.CreateEndpoint(t.Context(), Access{
		Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}, validOAuthWebhookInput())

	require.NoError(t, err)
	require.Equal(t, endpointID, created.Endpoint.ID)
	require.Equal(t, userID, created.Endpoint.OwnerPrincipalID)
	require.Equal(t, 1, principals.calls)
	require.Equal(t, 1, repository.createCalls)
	require.Equal(t, actor, repository.created.Actor)
	require.Equal(t, platformauth.PrincipalOAuthUser, repository.created.Actor.Kind)
	require.Equal(t, userID, repository.created.Actor.PrincipalID)
	require.Equal(t, grantID, repository.created.Actor.CredentialID)
	require.Equal(t, workspaceID, repository.created.WorkspaceID)
	require.Equal(t, authorization.WorkspaceRoleAdmin, repository.created.WorkspaceRole)
}

func validOAuthWebhookInput() CreateEndpointInput {
	return CreateEndpointInput{
		Name:          "Production events",
		URL:           "https://hooks.example.com/receive",
		Subscriptions: []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated},
		RequestID:     "request-oauth-webhook",
	}
}

func newOAuthWebhookActor(t *testing.T, workspaceID uuid.UUID, scopes ...platformauth.Scope) platformauth.Actor {
	t.Helper()
	return newOAuthWebhookActorWithIdentity(t, workspaceID, uuid.New(), uuid.New(), scopes...)
}

func newOAuthWebhookActorWithIdentity(
	t *testing.T,
	workspaceID,
	userID,
	grantID uuid.UUID,
	scopes ...platformauth.Scope,
) platformauth.Actor {
	t.Helper()
	actor, err := platformauth.NewActor(
		userID,
		platformauth.PrincipalOAuthUser,
		grantID,
		platformauth.MustScopeSet(scopes...),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	return actor
}

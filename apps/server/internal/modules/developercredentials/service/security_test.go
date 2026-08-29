package developercredentials

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountManagementRejectsNonHumanPrincipals(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	machineActor, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalServiceAccount, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeServiceAccountsManage),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	machineActor, err = machineActor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	service := newTestService(t, &fakeRepository{}, time.Now().UTC())

	_, err = service.CreateServiceAccount(context.Background(), developercredentialsdomain.Access{
		Actor: machineActor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}, CreateServiceAccountInput{Name: "build agent", WorkspaceRole: authorization.WorkspaceRoleMember})

	require.ErrorIs(t, err, developercredentialsdomain.ErrAccessDenied)
	require.ErrorIs(t, err, authorization.ErrPrincipalKindDenied)
}

func TestCreateServiceAccountRejectsPrivilegedRole(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &fakeRepository{}, time.Now().UTC())
	access := humanAdminAccess(t)

	_, err := service.CreateServiceAccount(context.Background(), access, CreateServiceAccountInput{
		Name: "over-privileged", WorkspaceRole: authorization.WorkspaceRoleAdmin,
	})

	require.ErrorIs(t, err, developercredentialsdomain.ErrInvalidServiceAccountRole)
}

func TestCreateServiceAccountAcceptsOnlyGuestAndMemberRoles(t *testing.T) {
	t.Parallel()

	for _, role := range []authorization.WorkspaceRole{authorization.WorkspaceRoleGuest, authorization.WorkspaceRoleMember} {
		role := role
		t.Run(string(role), func(t *testing.T) {
			t.Parallel()
			var observed developercredentialsdomain.CreateServiceAccount
			repository := &fakeRepository{createServiceAccount: func(command developercredentialsdomain.CreateServiceAccount) (developercredentialsdomain.ServiceAccount, error) {
				observed = command
				return developercredentialsdomain.ServiceAccount{ID: command.PrincipalID, WorkspaceRole: command.WorkspaceRole}, nil
			}}
			service := newTestService(t, repository, time.Now().UTC())

			account, err := service.CreateServiceAccount(context.Background(), humanAdminAccess(t), CreateServiceAccountInput{
				Name: "build agent", WorkspaceRole: role,
			})

			require.NoError(t, err)
			require.Equal(t, role, observed.WorkspaceRole)
			require.Equal(t, role, account.WorkspaceRole)
		})
	}
}

func TestServiceAccountKeyRejectsCredentialManagementScope(t *testing.T) {
	t.Parallel()

	repository := &fakeRepository{}
	now := time.Now().UTC()
	service := newTestService(t, repository, now)

	_, err := service.CreateServiceAccountKey(context.Background(), humanAdminAccess(t), uuid.New(), CreateServiceAccountKeyInput{
		Name: "recursive admin", Scopes: []platformauth.Scope{platformauth.ScopeServiceAccountsManage},
		ExpiresAt: now.Add(time.Hour),
	})

	require.ErrorIs(t, err, developercredentialsdomain.ErrInvalidScope)
}

func TestPersonalTokenActorPreservesUnderlyingUserAndCredentialIdentity(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	workspaceID := uuid.New()
	userID := uuid.New()
	principalRecordID := uuid.New()
	credentialID := uuid.New()
	tokens := testTokenManager(t)
	issued, err := tokens.issue(developercredentialsdomain.CredentialPersonalAccessToken, credentialID)
	require.NoError(t, err)
	touched := false
	repository := &fakeRepository{
		lookupCredential: func(prefix string, kind developercredentialsdomain.CredentialKind, version int16, authenticatedAt time.Time) (developercredentialsdomain.VerificationRecord, error) {
			require.Equal(t, issued.Material.LookupPrefix, prefix)
			require.Equal(t, developercredentialsdomain.CredentialPersonalAccessToken, kind)
			require.Equal(t, issued.Material.TokenVersion, version)
			require.Equal(t, now, authenticatedAt)
			return developercredentialsdomain.VerificationRecord{
				CredentialID: credentialID, WorkspaceID: workspaceID, PrincipalRecord: principalRecordID,
				PrincipalKind: "human_user", SubjectUserID: &userID,
				WorkspaceRole:  authorization.WorkspaceRoleAdmin,
				CredentialKind: developercredentialsdomain.CredentialPersonalAccessToken,
				LookupPrefix:   issued.Material.LookupPrefix,
				SecretDigest:   append([]byte(nil), issued.Material.SecretDigest...),
				TokenVersion:   issued.Material.TokenVersion, DigestKey: issued.Material.DigestKey,
				Scopes: []platformauth.Scope{platformauth.ScopeStoriesRead}, ExpiresAt: now.Add(time.Hour),
			}, nil
		},
		confirmCredential: func(gotID uuid.UUID, usedAt time.Time, touchBefore time.Time) error {
			touched = true
			require.Equal(t, credentialID, gotID)
			require.Equal(t, now, usedAt)
			require.Equal(t, now.Add(-lastUsedWriteInterval), touchBefore)
			return nil
		},
	}
	service, err := New(repository, tokens, fixedClock{now: now}, &sequenceIDGenerator{})
	require.NoError(t, err)

	actor, err := service.ResolveMachineCredential(context.Background(), issued.Plaintext.Reveal())

	require.NoError(t, err)
	require.True(t, touched)
	require.Equal(t, platformauth.PrincipalPersonalToken, actor.Kind)
	require.Equal(t, userID, actor.PrincipalID)
	require.Equal(t, credentialID, actor.CredentialID)
	require.Equal(t, workspaceID, actor.WorkspaceID)
	require.True(t, actor.IsUserActor())
	gotUserID, err := actor.UserID()
	require.NoError(t, err)
	require.Equal(t, userID, gotUserID)
	require.NotEqual(t, principalRecordID, actor.PrincipalID)
}

func TestPlaintextTokenFormattingIsAlwaysRedacted(t *testing.T) {
	t.Parallel()
	token := developercredentialsdomain.NewPlaintextToken("f41_pat_v1_public_secret")
	require.Equal(t, "[REDACTED]", token.String())
	require.NotContains(t, token.LogValue().String(), "secret")
	require.NotContains(t, fmt.Sprintf("%#v", token), "secret")
	encoded, err := json.Marshal(token)
	require.NoError(t, err)
	require.JSONEq(t, `"[REDACTED]"`, string(encoded))
}

func TestEnsureHumanPrincipalLetsPATResolveButNeverProvision(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	principalRecordID := uuid.New()
	patActor, err := platformauth.NewActor(
		userID, platformauth.PrincipalPersonalToken, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeWebhooksManage),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	patActor, err = patActor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	repository := &fakeRepository{resolveHumanPrincipal: func(gotWorkspaceID uuid.UUID, gotUserID uuid.UUID) (uuid.UUID, error) {
		require.Equal(t, workspaceID, gotWorkspaceID)
		require.Equal(t, userID, gotUserID)
		return principalRecordID, nil
	}}
	service, err := New(repository, testTokenManager(t), fixedClock{now: time.Now().UTC()}, &sequenceIDGenerator{})
	require.NoError(t, err)

	resolvedID, err := service.EnsureHumanPrincipal(context.Background(), developercredentialsdomain.Access{
		Actor: patActor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}, EnsureHumanPrincipalInput{RequestID: "webhook-create"})

	require.NoError(t, err)
	require.Equal(t, principalRecordID, resolvedID)
}

func TestEnsureHumanPrincipalRetriesSerializableProvisioningConflict(t *testing.T) {
	t.Parallel()
	wantPrincipalID := uuid.New()
	attempts := 0
	repository := &fakeRepository{ensureHumanPrincipal: func(developercredentialsdomain.EnsureHumanPrincipal) (uuid.UUID, error) {
		attempts++
		if attempts < ensureHumanPrincipalAttempts {
			return uuid.Nil, developercredentialsdomain.ErrConcurrentUpdate
		}
		return wantPrincipalID, nil
	}}
	service := newTestService(t, repository, time.Now().UTC())

	principalID, err := service.EnsureHumanPrincipal(
		context.Background(),
		humanAdminAccess(t),
		EnsureHumanPrincipalInput{RequestID: "webhook-create"},
	)

	require.NoError(t, err)
	require.Equal(t, wantPrincipalID, principalID)
	require.Equal(t, ensureHumanPrincipalAttempts, attempts)
}

func TestEnsureHumanPrincipalBoundsSerializableProvisioningRetries(t *testing.T) {
	t.Parallel()
	attempts := 0
	repository := &fakeRepository{ensureHumanPrincipal: func(developercredentialsdomain.EnsureHumanPrincipal) (uuid.UUID, error) {
		attempts++
		return uuid.Nil, developercredentialsdomain.ErrConcurrentUpdate
	}}
	service := newTestService(t, repository, time.Now().UTC())

	_, err := service.EnsureHumanPrincipal(
		context.Background(),
		humanAdminAccess(t),
		EnsureHumanPrincipalInput{RequestID: "webhook-create"},
	)

	require.ErrorIs(t, err, developercredentialsdomain.ErrConcurrentUpdate)
	require.Equal(t, ensureHumanPrincipalAttempts, attempts)
}

func TestEnsureHumanPrincipalDeniesServiceAccounts(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	actor, err := platformauth.NewActor(
		uuid.New(), platformauth.PrincipalServiceAccount, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeWebhooksManage),
		platformauth.UnrestrictedTeamAccess(),
	)
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	service := newTestService(t, &fakeRepository{}, time.Now().UTC())

	_, err = service.EnsureHumanPrincipal(context.Background(), developercredentialsdomain.Access{
		Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}, EnsureHumanPrincipalInput{})

	require.ErrorIs(t, err, developercredentialsdomain.ErrAccessDenied)
	require.ErrorIs(t, err, authorization.ErrPrincipalKindDenied)
}

func newTestService(t *testing.T, repository Repository, now time.Time) *Service {
	t.Helper()
	service, err := New(repository, testTokenManager(t), fixedClock{now: now}, &sequenceIDGenerator{
		values: []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()},
	})
	require.NoError(t, err)
	return service
}

func humanAdminAccess(t *testing.T) developercredentialsdomain.Access {
	t.Helper()
	workspaceID := uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	require.NoError(t, err)
	return developercredentialsdomain.Access{
		Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}
}

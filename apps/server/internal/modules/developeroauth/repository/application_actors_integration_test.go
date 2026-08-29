//go:build integration

package developeroauthrepository

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const applicationActorResource = "https://api.fortyone.app/api/v1"

type applicationActorHarness struct {
	pool    *pgxpool.Pool
	clock   *testkit.ManualClock
	manager *developeroauth.ApplicationManager
	service *developeroauth.Service
	access  developeroauthdomain.ManagementAccess
}

func TestOAuthApplicationActorLifecycleUsesDigestsDedicatedPrincipalsAndRevocationChecks(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	harness := newApplicationActorHarness(t, ctx, postgres.Pool)

	issued := createManagedApplication(t, ctx, harness)
	require.NotEmpty(t, issued.Secret.Plaintext.Reveal())
	var digest []byte
	var prefix string
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT secret_digest, lookup_prefix
		FROM oauth_client_secrets
		WHERE secret_id = $1
	`, issued.Secret.Secret.ID).Scan(&digest, &prefix))
	require.Len(t, digest, 32)
	require.Equal(t, issued.Secret.Secret.LookupPrefix, prefix)
	require.NotContains(t, string(digest), issued.Secret.Plaintext.Reveal())

	installation, err := harness.manager.InstallApplication(ctx, harness.access, developeroauth.InstallApplicationInput{
		ClientID: issued.Application.ClientID, Resource: applicationActorResource,
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, RequestID: "install-request",
	})
	require.NoError(t, err)
	require.NotEqual(t, harness.access.Actor.PrincipalID, installation.PrincipalID)
	require.NotEqual(t, installation.ID, installation.PrincipalID)
	var principalKind string
	var subjectUserID *uuid.UUID
	var workspaceRole string
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT kind, subject_user_id, CAST(workspace_role AS text)
		FROM principals
		WHERE principal_id = $1
	`, installation.PrincipalID).Scan(&principalKind, &subjectUserID, &workspaceRole))
	require.Equal(t, string(platformauth.PrincipalOAuthApplication), principalKind)
	require.Nil(t, subjectUserID)
	require.Equal(t, "member", workspaceRole)

	token := exchangeApplicationToken(t, ctx, harness, issued.Application.ClientID,
		issued.Secret.Plaintext.Reveal(), installation.ID, "initial-exchange")
	identity, err := harness.service.VerifyAccessToken(ctx, token.AccessToken.Reveal())
	require.NoError(t, err)
	require.Equal(t, installation.ID, identity.InstallationID)
	require.Equal(t, installation.PrincipalID, identity.PrincipalID)
	require.Equal(t, uuid.Nil, identity.UserID)
	require.NotEqual(t, installation.ID, identity.OAuthCredential, "access-token jti must remain audit-only")
	_, err = harness.manager.RotateClientSecret(ctx, harness.access, issued.Application.ID, developeroauth.RotateClientSecretInput{
		ExpiresAt: issued.Application.ExpiresAt.Add(time.Minute), Overlap: time.Minute,
		RequestID: "secret-outlives-application",
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidExpiry)

	shortLived, err := harness.manager.RotateClientSecret(ctx, harness.access, issued.Application.ID, developeroauth.RotateClientSecretInput{
		ExpiresAt: harness.clock.Now().Add(2 * time.Minute), Overlap: 5 * time.Minute,
		RequestID: "short-lived-rotation",
	})
	require.NoError(t, err)
	require.NotNil(t, shortLived.PreviousSecretOverlapExpiresAt)
	require.Equal(t, harness.clock.Now().Add(5*time.Minute), *shortLived.PreviousSecretOverlapExpiresAt)
	var overlapCutoff time.Time
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT overlap_expires_at
		FROM oauth_client_secrets
		WHERE secret_id = $1
	`, issued.Secret.Secret.ID).Scan(&overlapCutoff))
	require.WithinDuration(t, harness.clock.Now().Add(5*time.Minute), overlapCutoff, 0)
	exchangeApplicationToken(t, ctx, harness, issued.Application.ClientID,
		issued.Secret.Plaintext.Reveal(), installation.ID, "overlap-exchange")
	current, err := harness.manager.RotateClientSecret(ctx, harness.access, issued.Application.ID, developeroauth.RotateClientSecretInput{
		ExpiresAt: harness.clock.Now().Add(14 * 24 * time.Hour), Overlap: 5 * time.Minute,
		RequestID: "current-rotation",
	})
	require.NoError(t, err)
	require.NotNil(t, current.PreviousSecretOverlapExpiresAt)
	require.Equal(t, harness.clock.Now().Add(5*time.Minute), *current.PreviousSecretOverlapExpiresAt)
	exchangeApplicationToken(t, ctx, harness, issued.Application.ClientID,
		shortLived.Plaintext.Reveal(), installation.ID, "short-lived-overlap")

	harness.clock.Advance(3 * time.Minute)
	_, err = harness.service.ExchangeClientCredentials(ctx, developeroauth.ClientCredentialsExchange{
		ClientID: issued.Application.ClientID, ClientSecret: shortLived.Plaintext.Reveal(),
		InstallationID: installation.ID, Resource: applicationActorResource,
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, RequestID: "expired-before-overlap-cutoff",
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidClient)

	harness.clock.Advance(3 * time.Minute)
	_, err = harness.service.ExchangeClientCredentials(ctx, developeroauth.ClientCredentialsExchange{
		ClientID: issued.Application.ClientID, ClientSecret: issued.Secret.Plaintext.Reveal(),
		InstallationID: installation.ID, Resource: applicationActorResource,
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, RequestID: "expired-at-overlap-cutoff",
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidClient)
	currentToken := exchangeApplicationToken(t, ctx, harness, issued.Application.ClientID,
		current.Plaintext.Reveal(), installation.ID, "rotated-exchange")
	require.NoError(t, harness.manager.RevokeClientSecret(
		ctx,
		harness.access,
		issued.Application.ID,
		current.Secret.ID,
		developeroauth.RevokeApplicationInput{Reason: "secret_compromised", RequestID: "revoke-current-secret"},
	))
	_, err = harness.service.ExchangeClientCredentials(ctx, developeroauth.ClientCredentialsExchange{
		ClientID: issued.Application.ClientID, ClientSecret: current.Plaintext.Reveal(),
		InstallationID: installation.ID, Resource: applicationActorResource,
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, RequestID: "revoked-secret-exchange",
	})
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidClient)
	_, err = harness.service.VerifyAccessToken(ctx, currentToken.AccessToken.Reveal())
	require.NoError(t, err, "revoking one client secret must not rewrite an already-issued installation token")
	recovered, err := harness.manager.RotateClientSecret(ctx, harness.access, issued.Application.ID, developeroauth.RotateClientSecretInput{
		ExpiresAt: harness.clock.Now().Add(14 * 24 * time.Hour), Overlap: time.Minute,
		RequestID: "recover-after-current-secret-revocation",
	})
	require.NoError(t, err)
	require.Nil(t, recovered.Secret.RotatedFromID, "recovery after all active heads are revoked starts a new fenced chain")
	require.Nil(t, recovered.PreviousSecretOverlapExpiresAt)
	exchangeApplicationToken(t, ctx, harness, issued.Application.ClientID,
		recovered.Plaintext.Reveal(), installation.ID, "recovered-secret-exchange")

	require.NoError(t, harness.manager.RevokeApplicationInstallation(ctx, harness.access, installation.ID, developeroauth.RevokeApplicationInput{
		Reason: "security_response", RequestID: "revoke-installation",
	}))
	_, err = harness.service.VerifyAccessToken(ctx, currentToken.AccessToken.Reveal())
	require.ErrorIs(t, err, developeroauthdomain.ErrInvalidGrant)

	var auditActorID, auditCredentialID, auditSubjectID uuid.UUID
	var auditUserID *uuid.UUID
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT actor_id, actor_credential_id, subject_id, user_id
		FROM oauth_audit_events
		WHERE operation = 'client_credentials.exchanged'
		  AND request_id = 'rotated-exchange'
	`).Scan(&auditActorID, &auditCredentialID, &auditSubjectID, &auditUserID))
	require.Equal(t, installation.PrincipalID, auditActorID)
	require.Equal(t, installation.ID, auditCredentialID)
	require.NotEqual(t, installation.ID, auditSubjectID)
	require.Nil(t, auditUserID, "application issuance must never be attributed to the installer")

	_, err = postgres.Pool.Exec(ctx, `UPDATE oauth_audit_events SET result = 'failed'`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "immutable")
}

func TestOAuthApplicationActorConcurrentRotationAndRevocationStayCoherent(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	harness := newApplicationActorHarness(t, ctx, postgres.Pool)
	issued := createManagedApplication(t, ctx, harness)
	installation, err := harness.manager.InstallApplication(ctx, harness.access, developeroauth.InstallApplicationInput{
		ClientID: issued.Application.ClientID, Resource: applicationActorResource,
		Scopes: []string{string(platformauth.ScopeStoriesWrite)}, RequestID: "concurrent-install",
	})
	require.NoError(t, err)

	start := make(chan struct{})
	rotationErrors := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, rotateErr := harness.manager.RotateClientSecret(ctx, harness.access, issued.Application.ID, developeroauth.RotateClientSecretInput{
				ExpiresAt: harness.clock.Now().Add(10 * 24 * time.Hour), Overlap: 10 * time.Minute,
				RequestID: uuid.NewString(),
			})
			rotationErrors <- rotateErr
		}()
	}
	close(start)
	wait.Wait()
	close(rotationErrors)
	for rotateErr := range rotationErrors {
		require.NoError(t, rotateErr)
	}
	var secretCount, rotationHeads int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT
			count(*),
			count(*) FILTER (WHERE overlap_expires_at IS NULL AND revoked_at IS NULL)
		FROM oauth_client_secrets
		WHERE application_id = $1
	`, issued.Application.ID).Scan(&secretCount, &rotationHeads))
	require.Equal(t, 3, secretCount)
	require.Equal(t, 1, rotationHeads, "concurrent rotations must serialize into one chain head")

	start = make(chan struct{})
	revocationErrors := make(chan error, 2)
	wait = sync.WaitGroup{}
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			revocationErrors <- harness.manager.RevokeApplicationInstallation(ctx, harness.access, installation.ID, developeroauth.RevokeApplicationInput{
				Reason: "concurrent_revoke", RequestID: uuid.NewString(),
			})
		}()
	}
	close(start)
	wait.Wait()
	close(revocationErrors)
	for revokeErr := range revocationErrors {
		require.NoError(t, revokeErr)
	}
	var installationStatus, principalStatus string
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT installation.status, principal.status
		FROM oauth_application_installations AS installation
		INNER JOIN principals AS principal ON principal.principal_id = installation.principal_id
		WHERE installation.installation_id = $1
	`, installation.ID).Scan(&installationStatus, &principalStatus))
	require.Equal(t, "revoked", installationStatus)
	require.Equal(t, "disabled", principalStatus)
}

func TestOAuthApplicationManagementRechecksCurrentWorkspaceAdmin(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	harness := newApplicationActorHarness(t, ctx, postgres.Pool)
	createManagedApplication(t, ctx, harness)
	_, err := postgres.Pool.Exec(ctx, `
		UPDATE workspace_members
		SET role = 'member'
		WHERE workspace_id = $1 AND user_id = $2
	`, harness.access.WorkspaceID, harness.access.Actor.PrincipalID)
	require.NoError(t, err)
	_, err = harness.manager.ListManagedApplications(ctx, harness.access)
	require.ErrorIs(t, err, developeroauthdomain.ErrAccessDenied)
}

func newApplicationActorHarness(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) applicationActorHarness {
	t.Helper()
	now := time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC)
	clock := testkit.NewManualClock(now)
	adminID := insertOAuthUser(t, ctx, pool, true)
	workspaceID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'OAuth application workspace', $2, $3)
	`, workspaceID, "oauth-app-"+uuid.NewString(), adminID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, workspaceID, adminID)
	require.NoError(t, err)
	actor, err := platformauth.NewHumanActor(adminID).WithWorkspace(workspaceID)
	require.NoError(t, err)
	keyRef := developeroauthdomain.DigestKeyRef{ID: "application-integration"}
	tokens, err := developeroauth.NewTokenManager(developeroauth.TokenKeyringConfig{
		Active: keyRef,
		Keys:   []developeroauth.DigestKey{{Ref: keyRef, Material: bytes.Repeat([]byte{0x72}, 32)}},
	})
	require.NoError(t, err)
	store, err := New(pool)
	require.NoError(t, err)
	manager, err := developeroauth.NewApplicationManager(
		store, tokens, clock, developeroauth.RandomIDGenerator{}, applicationActorResource,
	)
	require.NoError(t, err)
	service, err := developeroauth.New(
		store, tokens, clock, developeroauth.RandomIDGenerator{}, developeroauth.Config{
			Issuer: "https://api.fortyone.app", Resource: applicationActorResource,
			ScopePolicy:           developeroauth.PublicAPIResourceScopePolicy(),
			AccessTokenSigningKey: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ApplicationActors:     store, ApplicationActorScopes: developeroauth.PublicAPIApplicationActorScopePolicy(),
		},
	)
	require.NoError(t, err)
	return applicationActorHarness{
		pool: pool, clock: clock, manager: manager, service: service,
		access: developeroauthdomain.ManagementAccess{
			Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: authorization.WorkspaceRoleAdmin,
		},
	}
}

func createManagedApplication(
	t *testing.T,
	ctx context.Context,
	harness applicationActorHarness,
) developeroauthdomain.IssuedManagedApplication {
	t.Helper()
	issued, err := harness.manager.CreateManagedApplication(ctx, harness.access, developeroauth.CreateManagedApplicationInput{
		Name: "Integration application", ExpiresAt: harness.clock.Now().Add(30 * 24 * time.Hour),
		SecretExpiresAt: harness.clock.Now().Add(14 * 24 * time.Hour), RequestID: uuid.NewString(),
	})
	require.NoError(t, err)
	return issued
}

func exchangeApplicationToken(
	t *testing.T,
	ctx context.Context,
	harness applicationActorHarness,
	clientID string,
	clientSecret string,
	installationID uuid.UUID,
	requestID string,
) developeroauthdomain.ApplicationAccessToken {
	t.Helper()
	token, err := harness.service.ExchangeClientCredentials(ctx, developeroauth.ClientCredentialsExchange{
		ClientID: clientID, ClientSecret: clientSecret, InstallationID: installationID,
		Resource: applicationActorResource, Scopes: []string{string(platformauth.ScopeStoriesWrite)},
		RequestID: requestID,
	})
	require.NoError(t, err)
	return token
}

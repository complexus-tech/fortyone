//go:build integration

package developercredentialsrepository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/migrations"
	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestEnsureHumanPrincipalConvergesConcurrentFirstUse(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := newCredentialFixture(t, ctx, postgres.Pool)
	service := newIntegrationCredentialService(t, postgres.Pool, testkit.NewFixedClock(
		time.Date(2026, 8, 28, 9, 0, 0, 0, time.UTC),
	))

	type result struct {
		principalID uuid.UUID
		err         error
	}
	const callers = 4
	start := make(chan struct{})
	results := make(chan result, callers)
	var waitGroup sync.WaitGroup
	for range callers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			principalID, err := service.EnsureHumanPrincipal(ctx, fixture.adminAccessA, developercredentials.EnsureHumanPrincipalInput{
				RequestID: "concurrent-webhook-owner-resolution",
			})
			results <- result{principalID: principalID, err: err}
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	var resolvedID uuid.UUID
	for outcome := range results {
		require.NoError(t, outcome.err)
		require.NotEqual(t, uuid.Nil, outcome.principalID)
		if resolvedID == uuid.Nil {
			resolvedID = outcome.principalID
		}
		require.Equal(t, resolvedID, outcome.principalID)
	}

	var principalCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM principals
		WHERE workspace_id = $1
		  AND subject_user_id = $2
		  AND kind = 'human_user'
	`, fixture.workspaceA, fixture.adminA).Scan(&principalCount))
	require.Equal(t, 1, principalCount)
}

func TestPersonalTokensEnforceTenantMembershipAndCoarseUsage(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := newCredentialFixture(t, ctx, postgres.Pool)
	clock := testkit.NewManualClock(time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC))
	service := newIntegrationCredentialService(t, postgres.Pool, clock)

	issued, err := service.CreatePersonalToken(ctx, fixture.adminAccessA, developercredentials.CreatePersonalTokenInput{
		Name: "local CLI", Scopes: []platformauth.Scope{platformauth.ScopeStoriesRead},
		TeamIDs: []uuid.UUID{fixture.teamA}, ExpiresAt: clock.Now().Add(24 * time.Hour),
		RequestID: "request-create-pat",
	})
	require.NoError(t, err)
	require.Contains(t, issued.Token.Reveal(), "f41_pat_v1_")
	require.Len(t, issued.Credential.Scopes, 1)
	require.Equal(t, []uuid.UUID{fixture.teamA}, issued.Credential.TeamIDs)

	var digest []byte
	var prefix string
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT secret_digest, lookup_prefix
		FROM api_credentials
		WHERE credential_id = $1
	`, issued.Credential.ID).Scan(&digest, &prefix))
	require.Len(t, digest, 32)
	require.Equal(t, issued.Credential.LookupPrefix, prefix)
	require.NotContains(t, string(digest), issued.Token.Reveal())

	actor, err := service.ResolveMachineCredential(ctx, issued.Token.Reveal())
	require.NoError(t, err)
	require.Equal(t, platformauth.PrincipalPersonalToken, actor.Kind)
	require.Equal(t, fixture.adminA, actor.PrincipalID)
	require.Equal(t, issued.Credential.ID, actor.CredentialID)
	require.Equal(t, fixture.workspaceA, actor.WorkspaceID)
	require.True(t, actor.TeamAccess.Allows(fixture.teamA))
	require.False(t, actor.TeamAccess.Allows(fixture.teamB))
	principalRecordID, err := service.EnsureHumanPrincipal(ctx, developercredentialsdomain.Access{
		Actor: actor, WorkspaceID: fixture.workspaceA, WorkspaceRole: authorization.WorkspaceRoleAdmin,
	}, developercredentials.EnsureHumanPrincipalInput{RequestID: "webhook-owner-resolution"})
	require.NoError(t, err)
	require.Equal(t, issued.Credential.PrincipalID, principalRecordID)
	assertCredentialLastUsedAt(t, ctx, postgres.Pool, issued.Credential.ID, clock.Now())

	clock.Advance(5 * time.Minute)
	_, err = service.ResolveMachineCredential(ctx, issued.Token.Reveal())
	require.NoError(t, err)
	assertCredentialLastUsedAt(t, ctx, postgres.Pool, issued.Credential.ID, clock.Now().Add(-5*time.Minute))

	clock.Advance(11 * time.Minute)
	_, err = service.ResolveMachineCredential(ctx, issued.Token.Reveal())
	require.NoError(t, err)
	assertCredentialLastUsedAt(t, ctx, postgres.Pool, issued.Credential.ID, clock.Now())

	err = service.RevokePersonalToken(ctx, fixture.adminAccessB, issued.Credential.ID, developercredentials.RevokeCredentialInput{
		Reason: "cross_tenant_attempt",
	})
	require.ErrorIs(t, err, developercredentialsdomain.ErrCredentialNotFound)
	_, err = service.ResolveMachineCredential(ctx, issued.Token.Reveal())
	require.NoError(t, err)
	require.NoError(t, service.RevokePersonalToken(ctx, fixture.adminAccessA, issued.Credential.ID, developercredentials.RevokeCredentialInput{
		Reason: "credential_compromised",
	}))
	_, err = service.ResolveMachineCredential(ctx, issued.Token.Reveal())
	require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)

	expiring, err := service.CreatePersonalToken(ctx, fixture.adminAccessA, developercredentials.CreatePersonalTokenInput{
		Name: "short-lived", Scopes: []platformauth.Scope{platformauth.ScopeStoriesRead},
		ExpiresAt: clock.Now().Add(time.Minute),
	})
	require.NoError(t, err)
	clock.Advance(time.Minute)
	_, err = service.ResolveMachineCredential(ctx, expiring.Token.Reveal())
	require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)

	membershipBound, err := service.CreatePersonalToken(ctx, fixture.adminAccessA, developercredentials.CreatePersonalTokenInput{
		Name: "membership-bound", Scopes: []platformauth.Scope{platformauth.ScopeStoriesRead},
		ExpiresAt: clock.Now().Add(time.Hour),
	})
	require.NoError(t, err)

	_, err = postgres.Pool.Exec(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, fixture.workspaceA, fixture.adminA)
	require.NoError(t, err)
	_, err = service.ResolveMachineCredential(ctx, membershipBound.Token.Reveal())
	require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)
}

func TestServiceAccountKeysEnforceLeastPrivilegeRotationAndAuditImmutability(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := newCredentialFixture(t, ctx, postgres.Pool)
	clock := testkit.NewManualClock(time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	service := newIntegrationCredentialService(t, postgres.Pool, clock)

	account, err := service.CreateServiceAccount(ctx, fixture.adminAccessA, developercredentials.CreateServiceAccountInput{
		Name: "deployment bot", WorkspaceRole: authorization.WorkspaceRoleMember,
	})
	require.NoError(t, err)
	issued, err := service.CreateServiceAccountKey(ctx, fixture.adminAccessA, account.ID, developercredentials.CreateServiceAccountKeyInput{
		Name: "production", Scopes: []platformauth.Scope{platformauth.ScopeStoriesRead},
		TeamIDs: []uuid.UUID{fixture.teamA}, ExpiresAt: clock.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	actor, err := service.ResolveMachineCredential(ctx, issued.Token.Reveal())
	require.NoError(t, err)
	require.Equal(t, platformauth.PrincipalServiceAccount, actor.Kind)
	require.Equal(t, account.ID, actor.PrincipalID)
	require.Equal(t, issued.Credential.ID, actor.CredentialID)

	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO api_credential_scopes (credential_id, scope)
		VALUES ($1, 'service_accounts:manage')
	`, issued.Credential.ID)
	assertPostgresConstraint(t, err, "api_credential_scopes_service_account_management_check")
	_, err = postgres.Pool.Exec(ctx, `
		UPDATE api_credentials SET kind = 'personal_access_token'
		WHERE credential_id = $1
	`, issued.Credential.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "identity and secret fields are immutable")

	_, err = postgres.Pool.Exec(ctx, `
		UPDATE principals SET workspace_role = 'admin'
		WHERE principal_id = $1
	`, account.ID)
	assertPostgresConstraint(t, err, "principals_service_account_role_check")

	const rotations = 2
	start := make(chan struct{})
	errorsChannel := make(chan error, rotations)
	results := make(chan developercredentialsdomain.IssuedCredential, rotations)
	var waitGroup sync.WaitGroup
	for range rotations {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			result, rotateErr := service.RotateServiceAccountKey(context.Background(), fixture.adminAccessA, account.ID, issued.Credential.ID, developercredentials.RotateServiceAccountKeyInput{
				ExpiresAt: clock.Now().Add(48 * time.Hour),
			})
			results <- result
			errorsChannel <- rotateErr
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	close(errorsChannel)

	successes := 0
	conflicts := 0
	for rotateErr := range errorsChannel {
		switch {
		case rotateErr == nil:
			successes++
		case errors.Is(rotateErr, developercredentialsdomain.ErrCredentialRotationConflict),
			errors.Is(rotateErr, developercredentialsdomain.ErrConcurrentUpdate):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent rotation error: %v", rotateErr)
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
	var rotated developercredentialsdomain.IssuedCredential
	for result := range results {
		if result.Token.Reveal() != "" {
			rotated = result
		}
	}
	require.NotEmpty(t, rotated.Token.Reveal())
	var rotatedCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM api_credentials WHERE rotated_from_id = $1
	`, issued.Credential.ID).Scan(&rotatedCount))
	require.Equal(t, 1, rotatedCount)
	_, err = service.ResolveMachineCredential(ctx, issued.Token.Reveal())
	require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)
	_, err = service.ResolveMachineCredential(ctx, rotated.Token.Reveal())
	require.NoError(t, err)

	require.NoError(t, service.DisableServiceAccount(ctx, fixture.adminAccessA, account.ID, developercredentials.RevokeCredentialInput{
		Reason: "retired",
	}))
	_, err = service.ResolveMachineCredential(ctx, rotated.Token.Reveal())
	require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)

	_, err = postgres.Pool.Exec(ctx, `UPDATE developer_credential_audit_events SET result = 'failed'`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "immutable")
	_, err = postgres.Pool.Exec(ctx, `DELETE FROM developer_credential_audit_events`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "immutable")

	_, err = postgres.Pool.Exec(ctx, `
		UPDATE workspace_members
		SET role = 'member'
		WHERE workspace_id = $1 AND user_id = $2
	`, fixture.workspaceA, fixture.adminA)
	require.NoError(t, err)
	_, err = service.ListServiceAccounts(ctx, fixture.adminAccessA)
	require.ErrorIs(t, err, developercredentialsdomain.ErrAccessDenied)
	_, err = service.ListServiceAccountKeys(ctx, fixture.adminAccessA, account.ID)
	require.ErrorIs(t, err, developercredentialsdomain.ErrAccessDenied)
}

func TestDeveloperCredentialMigrationRollsBackAtEmptyBoundary(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	for _, migrationName := range []string{
		"000160_code_host_webhook_fencing.down.sql",
		"000159_outbound_developer_webhooks.down.sql",
		"000158_developer_credentials.down.sql",
	} {
		script, err := migrations.FS.ReadFile(migrationName)
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, string(script))
		require.NoError(t, err, migrationName)
	}

	for _, relation := range []string{
		"public.principals",
		"public.api_credentials",
		"public.api_credential_scopes",
		"public.api_credential_team_restrictions",
		"public.developer_credential_audit_events",
	} {
		var absent bool
		require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT to_regclass($1) IS NULL`, relation).Scan(&absent))
		require.True(t, absent, relation)
	}

	var legacyActorForeignKeyPreserved bool
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint AS constraint_record
			INNER JOIN pg_class AS relation
				ON relation.oid = constraint_record.conrelid
			WHERE relation.relname = 'audit_events'
			  AND constraint_record.conname = 'audit_events_actor_id_fkey'
			  AND constraint_record.contype = 'f'
		)
	`).Scan(&legacyActorForeignKeyPreserved))
	require.True(t, legacyActorForeignKeyPreserved)
}

func assertPostgresConstraint(t *testing.T, err error, constraint string) {
	t.Helper()
	require.Error(t, err)
	var postgresError *pgconn.PgError
	require.ErrorAs(t, err, &postgresError)
	require.Equal(t, constraint, postgresError.ConstraintName)
}

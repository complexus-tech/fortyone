//go:build integration

package developercredentialsrepository

import (
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/migrations"
	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestTeamDeletionRevokesOnlyExhaustedCredentials(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := newCredentialFixture(t, ctx, postgres.Pool)
	remainingTeam := insertCredentialTeam(t, ctx, postgres.Pool, fixture.workspaceA, "remaining")
	insertCredentialTeamMember(t, ctx, postgres.Pool, remainingTeam, fixture.adminA)
	clock := testkit.NewFixedClock(time.Now().UTC())
	service := newIntegrationCredentialService(t, postgres.Pool, clock)

	type credentials struct {
		single    developercredentialsdomain.IssuedCredential
		multiple  developercredentialsdomain.IssuedCredential
		allTeams  developercredentialsdomain.IssuedCredential
		unrelated developercredentialsdomain.IssuedCredential
		revoked   developercredentialsdomain.IssuedCredential
	}
	issuedByKind := make(map[developercredentialsdomain.CredentialKind]credentials)
	for _, kind := range []developercredentialsdomain.CredentialKind{
		developercredentialsdomain.CredentialPersonalAccessToken,
		developercredentialsdomain.CredentialServiceAccountKey,
	} {
		issue := teamDeletionCredentialIssuer(t, ctx, service, fixture.adminAccessA, kind, clock.Now())
		issued := credentials{
			single: issue([]uuid.UUID{fixture.teamA}), multiple: issue([]uuid.UUID{fixture.teamA, remainingTeam}),
			allTeams: issue(nil), unrelated: issue([]uuid.UUID{remainingTeam}), revoked: issue([]uuid.UUID{fixture.teamA}),
		}
		_, err := postgres.Pool.Exec(ctx, `
			UPDATE api_credentials
			SET revoked_at = CURRENT_TIMESTAMP, revoked_reason = 'previously_revoked'
			WHERE credential_id = $1
		`, issued.revoked.Credential.ID)
		require.NoError(t, err)
		issuedByKind[kind] = issued
	}

	_, err := postgres.Pool.Exec(ctx, `DELETE FROM teams WHERE team_id = $1`, fixture.teamA)
	require.NoError(t, err)
	for kind, issued := range issuedByKind {
		t.Run(string(kind), func(t *testing.T) {
			_, err := service.ResolveMachineCredential(ctx, issued.single.Token.Reveal())
			require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)
			assertTeamDeletionRevocation(t, ctx, postgres.Pool, issued.single.Credential.ID, fixture.teamA)

			for _, surviving := range []developercredentialsdomain.IssuedCredential{issued.multiple, issued.unrelated} {
				actor, err := service.ResolveMachineCredential(ctx, surviving.Token.Reveal())
				require.NoError(t, err)
				require.True(t, actor.TeamAccess.Allows(remainingTeam))
				require.False(t, actor.TeamAccess.Allows(fixture.teamA))
				require.False(t, actor.TeamAccess.Allows(uuid.New()))
			}
			actor, err := service.ResolveMachineCredential(ctx, issued.allTeams.Token.Reveal())
			require.NoError(t, err)
			require.True(t, actor.TeamAccess.Allows(remainingTeam))
			require.True(t, actor.TeamAccess.Allows(uuid.New()))

			var priorReason string
			require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT revoked_reason FROM api_credentials WHERE credential_id = $1`, issued.revoked.Credential.ID).Scan(&priorReason))
			require.Equal(t, "previously_revoked", priorReason)
			var auditCount int
			require.NoError(t, postgres.Pool.QueryRow(ctx, `
				SELECT COUNT(*) FROM developer_credential_audit_events
				WHERE subject_id = $1 AND reason_code = 'team_restrictions_exhausted'
			`, issued.revoked.Credential.ID).Scan(&auditCount))
			require.Zero(t, auditCount)
		})
	}

	_, err = postgres.Pool.Exec(ctx, `DELETE FROM teams WHERE team_id = $1`, remainingTeam)
	require.NoError(t, err)
	for _, issued := range issuedByKind {
		_, err := service.ResolveMachineCredential(ctx, issued.multiple.Token.Reveal())
		require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)
		assertTeamDeletionRevocation(t, ctx, postgres.Pool, issued.multiple.Credential.ID, remainingTeam)
		_, err = service.ResolveMachineCredential(ctx, issued.allTeams.Token.Reveal())
		require.NoError(t, err)
	}
}

func TestTeamDeletionCredentialRevocationIsAtomicAndSurvivesMigrationRollback(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := newCredentialFixture(t, ctx, postgres.Pool)
	clock := testkit.NewFixedClock(time.Now().UTC())
	service := newIntegrationCredentialService(t, postgres.Pool, clock)
	issue := teamDeletionCredentialIssuer(t, ctx, service, fixture.adminAccessA, developercredentialsdomain.CredentialPersonalAccessToken, clock.Now())
	issued := issue([]uuid.UUID{fixture.teamA})

	tx, err := postgres.Pool.Begin(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	_, err = tx.Exec(ctx, `DELETE FROM teams WHERE team_id = $1`, fixture.teamA)
	require.NoError(t, err)
	var revoked bool
	require.NoError(t, tx.QueryRow(ctx, `SELECT revoked_at IS NOT NULL FROM api_credentials WHERE credential_id = $1`, issued.Credential.ID).Scan(&revoked))
	require.True(t, revoked)
	require.NoError(t, tx.Rollback(ctx))
	_, err = service.ResolveMachineCredential(ctx, issued.Token.Reveal())
	require.NoError(t, err)
	var auditCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM developer_credential_audit_events
		WHERE subject_id = $1 AND reason_code = 'team_restrictions_exhausted'
	`, issued.Credential.ID).Scan(&auditCount))
	require.Zero(t, auditCount)

	_, err = postgres.Pool.Exec(ctx, `DELETE FROM teams WHERE team_id = $1`, fixture.teamA)
	require.NoError(t, err)
	assertTeamDeletionRevocation(t, ctx, postgres.Pool, issued.Credential.ID, fixture.teamA)
	for _, migrationName := range []string{
		"000190_team_credential_deletion_safety.down.sql",
		"000190_team_credential_deletion_safety.up.sql",
	} {
		script, err := migrations.FS.ReadFile(migrationName)
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, string(script))
		require.NoError(t, err)
		_, err = service.ResolveMachineCredential(ctx, issued.Token.Reveal())
		require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)
		assertTeamDeletionRevocation(t, ctx, postgres.Pool, issued.Credential.ID, fixture.teamA)
	}
}

func TestConcurrentTeamDeletionsCannotBroadenCredentialAccess(t *testing.T) {
	for _, isolation := range []pgx.TxIsoLevel{pgx.ReadCommitted, pgx.RepeatableRead, pgx.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			postgres := testkit.NewPostgres(t)
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			fixture := newCredentialFixture(t, ctx, postgres.Pool)
			secondTeam := insertCredentialTeam(t, ctx, postgres.Pool, fixture.workspaceA, "second")
			clock := testkit.NewFixedClock(time.Now().UTC())
			service := newIntegrationCredentialService(t, postgres.Pool, clock)
			issue := teamDeletionCredentialIssuer(t, ctx, service, fixture.adminAccessA, developercredentialsdomain.CredentialServiceAccountKey, clock.Now())
			issued := issue([]uuid.UUID{fixture.teamA, secondTeam})

			first, err := postgres.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolation})
			require.NoError(t, err)
			t.Cleanup(func() { _ = first.Rollback(context.Background()) })
			second, err := postgres.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolation})
			require.NoError(t, err)
			t.Cleanup(func() { _ = second.Rollback(context.Background()) })
			// Establish the second transaction's snapshot before the first removal.
			var restrictionCount int
			require.NoError(t, second.QueryRow(ctx, `SELECT COUNT(*) FROM api_credential_team_restrictions WHERE credential_id = $1`, issued.Credential.ID).Scan(&restrictionCount))
			require.Equal(t, 2, restrictionCount)
			_, err = first.Exec(ctx, `DELETE FROM teams WHERE team_id = $1`, fixture.teamA)
			require.NoError(t, err)
			secondPID := second.Conn().PgConn().PID()
			result := make(chan error, 1)
			go func() {
				_, deleteErr := second.Exec(ctx, `DELETE FROM teams WHERE team_id = $1`, secondTeam)
				result <- deleteErr
			}()
			require.Eventually(t, func() bool {
				var blocked bool
				err := postgres.Pool.QueryRow(ctx, `SELECT wait_event_type = 'Lock' FROM pg_stat_activity WHERE pid = $1`, secondPID).Scan(&blocked)
				return err == nil && blocked
			}, 5*time.Second, 10*time.Millisecond, "the second removal must serialize on the credential")
			require.NoError(t, first.Commit(ctx))
			deleteErr := <-result
			if isolation == pgx.ReadCommitted {
				require.NoError(t, deleteErr)
				require.NoError(t, second.Commit(ctx))
			} else {
				var databaseErr *pgconn.PgError
				require.ErrorAs(t, deleteErr, &databaseErr)
				require.Equal(t, "40001", databaseErr.Code)
				require.NoError(t, second.Rollback(ctx))
				_, err = postgres.Pool.Exec(ctx, `DELETE FROM teams WHERE team_id = $1`, secondTeam)
				require.NoError(t, err)
			}
			_, err = service.ResolveMachineCredential(ctx, issued.Token.Reveal())
			require.ErrorIs(t, err, developercredentialsdomain.ErrAuthenticationFailed)
			assertTeamDeletionRevocation(t, ctx, postgres.Pool, issued.Credential.ID, secondTeam)
		})
	}
}

func teamDeletionCredentialIssuer(
	t *testing.T,
	ctx context.Context,
	service *developercredentials.Service,
	access developercredentialsdomain.Access,
	kind developercredentialsdomain.CredentialKind,
	now time.Time,
) func([]uuid.UUID) developercredentialsdomain.IssuedCredential {
	t.Helper()
	var accountID uuid.UUID
	if kind == developercredentialsdomain.CredentialServiceAccountKey {
		account, err := service.CreateServiceAccount(ctx, access, developercredentials.CreateServiceAccountInput{
			Name: "team-deletion bot", WorkspaceRole: authorization.WorkspaceRoleMember,
		})
		require.NoError(t, err)
		accountID = account.ID
	}
	return func(teamIDs []uuid.UUID) developercredentialsdomain.IssuedCredential {
		t.Helper()
		var issued developercredentialsdomain.IssuedCredential
		var err error
		if kind == developercredentialsdomain.CredentialPersonalAccessToken {
			issued, err = service.CreatePersonalToken(ctx, access, developercredentials.CreatePersonalTokenInput{
				Name: "team-deletion token", Scopes: []platformauth.Scope{platformauth.ScopeStoriesRead},
				TeamIDs: teamIDs, ExpiresAt: now.Add(24 * time.Hour),
			})
		} else {
			issued, err = service.CreateServiceAccountKey(ctx, access, accountID, developercredentials.CreateServiceAccountKeyInput{
				Name: "team-deletion key", Scopes: []platformauth.Scope{platformauth.ScopeStoriesRead},
				TeamIDs: teamIDs, ExpiresAt: now.Add(24 * time.Hour),
			})
		}
		require.NoError(t, err)
		return issued
	}
}

func assertTeamDeletionRevocation(t *testing.T, ctx context.Context, pool *pgxpool.Pool, credentialID, removedTeamID uuid.UUID) {
	t.Helper()
	var revokedAt *time.Time
	var revokedBy *uuid.UUID
	var reason string
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT revoked_at, revoked_by_user_id, revoked_reason
		FROM api_credentials WHERE credential_id = $1
	`, credentialID).Scan(&revokedAt, &revokedBy, &reason))
	require.NotNil(t, revokedAt)
	require.Nil(t, revokedBy)
	require.Equal(t, "team_restrictions_exhausted", reason)
	var eventCount int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM developer_credential_audit_events
		WHERE subject_id = $1
		  AND actor_kind = 'system'
		  AND actor_id = '00000000-0000-0000-0000-000000000000'
		  AND result = 'succeeded'
		  AND reason_code = 'team_restrictions_exhausted'
		  AND metadata->>'removed_team_id' = CAST($2 AS text)
	`, credentialID, removedTeamID.String()).Scan(&eventCount))
	require.Equal(t, 1, eventCount)
}

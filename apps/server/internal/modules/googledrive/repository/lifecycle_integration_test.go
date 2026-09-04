//go:build integration

package googledriverepository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestGoogleDriveLifecycleCleanupOnPostgres18(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()

	t.Run("provider admission preserves pool headroom and nested gates reuse their session", func(t *testing.T) {
		poolConfig, err := pgxpool.ParseConfig(postgres.DatabaseURL)
		require.NoError(t, err)
		poolConfig.MaxConns = 2
		poolConfig.MinConns = 0
		smallPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		require.NoError(t, err)
		defer smallPool.Close()

		repository := New(smallPool)
		flowCtx, cancelFlow := context.WithTimeout(ctx, 5*time.Second)
		defer cancelFlow()
		firstEntered := make(chan struct{})
		releaseFirst := make(chan struct{})
		firstResult := make(chan error, 1)
		go func() {
			firstResult <- repository.WithinProviderUserLifecycle(flowCtx, uuid.New(), func(userCtx context.Context) error {
				close(firstEntered)
				select {
				case <-releaseFirst:
				case <-flowCtx.Done():
					return flowCtx.Err()
				}
				return repository.WithinProviderSubjectLifecycle(
					userCtx,
					"small-pool-subject",
					func(subjectCtx context.Context) error {
						_, err := repository.GetActiveAccountBySubject(subjectCtx, "absent-small-pool-subject")
						if err == nil {
							return errors.New("expected absent Google Drive subject")
						}
						if !errors.Is(err, domain.ErrNotFound) {
							return fmt.Errorf("query through provider lifecycle session: %w", err)
						}
						return repository.withinTransaction(
							subjectCtx,
							func(googledrivesql.Querier) error { return nil },
						)
					},
				)
			})
		}()
		select {
		case <-firstEntered:
		case <-flowCtx.Done():
			require.FailNow(t, "provider lifecycle flow did not reserve its session", flowCtx.Err().Error())
		}

		ordinaryConnection, err := smallPool.Acquire(flowCtx)
		require.NoError(t, err)
		defer ordinaryConnection.Release()
		baselineEmptyAcquires := smallPool.Stat().EmptyAcquireCount()

		secondCtx, cancelSecond := context.WithTimeout(flowCtx, 250*time.Millisecond)
		defer cancelSecond()
		secondStarted := make(chan struct{})
		secondEntered := make(chan struct{}, 1)
		secondResult := make(chan error, 1)
		go func() {
			close(secondStarted)
			secondResult <- repository.WithinProviderUserLifecycle(secondCtx, uuid.New(), func(context.Context) error {
				secondEntered <- struct{}{}
				return nil
			})
		}()
		<-secondStarted

		var one int
		require.NoError(t, ordinaryConnection.QueryRow(flowCtx, "SELECT 1").Scan(&one))
		require.Equal(t, 1, one)
		require.ErrorIs(t, <-secondResult, context.DeadlineExceeded)
		require.Equal(t, baselineEmptyAcquires, smallPool.Stat().EmptyAcquireCount(),
			"a lifecycle operation waiting for admission must not wait on the shared pool")
		select {
		case <-secondEntered:
			require.FailNow(t, "lifecycle operation entered after its admission context expired")
		default:
		}

		close(releaseFirst)
		require.NoError(t, <-firstResult)
	})

	t.Run("membership removal destroys only the final binding credential", func(t *testing.T) {
		userID, workspaceIDs := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 2)
		accountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, userID, workspaceIDs)

		_, err := postgres.Pool.Exec(ctx, `
			DELETE FROM public.workspace_members
			WHERE workspace_id = $1 AND user_id = $2
		`, workspaceIDs[0], userID)
		require.NoError(t, err)
		assertGoogleDriveAccountLifecycle(t, ctx, postgres.Pool, accountID, 1, "vault.v2.test-envelope", false)
		assertGoogleDriveRevocationCount(t, ctx, postgres.Pool, accountID, 0)

		_, err = postgres.Pool.Exec(ctx, `
			DELETE FROM public.workspace_members
			WHERE workspace_id = $1 AND user_id = $2
		`, workspaceIDs[1], userID)
		require.NoError(t, err)
		assertGoogleDriveAccountLifecycle(t, ctx, postgres.Pool, accountID, 0, "", true)
		assertGoogleDriveRevocation(t, ctx, postgres.Pool, accountID, userID)
	})

	t.Run("workspace disconnect purges only that binding's Picker grants", func(t *testing.T) {
		userID, workspaceIDs := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 2)
		accountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, userID, workspaceIDs)
		fileIDs := make([]uuid.UUID, len(workspaceIDs))
		for index, workspaceID := range workspaceIDs {
			fileID := uuid.New()
			fileIDs[index] = fileID
			_, err := postgres.Pool.Exec(ctx, `
				INSERT INTO public.google_drive_files (
					file_id, workspace_id, google_file_id, name, mime_type, web_view_link
				) VALUES ($1, $2, $3, 'Lifecycle file', 'application/vnd.google-apps.document', $4)
			`, fileID, workspaceID, "provider-"+fileID.String(), "https://docs.google.com/document/d/"+fileID.String()+"/edit")
			require.NoError(t, err)
			_, err = postgres.Pool.Exec(ctx, `
				INSERT INTO public.google_drive_file_grants (file_id, user_id, account_id)
				VALUES ($1, $2, $3)
			`, fileID, userID, accountID)
			require.NoError(t, err)
		}

		repository := New(postgres.Pool)
		orphaned, err := repository.Disconnect(ctx, workspaceIDs[0], userID)
		require.NoError(t, err)
		require.False(t, orphaned)

		var disconnectedGrants, connectedGrants int
		require.NoError(t, postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM public.google_drive_file_grants WHERE file_id = $1
		`, fileIDs[0]).Scan(&disconnectedGrants))
		require.NoError(t, postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM public.google_drive_file_grants WHERE file_id = $1
		`, fileIDs[1]).Scan(&connectedGrants))
		require.Zero(t, disconnectedGrants)
		require.Equal(t, 1, connectedGrants)
		assertGoogleDriveRevocationCount(t, ctx, postgres.Pool, accountID, 0)
	})

	t.Run("concurrent explicit disconnects have exactly one final binding", func(t *testing.T) {
		userID, workspaceIDs := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 2)
		accountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, userID, workspaceIDs)
		repository := New(postgres.Pool)

		type disconnectResult struct {
			orphaned bool
			err      error
		}
		results := make(chan disconnectResult, len(workspaceIDs))
		for _, workspaceID := range workspaceIDs {
			workspaceID := workspaceID
			go func() {
				orphaned, err := repository.Disconnect(ctx, workspaceID, userID)
				results <- disconnectResult{orphaned: orphaned, err: err}
			}()
		}

		finalBindings := 0
		for range workspaceIDs {
			result := <-results
			require.NoError(t, result.err)
			if result.orphaned {
				finalBindings++
			}
		}
		require.Equal(t, 1, finalBindings)
		assertGoogleDriveAccountLifecycle(t, ctx, postgres.Pool, accountID, 0, "", true)
		assertGoogleDriveRevocation(t, ctx, postgres.Pool, accountID, userID)
	})

	t.Run("callback persistence rechecks a downgraded workspace role", func(t *testing.T) {
		_, workspaceIDs := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 1)
		workspaceID := workspaceIDs[0]
		userID := uuid.New()
		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO public.users (user_id, username, email)
			VALUES ($1, $2, $3)
		`, userID, "drive-downgraded-"+userID.String(), userID.String()+"@example.com")
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO public.workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, 'member')
		`, workspaceID, userID)
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, `
			UPDATE public.workspace_members
			SET role = 'guest'
			WHERE workspace_id = $1 AND user_id = $2
		`, workspaceID, userID)
		require.NoError(t, err)

		repository := New(postgres.Pool)
		_, err = repository.UpsertConnection(ctx, workspaceID, domain.Account{
			UserID: userID, GoogleSubject: "guest-subject-" + userID.String(),
			Email: userID.String() + "@example.com", CredentialPayload: "vault.v2.guest-envelope",
			CredentialVersion: 2, InstallationGeneration: uuid.New(),
			Scopes: []string{"https://www.googleapis.com/auth/drive.file"}, ExpiresAt: time.Now().Add(time.Hour),
		})
		require.ErrorIs(t, err, domain.ErrForbidden)

		var accountCount int
		require.NoError(t, postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM public.google_drive_accounts
			WHERE user_id = $1 AND google_subject = $2
		`, userID, "guest-subject-"+userID.String()).Scan(&accountCount))
		require.Zero(t, accountCount)
	})

	t.Run("same-subject reconnect supersedes the staged generation", func(t *testing.T) {
		userID, workspaceIDs := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 1)
		accountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, userID, workspaceIDs)
		repository := New(postgres.Pool)
		orphaned, err := repository.Disconnect(ctx, workspaceIDs[0], userID)
		require.NoError(t, err)
		require.True(t, orphaned)

		newGeneration := uuid.New()
		_, err = repository.UpsertConnection(ctx, workspaceIDs[0], domain.Account{
			UserID: userID, GoogleSubject: "subject-" + accountID.String(),
			Email: userID.String() + "@example.com", CredentialPayload: "vault.v2.new-envelope",
			CredentialVersion: 2, InstallationGeneration: newGeneration,
			Scopes: []string{"https://www.googleapis.com/auth/drive.file"}, ExpiresAt: time.Now().Add(time.Hour),
		})
		require.NoError(t, err)

		var status string
		var payload *string
		err = postgres.Pool.QueryRow(ctx, `
			SELECT status, credential_payload
			FROM public.google_drive_revocation_outbox
			WHERE source_account_id = $1
		`, accountID).Scan(&status, &payload)
		require.NoError(t, err)
		require.Equal(t, "superseded", status)
		require.Nil(t, payload)
	})

	t.Run("one Google subject has one active FortyOne owner", func(t *testing.T) {
		firstUserID, firstWorkspaces := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 1)
		accountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, firstUserID, firstWorkspaces)
		secondUserID, _ := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 1)

		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO public.google_drive_accounts (
				user_id, google_subject, email, credential_payload,
				credential_key_version, installation_generation, scopes, expires_at
			) VALUES ($1, $2, $3, 'vault.v2.second-envelope', 2, $4,
			          ARRAY['https://www.googleapis.com/auth/drive.file'], $5)
		`, secondUserID, "subject-"+accountID.String(), secondUserID.String()+"@example.com", uuid.New(), time.Now().Add(time.Hour))
		require.Error(t, err)
	})

	t.Run("user deletion preserves the encrypted remote revocation tombstone", func(t *testing.T) {
		_, ownerWorkspaces := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 1)
		workspaceID := ownerWorkspaces[0]
		userID := uuid.New()
		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO public.users (user_id, username, email)
			VALUES ($1, $2, $3)
		`, userID, "drive-deleted-"+userID.String(), userID.String()+"@example.com")
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO public.workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, 'member')
		`, workspaceID, userID)
		require.NoError(t, err)
		accountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, userID, []uuid.UUID{workspaceID})

		_, err = postgres.Pool.Exec(ctx, `DELETE FROM public.users WHERE user_id = $1`, userID)
		require.NoError(t, err)
		var accountCount int
		require.NoError(t, postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM public.google_drive_accounts WHERE account_id = $1
		`, accountID).Scan(&accountCount))
		require.Zero(t, accountCount)
		assertGoogleDriveRevocation(t, ctx, postgres.Pool, accountID, userID)
	})

	t.Run("soft deactivation tears down Drive state without reactivation revival", func(t *testing.T) {
		userID, workspaceIDs := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 3)
		firstAccountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, userID, workspaceIDs[:2])
		secondAccountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, userID, workspaceIDs[2:])
		users := usersrepository.New(postgres.Pool)
		deactivatedAt := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)

		require.NoError(t, users.DeleteUser(ctx, userID, deactivatedAt))

		var active bool
		var connectionCount int
		require.NoError(t, postgres.Pool.QueryRow(ctx, `
			SELECT is_active FROM public.users WHERE user_id = $1
		`, userID).Scan(&active))
		require.False(t, active)
		require.NoError(t, postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM public.google_drive_workspace_connections
			WHERE user_id = $1
		`, userID).Scan(&connectionCount))
		require.Zero(t, connectionCount)
		for _, accountID := range []uuid.UUID{firstAccountID, secondAccountID} {
			assertGoogleDriveAccountLifecycle(t, ctx, postgres.Pool, accountID, 0, "", true)
			assertGoogleDriveRevocation(t, ctx, postgres.Pool, accountID, userID)
		}

		reactivated, err := users.ReactivateUserForVerifiedSignIn(
			ctx,
			usersdomain.VerifiedSignInReactivation{
				UserID: userID, SignedInAt: deactivatedAt.Add(time.Hour),
			},
		)
		require.NoError(t, err)
		require.True(t, reactivated.IsActive)
		require.NoError(t, postgres.Pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM public.google_drive_workspace_connections
			WHERE user_id = $1
		`, userID).Scan(&connectionCount))
		require.Zero(t, connectionCount)
		for _, accountID := range []uuid.UUID{firstAccountID, secondAccountID} {
			assertGoogleDriveAccountLifecycle(t, ctx, postgres.Pool, accountID, 0, "", true)
			assertGoogleDriveRevocationCount(t, ctx, postgres.Pool, accountID, 1)
		}

		// A later active-to-inactive transition is inert when no connection was
		// explicitly recreated; it must not duplicate the prior generation.
		require.NoError(t, users.DeleteUser(ctx, userID, deactivatedAt.Add(2*time.Hour)))
		assertGoogleDriveRevocationCount(t, ctx, postgres.Pool, firstAccountID, 1)
		assertGoogleDriveRevocationCount(t, ctx, postgres.Pool, secondAccountID, 1)
	})

	t.Run("target cascades remove file metadata and grants after the final reference", func(t *testing.T) {
		userID, workspaceIDs := seedGoogleDriveLifecyclePrincipal(t, ctx, postgres.Pool, 1)
		accountID := seedGoogleDriveLifecycleAccount(t, ctx, postgres.Pool, userID, workspaceIDs)
		workspaceID := workspaceIDs[0]
		documentIDs := []uuid.UUID{uuid.New(), uuid.New()}
		for index, documentID := range documentIDs {
			_, err := postgres.Pool.Exec(ctx, `
				INSERT INTO public.documents (
					document_id, workspace_id, title, created_by, updated_by
				) VALUES ($1, $2, $3, $4, $4)
			`, documentID, workspaceID, fmt.Sprintf("Lifecycle document %d", index), userID)
			require.NoError(t, err)
		}

		fileID := uuid.New()
		_, err := postgres.Pool.Exec(ctx, `
			INSERT INTO public.google_drive_files (
				file_id, workspace_id, google_file_id, name, mime_type, web_view_link
			) VALUES ($1, $2, $3, 'Lifecycle file', 'application/vnd.google-apps.document', $4)
		`, fileID, workspaceID, "provider-"+fileID.String(), "https://docs.google.com/document/d/"+fileID.String()+"/edit")
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO public.google_drive_file_grants (file_id, user_id, account_id)
			VALUES ($1, $2, $3)
		`, fileID, userID, accountID)
		require.NoError(t, err)
		for _, documentID := range documentIDs {
			_, err = postgres.Pool.Exec(ctx, `
				INSERT INTO public.google_drive_file_references (
					workspace_id, file_id, target_type, document_id, created_by_user_id
				) VALUES ($1, $2, 'document', $3, $4)
			`, workspaceID, fileID, documentID, userID)
			require.NoError(t, err)
		}

		_, err = postgres.Pool.Exec(ctx, `DELETE FROM public.documents WHERE document_id = $1`, documentIDs[0])
		require.NoError(t, err)
		assertGoogleDriveFileLifecycle(t, ctx, postgres.Pool, fileID, 1, 1, 1)

		_, err = postgres.Pool.Exec(ctx, `DELETE FROM public.documents WHERE document_id = $1`, documentIDs[1])
		require.NoError(t, err)
		assertGoogleDriveFileLifecycle(t, ctx, postgres.Pool, fileID, 0, 0, 0)
	})
}

func seedGoogleDriveLifecyclePrincipal(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceCount int,
) (uuid.UUID, []uuid.UUID) {
	t.Helper()

	userID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO public.users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "drive-lifecycle-"+userID.String(), userID.String()+"@example.com")
	require.NoError(t, err)

	workspaceIDs := make([]uuid.UUID, 0, workspaceCount)
	for range workspaceCount {
		workspaceID := uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO public.workspaces (workspace_id, name, slug, created_by)
			VALUES ($1, 'Drive lifecycle', $2, $3)
		`, workspaceID, "drive-lifecycle-"+workspaceID.String(), userID)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `
			INSERT INTO public.workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, 'admin')
		`, workspaceID, userID)
		require.NoError(t, err)
		workspaceIDs = append(workspaceIDs, workspaceID)
	}
	return userID, workspaceIDs
}

func seedGoogleDriveLifecycleAccount(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	workspaceIDs []uuid.UUID,
) uuid.UUID {
	t.Helper()

	accountID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO public.google_drive_accounts (
			account_id, user_id, google_subject, email, credential_payload,
			credential_key_version, installation_generation, scopes, expires_at
		) VALUES ($1, $2, $3, $4, 'vault.v2.test-envelope', 2, $5, ARRAY['https://www.googleapis.com/auth/drive.file'], $6)
	`, accountID, userID, "subject-"+accountID.String(), userID.String()+"@example.com", uuid.New(), time.Now().Add(time.Hour))
	require.NoError(t, err)
	for _, workspaceID := range workspaceIDs {
		_, err = pool.Exec(ctx, `
			INSERT INTO public.google_drive_workspace_connections (workspace_id, user_id, account_id)
			VALUES ($1, $2, $3)
		`, workspaceID, userID, accountID)
		require.NoError(t, err)
	}
	return accountID
}

func assertGoogleDriveRevocationCount(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	accountID uuid.UUID,
	want int,
) {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM public.google_drive_revocation_outbox
		WHERE source_account_id = $1
	`, accountID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, want, count)
}

func assertGoogleDriveRevocation(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	accountID, userID uuid.UUID,
) {
	t.Helper()

	var storedAccountID, storedUserID, generation uuid.UUID
	var subject, payload, status string
	err := pool.QueryRow(ctx, `
		SELECT source_account_id, user_id, google_subject,
		       installation_generation, credential_payload, status
		FROM public.google_drive_revocation_outbox
		WHERE source_account_id = $1
	`, accountID).Scan(&storedAccountID, &storedUserID, &subject, &generation, &payload, &status)
	require.NoError(t, err)
	require.Equal(t, accountID, storedAccountID)
	require.Equal(t, userID, storedUserID)
	require.Equal(t, "subject-"+accountID.String(), subject)
	require.NotEqual(t, uuid.Nil, generation)
	require.Equal(t, "vault.v2.test-envelope", payload)
	require.Equal(t, "pending", status)
}

func assertGoogleDriveAccountLifecycle(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	accountID uuid.UUID,
	wantConnections int,
	wantCredential string,
	wantRevoked bool,
) {
	t.Helper()

	var connectionCount int
	var credential string
	var googleSubject string
	var email string
	var scopes []string
	var revoked bool
	err := pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM public.google_drive_workspace_connections WHERE account_id = $1),
			credential_payload,
			google_subject,
			email,
			scopes,
			revoked_at IS NOT NULL
		FROM public.google_drive_accounts
		WHERE account_id = $1
	`, accountID).Scan(&connectionCount, &credential, &googleSubject, &email, &scopes, &revoked)
	require.NoError(t, err)
	require.Equal(t, wantConnections, connectionCount)
	require.Equal(t, wantCredential, credential)
	require.Equal(t, wantRevoked, revoked)
	if wantRevoked {
		require.Empty(t, googleSubject)
		require.Empty(t, email)
		require.Empty(t, scopes)
	}
}

func assertGoogleDriveFileLifecycle(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	fileID uuid.UUID,
	wantFiles, wantReferences, wantGrants int,
) {
	t.Helper()

	var files, references, grants int
	err := pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM public.google_drive_files WHERE file_id = $1),
			(SELECT COUNT(*) FROM public.google_drive_file_references WHERE file_id = $1),
			(SELECT COUNT(*) FROM public.google_drive_file_grants WHERE file_id = $1)
	`, fileID).Scan(&files, &references, &grants)
	require.NoError(t, err)
	require.Equal(t, wantFiles, files)
	require.Equal(t, wantReferences, references)
	require.Equal(t, wantGrants, grants)
}

package googledriverepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestDisconnectCommitsFinalBindingCleanupWithoutProviderCallback(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	accountID := uuid.New()
	queries := &connectionLifecycleQueries{
		accountID: accountID,
		// The database trigger normally revokes the account before the
		// repository's idempotent fallback update executes.
		revokeRows: 0,
	}
	repository := connectionLifecycleRepository(queries)
	orphaned, err := repository.Disconnect(t.Context(), workspaceID, userID)

	require.NoError(t, err)
	require.True(t, orphaned)
	require.Equal(t, []string{"lock", "delete", "count", "revoke"}, queries.calls)
	require.Equal(t, workspaceID, queries.workspaceID)
	require.Equal(t, userID, queries.userID)
	require.Equal(t, accountID, queries.revokedAccountID)
}

func TestDisconnectPreservesCredentialWithAnotherWorkspaceBinding(t *testing.T) {
	t.Parallel()

	queries := &connectionLifecycleQueries{accountID: uuid.New(), connectionCount: 1}
	repository := connectionLifecycleRepository(queries)

	orphaned, err := repository.Disconnect(t.Context(), uuid.New(), uuid.New())

	require.NoError(t, err)
	require.False(t, orphaned)
	require.Equal(t, []string{"lock", "delete", "count"}, queries.calls)
}

func TestDisconnectIsIdempotentWhenBindingIsAlreadyAbsent(t *testing.T) {
	t.Parallel()

	queries := &connectionLifecycleQueries{deleteErr: pgx.ErrNoRows}
	repository := connectionLifecycleRepository(queries)

	orphaned, err := repository.Disconnect(t.Context(), uuid.New(), uuid.New())

	require.NoError(t, err)
	require.False(t, orphaned)
	require.Equal(t, []string{"lock", "delete"}, queries.calls)
}

func TestDisconnectStopsBeforeMutationWhenLifecycleLockFails(t *testing.T) {
	t.Parallel()

	lockErr := errors.New("lock failed")
	queries := &connectionLifecycleQueries{lockErr: lockErr}
	repository := connectionLifecycleRepository(queries)

	orphaned, err := repository.Disconnect(t.Context(), uuid.New(), uuid.New())

	require.ErrorIs(t, err, lockErr)
	require.False(t, orphaned)
	require.Equal(t, []string{"lock"}, queries.calls)
}

func TestUpsertConnectionTakesLifecycleLockBeforeCredentialWrite(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	account := domain.Account{
		UserID: uuid.New(), GoogleSubject: "google-subject", Email: "owner@example.com",
		CredentialPayload: "sealed", CredentialVersion: 2,
		InstallationGeneration: uuid.New(), ExpiresAt: time.Now().Add(time.Hour),
	}
	queries := &connectionLifecycleQueries{accountID: uuid.New(), upsertConnectionRows: 1}
	repository := connectionLifecycleRepository(queries)

	connection, err := repository.UpsertConnection(t.Context(), workspaceID, account)

	require.NoError(t, err)
	require.Equal(t, []string{"lock", "lock_subject", "get_subject", "upsert_account", "upsert_connection", "supersede"}, queries.calls)
	require.Equal(t, workspaceID, connection.WorkspaceID)
	require.Equal(t, account.UserID, queries.userID)
	require.Equal(t, account.GoogleSubject, queries.supersededSubject)
	require.Equal(t, account.InstallationGeneration, queries.supersededGeneration)
}

func TestUpsertConnectionRejectsGoogleSubjectOwnedByAnotherUser(t *testing.T) {
	t.Parallel()

	account := domain.Account{
		UserID: uuid.New(), GoogleSubject: "shared-google-subject", Email: "owner@example.com",
		CredentialPayload: "sealed", CredentialVersion: 2,
		InstallationGeneration: uuid.New(), ExpiresAt: time.Now().Add(time.Hour),
	}
	queries := &connectionLifecycleQueries{
		activeSubjectAccount: &googledrivesql.GetActiveGoogleDriveAccountBySubjectRow{
			UserID: uuid.New(), GoogleSubject: account.GoogleSubject,
		},
	}
	repository := connectionLifecycleRepository(queries)

	_, err := repository.UpsertConnection(t.Context(), uuid.New(), account)

	require.ErrorIs(t, err, domain.ErrAccountOwned)
	require.Equal(t, []string{"lock", "lock_subject", "get_subject"}, queries.calls)
}

type connectionLifecycleQueries struct {
	googledrivesql.Querier

	accountID            uuid.UUID
	connectionCount      int64
	revokeRows           int64
	upsertConnectionRows int64
	lockErr              error
	deleteErr            error
	calls                []string
	workspaceID          uuid.UUID
	userID               uuid.UUID
	revokedAccountID     uuid.UUID
	activeSubjectAccount *googledrivesql.GetActiveGoogleDriveAccountBySubjectRow
	supersededSubject    string
	supersededGeneration uuid.UUID
}

func (queries *connectionLifecycleQueries) UpsertGoogleDriveAccount(
	_ context.Context,
	params googledrivesql.UpsertGoogleDriveAccountParams,
) (googledrivesql.UpsertGoogleDriveAccountRow, error) {
	queries.calls = append(queries.calls, "upsert_account")
	queries.userID = params.UserID
	return googledrivesql.UpsertGoogleDriveAccountRow{
		AccountID: queries.accountID, UserID: params.UserID,
		GoogleSubject: params.GoogleSubject, Email: params.Email, DisplayName: params.DisplayName,
		CredentialPayload: params.CredentialPayload, CredentialKeyVersion: params.CredentialKeyVersion,
		InstallationGeneration: params.InstallationGeneration, Scopes: params.Scopes,
		ExpiresAt: params.ExpiresAt, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}, nil
}

func (queries *connectionLifecycleQueries) UpsertGoogleDriveWorkspaceConnection(
	_ context.Context,
	_ googledrivesql.UpsertGoogleDriveWorkspaceConnectionParams,
) (int64, error) {
	queries.calls = append(queries.calls, "upsert_connection")
	return queries.upsertConnectionRows, nil
}

func (queries *connectionLifecycleQueries) LockGoogleDriveUserLifecycle(
	_ context.Context,
	params googledrivesql.LockGoogleDriveUserLifecycleParams,
) error {
	queries.calls = append(queries.calls, "lock")
	queries.userID = params.UserID
	return queries.lockErr
}

func (queries *connectionLifecycleQueries) LockGoogleDriveSubjectLifecycle(
	_ context.Context,
	_ googledrivesql.LockGoogleDriveSubjectLifecycleParams,
) error {
	queries.calls = append(queries.calls, "lock_subject")
	return nil
}

func (queries *connectionLifecycleQueries) GetActiveGoogleDriveAccountBySubject(
	_ context.Context,
	_ googledrivesql.GetActiveGoogleDriveAccountBySubjectParams,
) (googledrivesql.GetActiveGoogleDriveAccountBySubjectRow, error) {
	queries.calls = append(queries.calls, "get_subject")
	if queries.activeSubjectAccount != nil {
		return *queries.activeSubjectAccount, nil
	}
	return googledrivesql.GetActiveGoogleDriveAccountBySubjectRow{}, pgx.ErrNoRows
}

func (queries *connectionLifecycleQueries) SupersedeGoogleDriveRevocationsForGeneration(
	_ context.Context,
	params googledrivesql.SupersedeGoogleDriveRevocationsForGenerationParams,
) (int64, error) {
	queries.calls = append(queries.calls, "supersede")
	queries.supersededSubject = params.GoogleSubject
	queries.supersededGeneration = params.ActiveGeneration
	return 0, nil
}

func (queries *connectionLifecycleQueries) DeleteGoogleDriveWorkspaceConnection(
	_ context.Context,
	params googledrivesql.DeleteGoogleDriveWorkspaceConnectionParams,
) (uuid.UUID, error) {
	queries.calls = append(queries.calls, "delete")
	queries.workspaceID = params.WorkspaceID
	queries.userID = params.UserID
	return queries.accountID, queries.deleteErr
}

func (queries *connectionLifecycleQueries) CountGoogleDriveAccountConnections(
	_ context.Context,
	_ googledrivesql.CountGoogleDriveAccountConnectionsParams,
) (int64, error) {
	queries.calls = append(queries.calls, "count")
	return queries.connectionCount, nil
}

func (queries *connectionLifecycleQueries) RevokeUnusedGoogleDriveAccount(
	_ context.Context,
	params googledrivesql.RevokeUnusedGoogleDriveAccountParams,
) (int64, error) {
	queries.calls = append(queries.calls, "revoke")
	queries.revokedAccountID = params.AccountID
	return queries.revokeRows, nil
}

func connectionLifecycleRepository(queries googledrivesql.Querier) *Repository {
	return &Repository{
		queries: queries,
		runTransaction: func(ctx context.Context, operation func(googledrivesql.Querier) error) error {
			return operation(queries)
		},
	}
}

package googledriverepository

import (
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestEnqueueRevocationSupportsFailedOAuthWithoutSourceAccount(t *testing.T) {
	t.Parallel()

	revocation := domain.Revocation{
		UserID: uuid.New(), GoogleSubject: "google-subject",
		InstallationGeneration: uuid.New(), CredentialPayload: "vault.v2.encrypted",
		CredentialVersion: 2,
	}
	queries := &revocationQueries{
		enqueue: googledrivesql.EnqueueGoogleDriveRevocationRow{
			RevocationID: uuid.New(), UserID: revocation.UserID, GoogleSubject: revocation.GoogleSubject,
		},
	}
	repository := revocationTestRepository(queries)

	candidate, err := repository.EnqueueRevocation(t.Context(), revocation)

	require.NoError(t, err)
	require.Equal(t, queries.enqueue.RevocationID, candidate.ID)
	require.Nil(t, queries.enqueueParams.SourceAccountID)
	require.NotNil(t, queries.enqueueParams.CredentialPayload)
	require.Equal(t, revocation.CredentialPayload, *queries.enqueueParams.CredentialPayload)
	require.Equal(t, []string{"lock_user", "lock_subject", "get_subject", "enqueue"}, queries.calls)
}

func TestClaimRevocationSupersedesWithoutClaimWhenSubjectReconnected(t *testing.T) {
	t.Parallel()

	candidate := domain.RevocationCandidate{ID: uuid.New(), UserID: uuid.New(), GoogleSubject: "google-subject"}
	queries := &revocationQueries{candidate: candidate, supersedeRows: 1}
	repository := revocationTestRepository(queries)
	now := time.Now().UTC()

	_, claimed, err := repository.ClaimRevocation(t.Context(), candidate, uuid.New(), now, now.Add(time.Minute))

	require.NoError(t, err)
	require.False(t, claimed)
	require.Equal(t, []string{"identity", "lock_user", "lock_subject", "supersede"}, queries.calls)
}

func TestClaimRevocationFencesAndReturnsEncryptedGeneration(t *testing.T) {
	t.Parallel()

	candidate := domain.RevocationCandidate{ID: uuid.New(), UserID: uuid.New(), GoogleSubject: "google-subject"}
	claimToken := uuid.New()
	generation := uuid.New()
	accountID := uuid.New()
	payload := "vault.v2.encrypted"
	now := time.Now().UTC()
	leaseExpiresAt := now.Add(time.Minute)
	queries := &revocationQueries{
		candidate: candidate,
		claim: googledrivesql.ClaimGoogleDriveRevocationRow{
			RevocationID: candidate.ID, SourceAccountID: &accountID,
			UserID: candidate.UserID, GoogleSubject: candidate.GoogleSubject,
			InstallationGeneration: generation, CredentialPayload: &payload,
			CredentialKeyVersion: 2, AttemptCount: 1, ClaimToken: &claimToken,
			LeaseExpiresAt: &leaseExpiresAt, CreatedAt: now, UpdatedAt: now,
		},
	}
	repository := revocationTestRepository(queries)

	revocation, claimed, err := repository.ClaimRevocation(
		t.Context(), candidate, claimToken, now, leaseExpiresAt,
	)

	require.NoError(t, err)
	require.True(t, claimed)
	require.Equal(t, accountID, *revocation.SourceAccountID)
	require.Equal(t, generation, revocation.InstallationGeneration)
	require.Equal(t, payload, revocation.CredentialPayload)
	require.Equal(t, []string{"identity", "lock_user", "lock_subject", "supersede", "claim"}, queries.calls)
}

type revocationQueries struct {
	googledrivesql.Querier
	candidate     domain.RevocationCandidate
	claim         googledrivesql.ClaimGoogleDriveRevocationRow
	enqueue       googledrivesql.EnqueueGoogleDriveRevocationRow
	enqueueParams googledrivesql.EnqueueGoogleDriveRevocationParams
	supersedeRows int64
	calls         []string
}

func (queries *revocationQueries) GetActiveGoogleDriveAccountBySubject(
	context.Context,
	googledrivesql.GetActiveGoogleDriveAccountBySubjectParams,
) (googledrivesql.GetActiveGoogleDriveAccountBySubjectRow, error) {
	queries.calls = append(queries.calls, "get_subject")
	return googledrivesql.GetActiveGoogleDriveAccountBySubjectRow{}, pgx.ErrNoRows
}

func (queries *revocationQueries) EnqueueGoogleDriveRevocation(
	_ context.Context,
	params googledrivesql.EnqueueGoogleDriveRevocationParams,
) (googledrivesql.EnqueueGoogleDriveRevocationRow, error) {
	queries.calls = append(queries.calls, "enqueue")
	queries.enqueueParams = params
	return queries.enqueue, nil
}

func (queries *revocationQueries) GetGoogleDriveRevocationIdentityForUpdate(
	context.Context,
	googledrivesql.GetGoogleDriveRevocationIdentityForUpdateParams,
) (googledrivesql.GetGoogleDriveRevocationIdentityForUpdateRow, error) {
	queries.calls = append(queries.calls, "identity")
	return googledrivesql.GetGoogleDriveRevocationIdentityForUpdateRow{
		UserID: queries.candidate.UserID, GoogleSubject: queries.candidate.GoogleSubject,
	}, nil
}

func (queries *revocationQueries) LockGoogleDriveUserLifecycle(
	context.Context,
	googledrivesql.LockGoogleDriveUserLifecycleParams,
) error {
	queries.calls = append(queries.calls, "lock_user")
	return nil
}

func (queries *revocationQueries) LockGoogleDriveSubjectLifecycle(
	context.Context,
	googledrivesql.LockGoogleDriveSubjectLifecycleParams,
) error {
	queries.calls = append(queries.calls, "lock_subject")
	return nil
}

func (queries *revocationQueries) SupersedeGoogleDriveRevocationIfSubjectActive(
	context.Context,
	googledrivesql.SupersedeGoogleDriveRevocationIfSubjectActiveParams,
) (int64, error) {
	queries.calls = append(queries.calls, "supersede")
	return queries.supersedeRows, nil
}

func (queries *revocationQueries) ClaimGoogleDriveRevocation(
	context.Context,
	googledrivesql.ClaimGoogleDriveRevocationParams,
) (googledrivesql.ClaimGoogleDriveRevocationRow, error) {
	queries.calls = append(queries.calls, "claim")
	return queries.claim, nil
}

func revocationTestRepository(queries googledrivesql.Querier) *Repository {
	return &Repository{
		queries: queries,
		runTransaction: func(ctx context.Context, operation func(googledrivesql.Querier) error) error {
			return operation(queries)
		},
	}
}

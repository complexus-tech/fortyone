package usersrepository

import (
	"context"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/stretchr/testify/require"
)

func TestUserMaintenanceQueriesUseExplicitCutoffsAndBoundedLocks(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/maintenance.sql")
	if err != nil {
		t.Fatalf("read user maintenance queries: %v", err)
	}
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")

	for _, contract := range []string{
		"-- name: purgeexpiredverificationtokens :execrows",
		"verification_token.expires_at < sqlc.arg(retained_before)",
		"order by verification_token.expires_at, verification_token.id",
		"for update of verification_token skip locked",
		"delete from public.verification_tokens as verification_token using candidates",
		"-- name: deactivateinactiveusers :execrows",
		"account.last_login_at < sqlc.arg(inactive_before)",
		"account.inactivity_warning_sent_at is not null",
		"cast(sqlc.arg(warning_sent_before) as timestamptz) at time zone 'utc'",
		"account.is_active = true",
		"account.is_system = false",
		"order by account.inactivity_warning_sent_at, account.last_login_at, account.user_id",
		"for update of account skip locked",
		"is_active = false",
		"login_reactivation_policy = 'verified_sign_in'",
		"auth_session_version = account.auth_session_version + 1",
		"updated_at = sqlc.arg(deactivated_at)",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("user maintenance queries are missing contract %q", contract)
		}
	}

	if count := strings.Count(query, "limit cast(sqlc.arg(batch_size) as integer)"); count != 2 {
		t.Fatalf("user maintenance batch limit count = %d, want 2", count)
	}
	if strings.Contains(query, "verification_token.created_at <") {
		t.Fatal("verification-token retention must be based on expiry, not issuance")
	}
	if strings.Contains(query, "current_timestamp") || strings.Contains(query, "now()") || strings.Contains(query, "interval '") {
		t.Fatal("user maintenance cutoffs and mutation time must be supplied by the application clock")
	}
}

func TestUserMaintenanceRepositoryMapsUTCAndBatchesToSQLC(t *testing.T) {
	t.Parallel()

	queries := &userMaintenanceQueries{tokenRows: 3, deactivationRows: 2}
	repository := newWithQueries(queries)
	location := time.FixedZone("CAT", 2*60*60)
	retainedBefore := time.Date(2026, time.August, 21, 10, 15, 0, 0, location)
	inactiveBefore := time.Date(2025, time.December, 28, 10, 15, 0, 0, location)
	warningSentBefore := time.Date(2026, time.July, 29, 10, 15, 0, 0, location)
	deactivatedAt := time.Date(2026, time.August, 28, 10, 15, 0, 0, location)

	tokensDeleted, err := repository.PurgeExpiredVerificationTokens(
		context.Background(),
		retainedBefore,
		500,
	)
	require.NoError(t, err)
	require.Equal(t, int64(3), tokensDeleted)
	require.Equal(t, usersql.PurgeExpiredVerificationTokensParams{
		RetainedBefore: retainedBefore.UTC(),
		BatchSize:      500,
	}, queries.tokenParams)

	usersDeactivated, err := repository.DeactivateInactiveUsers(
		context.Background(),
		inactiveBefore,
		warningSentBefore,
		deactivatedAt,
		500,
	)
	require.NoError(t, err)
	require.Equal(t, int64(2), usersDeactivated)
	require.NotNil(t, queries.deactivationParams.InactiveBefore)
	require.Equal(t, inactiveBefore.UTC(), *queries.deactivationParams.InactiveBefore)
	require.Equal(t, warningSentBefore.UTC(), queries.deactivationParams.WarningSentBefore)
	require.Equal(t, deactivatedAt.UTC(), queries.deactivationParams.DeactivatedAt)
	require.Equal(t, int32(500), queries.deactivationParams.BatchSize)
}

func TestUserMaintenanceRepositoryRejectsBatchesOutsideSQLCRange(t *testing.T) {
	t.Parallel()

	queries := &userMaintenanceQueries{}
	repository := newWithQueries(queries)
	cutoff := time.Date(2026, time.July, 29, 8, 15, 0, 0, time.UTC)
	now := time.Date(2026, time.August, 28, 8, 15, 0, 0, time.UTC)
	batchSize := int(math.MaxInt32) + 1

	_, tokenErr := repository.PurgeExpiredVerificationTokens(context.Background(), cutoff, batchSize)
	_, deactivationErr := repository.DeactivateInactiveUsers(
		context.Background(),
		cutoff.AddDate(0, -7, 0),
		cutoff,
		now,
		batchSize,
	)

	require.ErrorIs(t, tokenErr, safecast.ErrOutOfRange)
	require.ErrorIs(t, deactivationErr, safecast.ErrOutOfRange)
	require.Zero(t, queries.tokenCalls)
	require.Zero(t, queries.deactivationCalls)
}

type userMaintenanceQueries struct {
	usersql.Querier
	tokenParams        usersql.PurgeExpiredVerificationTokensParams
	deactivationParams usersql.DeactivateInactiveUsersParams
	tokenRows          int64
	deactivationRows   int64
	tokenErr           error
	deactivationErr    error
	tokenCalls         int
	deactivationCalls  int
}

func (queries *userMaintenanceQueries) PurgeExpiredVerificationTokens(
	_ context.Context,
	params usersql.PurgeExpiredVerificationTokensParams,
) (int64, error) {
	queries.tokenCalls++
	queries.tokenParams = params
	return queries.tokenRows, queries.tokenErr
}

func (queries *userMaintenanceQueries) DeactivateInactiveUsers(
	_ context.Context,
	params usersql.DeactivateInactiveUsersParams,
) (int64, error) {
	queries.deactivationCalls++
	queries.deactivationParams = params
	return queries.deactivationRows, queries.deactivationErr
}

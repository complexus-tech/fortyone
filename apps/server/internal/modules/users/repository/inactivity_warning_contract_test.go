package usersrepository

import (
	"context"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestUserInactivityWarningQueriesUseExplicitTimeStableKeysetsAndEligibilityGuards(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/inactivity_warning.sql")
	require.NoError(t, err)
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")

	for _, contract := range []string{
		"-- name: listuserinactivitywarningcandidates :many",
		"account.last_login_at < sqlc.arg(inactive_before)",
		"account.is_active = true",
		"account.is_system = false",
		"cast(btrim(account.email) as text) as email",
		"nullif(btrim(account.email), '') is not null",
		"account.inactivity_warning_sent_at is null",
		"not cast(sqlc.arg(has_cursor) as boolean)",
		"account.last_login_at > sqlc.arg(after_last_login_at)",
		"account.user_id > sqlc.arg(after_user_id)",
		"order by account.last_login_at, account.user_id",
		"limit cast(sqlc.arg(batch_size) as integer)",
		"-- name: geteligibleuserinactivitywarningcandidate :one",
		"-- name: markuserinactivitywarningsent :execrows",
		"cast(sqlc.arg(warning_sent_at) as timestamptz) at time zone 'utc'",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("user inactivity warning queries are missing contract %q", contract)
		}
	}
	if count := strings.Count(query, "account.is_active = true"); count != 3 {
		t.Fatalf("active eligibility guard count = %d, want 3", count)
	}
	if count := strings.Count(query, "account.inactivity_warning_sent_at is null"); count != 3 {
		t.Fatalf("unwarned eligibility guard count = %d, want 3", count)
	}
	if count := strings.Count(query, "nullif(btrim(account.email), '') is not null"); count != 2 {
		t.Fatalf("non-empty recipient guard count = %d, want 2", count)
	}
	for _, forbidden := range []string{
		"now()",
		"current_timestamp",
		"current_date",
		"interval '",
		" offset ",
		"select *",
		"notification_preferences",
	} {
		if strings.Contains(query, forbidden) {
			t.Errorf("user inactivity warning queries contain forbidden contract %q", forbidden)
		}
	}
}

func TestUserInactivityWarningRepositoryMapsTypedUTCParameters(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	inactiveBefore := time.Date(2025, time.December, 28, 10, 15, 0, 0, location)
	lastLoginAt := time.Date(2025, time.November, 1, 9, 30, 0, 0, location)
	userID := uuid.New()
	queries := &userInactivityWarningQueries{
		listRows: []usersql.ListUserInactivityWarningCandidatesRow{{
			UserID:      userID,
			Email:       "user@example.com",
			FullName:    "User Name",
			LastLoginAt: &lastLoginAt,
		}},
		eligibleRow: usersql.GetEligibleUserInactivityWarningCandidateRow{
			UserID:      userID,
			Email:       "current@example.com",
			FullName:    "Current Name",
			LastLoginAt: &lastLoginAt,
		},
		markRows: 1,
	}
	repository := newWithQueries(queries)
	cursor := usersdomain.InactivityWarningCursor{
		LastLoginAt: lastLoginAt,
		UserID:      userID,
		Valid:       true,
	}

	candidates, err := repository.ListUserInactivityWarningCandidates(
		context.Background(),
		usersdomain.InactivityWarningQuery{
			InactiveBefore: inactiveBefore,
			Cursor:         cursor,
			BatchSize:      100,
		},
	)
	require.NoError(t, err)
	require.Equal(t, usersql.ListUserInactivityWarningCandidatesParams{
		InactiveBefore:   timePointer(inactiveBefore.UTC()),
		HasCursor:        true,
		AfterLastLoginAt: timePointer(lastLoginAt.UTC()),
		AfterUserID:      userID,
		BatchSize:        100,
	}, queries.listParams)
	require.Equal(t, []usersdomain.InactivityWarningCandidate{{
		UserID:      userID,
		Email:       "user@example.com",
		FullName:    "User Name",
		LastLoginAt: lastLoginAt.UTC(),
	}}, candidates)

	eligible, found, err := repository.GetEligibleUserInactivityWarningCandidate(
		context.Background(),
		usersdomain.InactivityWarningEligibility{
			UserID:         userID,
			InactiveBefore: inactiveBefore,
		},
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "current@example.com", eligible.Email)
	require.Equal(t, usersql.GetEligibleUserInactivityWarningCandidateParams{
		UserID:         userID,
		InactiveBefore: timePointer(inactiveBefore.UTC()),
	}, queries.eligibleParams)

	warningSentAt := time.Date(2026, time.August, 28, 10, 15, 0, 0, location)
	recorded, err := repository.RecordUserInactivityWarning(
		context.Background(),
		usersdomain.InactivityWarningReceipt{
			UserID:         userID,
			InactiveBefore: inactiveBefore,
			WarningSentAt:  warningSentAt,
		},
	)
	require.NoError(t, err)
	require.True(t, recorded)
	require.Equal(t, usersql.MarkUserInactivityWarningSentParams{
		WarningSentAt:  warningSentAt.UTC(),
		UserID:         userID,
		InactiveBefore: timePointer(inactiveBefore.UTC()),
	}, queries.markParams)
}

func TestUserInactivityWarningRepositoryMapsChangedEligibilityToNotFound(t *testing.T) {
	t.Parallel()

	queries := &userInactivityWarningQueries{eligibleErr: pgx.ErrNoRows}
	repository := newWithQueries(queries)

	candidate, found, err := repository.GetEligibleUserInactivityWarningCandidate(
		context.Background(),
		usersdomain.InactivityWarningEligibility{
			UserID:         uuid.New(),
			InactiveBefore: time.Date(2025, time.December, 28, 8, 0, 0, 0, time.UTC),
		},
	)

	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, candidate)
}

func TestUserInactivityWarningRepositoryRejectsInvalidBatchesAndRows(t *testing.T) {
	t.Parallel()

	queries := &userInactivityWarningQueries{
		listRows: []usersql.ListUserInactivityWarningCandidatesRow{{
			UserID: uuid.New(), Email: "user@example.com", FullName: "User",
		}},
	}
	repository := newWithQueries(queries)
	cutoff := time.Date(2025, time.December, 28, 8, 0, 0, 0, time.UTC)

	_, rangeErr := repository.ListUserInactivityWarningCandidates(
		context.Background(),
		usersdomain.InactivityWarningQuery{
			InactiveBefore: cutoff,
			BatchSize:      int(math.MaxInt32) + 1,
		},
	)
	require.ErrorIs(t, rangeErr, safecast.ErrOutOfRange)
	require.Zero(t, queries.listCalls)

	_, mappingErr := repository.ListUserInactivityWarningCandidates(
		context.Background(),
		usersdomain.InactivityWarningQuery{InactiveBefore: cutoff, BatchSize: 1},
	)
	require.ErrorContains(t, mappingErr, "last login time")
}

type userInactivityWarningQueries struct {
	usersql.Querier
	listParams     usersql.ListUserInactivityWarningCandidatesParams
	listRows       []usersql.ListUserInactivityWarningCandidatesRow
	listErr        error
	listCalls      int
	eligibleParams usersql.GetEligibleUserInactivityWarningCandidateParams
	eligibleRow    usersql.GetEligibleUserInactivityWarningCandidateRow
	eligibleErr    error
	markParams     usersql.MarkUserInactivityWarningSentParams
	markRows       int64
	markErr        error
}

func (queries *userInactivityWarningQueries) ListUserInactivityWarningCandidates(
	_ context.Context,
	params usersql.ListUserInactivityWarningCandidatesParams,
) ([]usersql.ListUserInactivityWarningCandidatesRow, error) {
	queries.listCalls++
	queries.listParams = params
	return queries.listRows, queries.listErr
}

func (queries *userInactivityWarningQueries) GetEligibleUserInactivityWarningCandidate(
	_ context.Context,
	params usersql.GetEligibleUserInactivityWarningCandidateParams,
) (usersql.GetEligibleUserInactivityWarningCandidateRow, error) {
	queries.eligibleParams = params
	return queries.eligibleRow, queries.eligibleErr
}

func (queries *userInactivityWarningQueries) MarkUserInactivityWarningSent(
	_ context.Context,
	params usersql.MarkUserInactivityWarningSentParams,
) (int64, error) {
	queries.markParams = params
	return queries.markRows, queries.markErr
}

func timePointer(value time.Time) *time.Time {
	return &value
}

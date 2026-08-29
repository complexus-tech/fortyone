package jobs

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserInactivityWarningUsesOneUTCClockAndFreshEligibilityData(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))
	userID := uuid.New()
	lastLoginAt := time.Date(2025, time.December, 12, 8, 0, 0, 0, time.UTC)
	listed := usersdomain.InactivityWarningCandidate{
		UserID: userID, Email: "stale@example.com", FullName: "Stale Name", LastLoginAt: lastLoginAt,
	}
	refreshed := usersdomain.InactivityWarningCandidate{
		UserID: userID, Email: "current@example.com", FullName: "Current Name", LastLoginAt: lastLoginAt,
	}
	store := &userInactivityWarningStoreStub{
		pages:    [][]usersdomain.InactivityWarningCandidate{{listed}},
		eligible: map[uuid.UUID]usersdomain.InactivityWarningCandidate{userID: refreshed},
	}
	mailerService := &userInactivityWarningMailerStub{}

	err := processUserInactivityWarningAt(
		context.Background(),
		store,
		newUserInactivityWarningTestLogger(),
		mailerService,
		now,
	)

	require.NoError(t, err)
	inactiveBefore := now.UTC().AddDate(0, -8, 0)
	require.Equal(t, []usersdomain.InactivityWarningQuery{{
		InactiveBefore: inactiveBefore,
		BatchSize:      userInactivityWarningBatchSize,
	}}, store.queries)
	require.Equal(t, []usersdomain.InactivityWarningEligibility{{
		UserID:         userID,
		InactiveBefore: inactiveBefore,
	}}, store.eligibilityChecks)
	require.Equal(t, []usersdomain.InactivityWarningReceipt{{
		UserID:         userID,
		InactiveBefore: inactiveBefore,
		WarningSentAt:  now.UTC(),
	}}, store.receipts)
	require.Len(t, mailerService.templated, 1)
	require.Equal(t, mailer.TemplatedEmail{
		To:       []string{"current@example.com"},
		Template: "users/inactivity_warning",
		Subject:  "Your account is scheduled for deactivation",
		Data: map[string]any{
			"UserName": "Current Name",
			"LoginURL": "https://fortyone.app/login",
		},
	}, mailerService.templated[0])
}

func TestUserInactivityWarningUsesStableCursorAcrossFullPages(t *testing.T) {
	firstPage := make([]usersdomain.InactivityWarningCandidate, userInactivityWarningBatchSize)
	eligible := make(map[uuid.UUID]usersdomain.InactivityWarningCandidate, userInactivityWarningBatchSize+1)
	baseTime := time.Date(2025, time.October, 1, 0, 0, 0, 0, time.UTC)
	for index := range firstPage {
		candidate := usersdomain.InactivityWarningCandidate{
			UserID:      uuid.New(),
			Email:       "user@example.com",
			FullName:    "User",
			LastLoginAt: baseTime.Add(time.Duration(index) * time.Minute),
		}
		firstPage[index] = candidate
		eligible[candidate.UserID] = candidate
	}
	lastCandidate := firstPage[len(firstPage)-1]
	secondCandidate := usersdomain.InactivityWarningCandidate{
		UserID:      uuid.New(),
		Email:       "last@example.com",
		FullName:    "Last User",
		LastLoginAt: lastCandidate.LastLoginAt.Add(time.Minute),
	}
	eligible[secondCandidate.UserID] = secondCandidate
	store := &userInactivityWarningStoreStub{
		pages: [][]usersdomain.InactivityWarningCandidate{
			firstPage,
			{secondCandidate},
		},
		eligible: eligible,
	}

	err := processUserInactivityWarningAt(
		context.Background(),
		store,
		newUserInactivityWarningTestLogger(),
		&userInactivityWarningMailerStub{},
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Len(t, store.queries, 2)
	require.Equal(t, usersdomain.InactivityWarningCursor{
		LastLoginAt: lastCandidate.LastLoginAt,
		UserID:      lastCandidate.UserID,
		Valid:       true,
	}, store.queries[1].Cursor)
	require.Len(t, store.receipts, userInactivityWarningBatchSize+1)
}

func TestUserInactivityWarningSkipsCandidateThatIsNoLongerEligible(t *testing.T) {
	t.Parallel()

	candidate := usersdomain.InactivityWarningCandidate{
		UserID:      uuid.New(),
		Email:       "active-again@example.com",
		FullName:    "Active Again",
		LastLoginAt: time.Date(2025, time.October, 1, 0, 0, 0, 0, time.UTC),
	}
	store := &userInactivityWarningStoreStub{
		pages:      [][]usersdomain.InactivityWarningCandidate{{candidate}},
		ineligible: map[uuid.UUID]bool{candidate.UserID: true},
	}
	mailerService := &userInactivityWarningMailerStub{}

	err := processUserInactivityWarningAt(
		context.Background(),
		store,
		newUserInactivityWarningTestLogger(),
		mailerService,
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Empty(t, mailerService.templated)
	require.Empty(t, store.receipts)
}

func TestUserInactivityWarningDoesNotRecordFailedDelivery(t *testing.T) {
	t.Parallel()

	candidate := usersdomain.InactivityWarningCandidate{
		UserID:      uuid.New(),
		Email:       "inactive@example.com",
		FullName:    "Inactive User",
		LastLoginAt: time.Date(2025, time.October, 1, 0, 0, 0, 0, time.UTC),
	}
	store := &userInactivityWarningStoreStub{
		pages:    [][]usersdomain.InactivityWarningCandidate{{candidate}},
		eligible: map[uuid.UUID]usersdomain.InactivityWarningCandidate{candidate.UserID: candidate},
	}

	err := processUserInactivityWarningAt(
		context.Background(),
		store,
		newUserInactivityWarningTestLogger(),
		&userInactivityWarningMailerStub{templatedErr: errors.New("mailer unavailable")},
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Empty(t, store.receipts)
}

func TestUserInactivityWarningRejectsOversizedRepositoryPage(t *testing.T) {
	t.Parallel()

	store := &userInactivityWarningStoreStub{
		pages: [][]usersdomain.InactivityWarningCandidate{
			make([]usersdomain.InactivityWarningCandidate, userInactivityWarningBatchSize+1),
		},
	}

	err := processUserInactivityWarningAt(
		context.Background(),
		store,
		newUserInactivityWarningTestLogger(),
		&userInactivityWarningMailerStub{},
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.ErrorContains(t, err, "want at most 100")
	require.Empty(t, store.eligibilityChecks)
}

type userInactivityWarningStoreStub struct {
	pages             [][]usersdomain.InactivityWarningCandidate
	queries           []usersdomain.InactivityWarningQuery
	eligibilityChecks []usersdomain.InactivityWarningEligibility
	eligible          map[uuid.UUID]usersdomain.InactivityWarningCandidate
	ineligible        map[uuid.UUID]bool
	receipts          []usersdomain.InactivityWarningReceipt
	listErr           error
	eligibilityErr    error
	recordErr         error
	recordResult      *bool
	currentPage       int
}

func (store *userInactivityWarningStoreStub) ListUserInactivityWarningCandidates(
	_ context.Context,
	query usersdomain.InactivityWarningQuery,
) ([]usersdomain.InactivityWarningCandidate, error) {
	store.queries = append(store.queries, query)
	if store.listErr != nil {
		return nil, store.listErr
	}
	if store.currentPage >= len(store.pages) {
		return nil, nil
	}
	page := store.pages[store.currentPage]
	store.currentPage++
	return page, nil
}

func (store *userInactivityWarningStoreStub) GetEligibleUserInactivityWarningCandidate(
	_ context.Context,
	eligibility usersdomain.InactivityWarningEligibility,
) (usersdomain.InactivityWarningCandidate, bool, error) {
	store.eligibilityChecks = append(store.eligibilityChecks, eligibility)
	if store.eligibilityErr != nil {
		return usersdomain.InactivityWarningCandidate{}, false, store.eligibilityErr
	}
	if store.ineligible[eligibility.UserID] {
		return usersdomain.InactivityWarningCandidate{}, false, nil
	}
	candidate, exists := store.eligible[eligibility.UserID]
	return candidate, exists, nil
}

func (store *userInactivityWarningStoreStub) RecordUserInactivityWarning(
	_ context.Context,
	receipt usersdomain.InactivityWarningReceipt,
) (bool, error) {
	store.receipts = append(store.receipts, receipt)
	if store.recordErr != nil {
		return false, store.recordErr
	}
	if store.recordResult != nil {
		return *store.recordResult, nil
	}
	return true, nil
}

type userInactivityWarningMailerStub struct {
	templated    []mailer.TemplatedEmail
	templatedErr error
}

func (*userInactivityWarningMailerStub) Send(context.Context, mailer.Email) error {
	return nil
}

func (service *userInactivityWarningMailerStub) SendTemplated(
	_ context.Context,
	email mailer.TemplatedEmail,
) error {
	service.templated = append(service.templated, email)
	return service.templatedErr
}

func newUserInactivityWarningTestLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelDebug, "user-inactivity-warning-test")
}

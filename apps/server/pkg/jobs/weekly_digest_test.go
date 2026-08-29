package jobs

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeWeeklyDigestStore struct {
	mu sync.Mutex

	pages       [][]notificationsdomain.WeeklyDigestRecipient
	page        int
	cursors     []*notificationsdomain.WeeklyDigestCursor
	listLimits  []int
	listErr     error
	stats       map[string]notificationsdomain.WeeklyDigestStats
	statsCalls  map[string]int
	statsAsOf   []time.Time
	statsErrors map[string][]error
}

func (store *fakeWeeklyDigestStore) ListWeeklyDigestRecipients(
	_ context.Context,
	cursor *notificationsdomain.WeeklyDigestCursor,
	limit int,
) ([]notificationsdomain.WeeklyDigestRecipient, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.listLimits = append(store.listLimits, limit)
	if cursor == nil {
		store.cursors = append(store.cursors, nil)
	} else {
		copied := *cursor
		store.cursors = append(store.cursors, &copied)
	}
	if store.listErr != nil {
		return nil, store.listErr
	}
	if store.page >= len(store.pages) {
		return []notificationsdomain.WeeklyDigestRecipient{}, nil
	}
	page := append([]notificationsdomain.WeeklyDigestRecipient(nil), store.pages[store.page]...)
	store.page++
	return page, nil
}

func (store *fakeWeeklyDigestStore) GetWeeklyDigestStats(
	_ context.Context,
	query notificationsdomain.WeeklyDigestStatsQuery,
) (notificationsdomain.WeeklyDigestStats, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	key := weeklyDigestRecipientKey(query.UserID, query.WorkspaceID)
	store.statsCalls[key]++
	store.statsAsOf = append(store.statsAsOf, query.AsOf)
	if failures := store.statsErrors[key]; len(failures) > 0 {
		err := failures[0]
		store.statsErrors[key] = failures[1:]
		if err != nil {
			return notificationsdomain.WeeklyDigestStats{}, err
		}
	}
	return store.stats[key], nil
}

type recordingWeeklyDigestMailer struct {
	mu     sync.Mutex
	emails []mailer.TemplatedEmail
	err    error
}

func (service *recordingWeeklyDigestMailer) Send(context.Context, mailer.Email) error {
	return service.err
}

func (service *recordingWeeklyDigestMailer) SendTemplated(_ context.Context, email mailer.TemplatedEmail) error {
	service.mu.Lock()
	defer service.mu.Unlock()
	service.emails = append(service.emails, email)
	return service.err
}

func (service *recordingWeeklyDigestMailer) count() int {
	service.mu.Lock()
	defer service.mu.Unlock()
	return len(service.emails)
}

func TestProcessWeeklyDigestUsesCompositeCursorAndOneUTCAsOf(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	firstPage := make([]notificationsdomain.WeeklyDigestRecipient, weeklyDigestBatchSize)
	for index := range firstPage {
		firstPage[index] = weeklyDigestRecipient(uuid.New(), workspaceID)
	}
	lastRecipient := weeklyDigestRecipient(uuid.New(), workspaceID)
	store := &fakeWeeklyDigestStore{
		pages:      [][]notificationsdomain.WeeklyDigestRecipient{firstPage, {lastRecipient}},
		stats:      make(map[string]notificationsdomain.WeeklyDigestStats),
		statsCalls: make(map[string]int),
	}
	store.stats[weeklyDigestRecipientKey(firstPage[0].UserID, workspaceID)] = notificationsdomain.WeeklyDigestStats{OverdueStories: 1}
	store.stats[weeklyDigestRecipientKey(lastRecipient.UserID, workspaceID)] = notificationsdomain.WeeklyDigestStats{ObjectiveRisks: 1}
	mailerService := &recordingWeeklyDigestMailer{}
	asOfInput := time.Date(2026, time.August, 31, 10, 15, 0, 0, time.FixedZone("test", 2*60*60))
	asOfUTC := asOfInput.UTC()

	err := processWeeklyDigestEmailAt(context.Background(), store, newWeeklyDigestTestLogger(), mailerService, nil, nil, asOfInput)
	require.NoError(t, err)
	require.Equal(t, 2, mailerService.count())
	require.Equal(t, []int{weeklyDigestBatchSize, weeklyDigestBatchSize}, store.listLimits)
	require.Len(t, store.cursors, 2)
	require.Nil(t, store.cursors[0])
	require.Equal(t, firstPage[len(firstPage)-1].WorkspaceID, store.cursors[1].WorkspaceID)
	require.Equal(t, firstPage[len(firstPage)-1].UserID, store.cursors[1].UserID)
	require.Len(t, store.statsAsOf, weeklyDigestBatchSize+1)
	for _, actual := range store.statsAsOf {
		require.Equal(t, asOfUTC, actual)
	}
	for key, calls := range store.statsCalls {
		require.Equalf(t, 1, calls, "recipient %s should be loaded once", key)
	}
}

func TestProcessWeeklyDigestRetriesOnlyStatsReads(t *testing.T) {
	t.Parallel()

	recipient := weeklyDigestRecipient(uuid.New(), uuid.New())
	key := weeklyDigestRecipientKey(recipient.UserID, recipient.WorkspaceID)
	store := &fakeWeeklyDigestStore{
		pages:       [][]notificationsdomain.WeeklyDigestRecipient{{recipient}},
		stats:       map[string]notificationsdomain.WeeklyDigestStats{key: {OverdueStories: 1}},
		statsCalls:  make(map[string]int),
		statsErrors: map[string][]error{key: {errors.New("temporary stats read failure")}},
	}
	mailerService := &recordingWeeklyDigestMailer{}

	err := processWeeklyDigestEmailAt(
		context.Background(), store, newWeeklyDigestTestLogger(), mailerService, nil, nil,
		time.Date(2026, time.August, 31, 8, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)
	require.Equal(t, 2, store.statsCalls[key])
	require.Equal(t, 1, mailerService.count())

	failingMailer := &recordingWeeklyDigestMailer{err: errors.New("mail delivery uncertain")}
	store = &fakeWeeklyDigestStore{
		pages:      [][]notificationsdomain.WeeklyDigestRecipient{{recipient}},
		stats:      map[string]notificationsdomain.WeeklyDigestStats{key: {OverdueStories: 1}},
		statsCalls: make(map[string]int),
	}
	err = processWeeklyDigestEmailAt(
		context.Background(), store, newWeeklyDigestTestLogger(), failingMailer, nil, nil,
		time.Date(2026, time.August, 31, 8, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err, "recipient delivery failures remain isolated from already successful recipients")
	require.Equal(t, 1, store.statsCalls[key])
	require.Equal(t, 1, failingMailer.count(), "uncertain mail delivery must not be retried in-process")
}

func TestProcessWeeklyDigestValidatesDependenciesAndPropagatesPageFailure(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 31, 8, 0, 0, 0, time.UTC)
	mailerService := &recordingWeeklyDigestMailer{}
	log := newWeeklyDigestTestLogger()

	err := processWeeklyDigestEmailAt(context.Background(), nil, log, mailerService, nil, nil, asOf)
	require.EqualError(t, err, "weekly digest store is required")
	err = processWeeklyDigestEmailAt(context.Background(), &fakeWeeklyDigestStore{}, nil, mailerService, nil, nil, asOf)
	require.EqualError(t, err, "weekly digest logger is required")
	err = processWeeklyDigestEmailAt(context.Background(), &fakeWeeklyDigestStore{}, log, nil, nil, nil, asOf)
	require.EqualError(t, err, "weekly digest mailer is required")
	err = processWeeklyDigestEmailAt(context.Background(), &fakeWeeklyDigestStore{}, log, mailerService, nil, nil, time.Time{})
	require.EqualError(t, err, "weekly digest as-of time is required")

	wantErr := errors.New("recipient page unavailable")
	err = processWeeklyDigestEmailAt(
		context.Background(), &fakeWeeklyDigestStore{listErr: wantErr}, log, mailerService, nil, nil, asOf,
	)
	require.ErrorIs(t, err, wantErr)
}

func TestWaitForNextWeeklyDigestBatchHonorsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	startedAt := time.Now()
	err := waitForNextWeeklyDigestBatch(ctx)
	require.ErrorIs(t, err, context.Canceled)
	require.Less(t, time.Since(startedAt), weeklyDigestBatchDelay)
}

func weeklyDigestRecipient(userID, workspaceID uuid.UUID) notificationsdomain.WeeklyDigestRecipient {
	return notificationsdomain.WeeklyDigestRecipient{
		UserID: userID, UserEmail: userID.String() + "@example.com", UserName: "Recipient",
		WorkspaceID: workspaceID, WorkspaceName: "Product", WorkspaceSlug: "product",
	}
}

func weeklyDigestRecipientKey(userID, workspaceID uuid.UUID) string {
	return userID.String() + ":" + workspaceID.String()
}

func newWeeklyDigestTestLogger() *logger.Logger {
	return logger.NewWithJSON(io.Discard, slog.LevelDebug, "weekly-digest-test")
}

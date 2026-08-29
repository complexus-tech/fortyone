package calendarrepository

import (
	"context"
	"os"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestUserScopedCalendarReadQueriesExecute(t *testing.T) {
	databaseURL := os.Getenv("CALENDAR_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CALENDAR_TEST_DATABASE_URL is not configured")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	require.NoError(t, pool.Ping(ctx))

	repo := New(pool)
	workspaceID := uuid.New()
	userID := uuid.New()
	startAt := time.Now().UTC().Add(-time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	events, err := repo.ListCalendarEvents(ctx, workspaceID, userID, startAt, endAt)
	require.NoError(t, err)
	require.Empty(t, events)

	_, err = repo.GetCalendarEvent(ctx, workspaceID, userID, uuid.New())
	require.ErrorIs(t, err, calendar.ErrCalendarEventNotFound)

	busyWindows, err := repo.ListBusyWindows(ctx, workspaceID, userID, startAt, endAt)
	require.NoError(t, err)
	require.Empty(t, busyWindows)
}

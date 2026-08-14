package calendarrepository

import (
	"context"
	"os"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestUserScopedCalendarReadQueriesExecute(t *testing.T) {
	databaseURL := os.Getenv("CALENDAR_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CALENDAR_TEST_DATABASE_URL is not configured")
	}

	db, err := sqlx.Connect("pgx", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	repo := New(nil, db)
	workspaceID := uuid.New()
	userID := uuid.New()
	startAt := time.Now().UTC().Add(-time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	events, err := repo.ListCalendarEvents(context.Background(), workspaceID, userID, startAt, endAt)
	require.NoError(t, err)
	require.Empty(t, events)

	_, err = repo.GetCalendarEvent(context.Background(), workspaceID, userID, uuid.New())
	require.ErrorIs(t, err, calendar.ErrCalendarEventNotFound)

	busyWindows, err := repo.ListBusyWindows(context.Background(), workspaceID, userID, startAt, endAt)
	require.NoError(t, err)
	require.Empty(t, busyWindows)
}

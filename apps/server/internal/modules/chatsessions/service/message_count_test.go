package chatsessions

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type messageCountRepositoryStub struct {
	Repository
	count       int
	err         error
	userID      uuid.UUID
	workspaceID uuid.UUID
	start       time.Time
	end         time.Time
}

func (stub *messageCountRepositoryStub) CountUserMessages(
	_ context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	start time.Time,
	end time.Time,
) (int, error) {
	stub.userID = userID
	stub.workspaceID = workspaceID
	stub.start = start
	stub.end = end
	return stub.count, stub.err
}

func TestCountUserMessagesCurrentMonthUsesClockLocationAndCalendarBounds(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("workspace-test", 2*60*60)
	now := time.Date(2026, time.August, 31, 23, 30, 0, 0, location)
	userID, workspaceID := uuid.New(), uuid.New()
	repository := &messageCountRepositoryStub{count: 42}
	service := NewWithClock(
		testLogger(),
		repository,
		testkit.NewFixedClock(now),
	)

	count, err := service.CountUserMessagesCurrentMonth(t.Context(), userID, workspaceID)
	if err != nil {
		t.Fatalf("count current-month messages: %v", err)
	}
	if count != 42 {
		t.Fatalf("count = %d, want 42", count)
	}
	if repository.userID != userID || repository.workspaceID != workspaceID {
		t.Fatalf("scope = %s/%s, want %s/%s", repository.userID, repository.workspaceID, userID, workspaceID)
	}
	wantStart := time.Date(2026, time.August, 1, 0, 0, 0, 0, location)
	wantEnd := time.Date(2026, time.September, 1, 0, 0, 0, 0, location)
	if !repository.start.Equal(wantStart) || !repository.end.Equal(wantEnd) {
		t.Fatalf("range = %v through %v, want %v through %v", repository.start, repository.end, wantStart, wantEnd)
	}
}

func TestCountUserMessagesCurrentMonthWrapsRepositoryErrors(t *testing.T) {
	t.Parallel()

	repositoryError := errors.New("repository unavailable")
	repository := &messageCountRepositoryStub{err: repositoryError}
	service := NewWithClock(
		testLogger(),
		repository,
		testkit.NewFixedClock(time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)),
	)

	_, err := service.CountUserMessagesCurrentMonth(t.Context(), uuid.New(), uuid.New())
	if !errors.Is(err, repositoryError) {
		t.Fatalf("error = %v, want wrapped repository error", err)
	}
}

func testLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelError, "chatsessions-test")
}
